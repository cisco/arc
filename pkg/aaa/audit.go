//
// Copyright (c) 2017, Cisco Systems
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without modification,
// are permitted provided that the following conditions are met:
//
// * Redistributions of source code must retain the above copyright notice, this
//   list of conditions and the following disclaimer.
//
// * Redistributions in binary form must reproduce the above copyright notice, this
//   list of conditions and the following disclaimer in the documentation and/or
//   other materials provided with the distribution.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
// ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
// WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
// DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR
// ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
// (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
// LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON
// ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
//

package aaa

import (
	"fmt"
	"os"
	"strings"

	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/spark"
)

type Audit struct {
	name             string
	printDeployed    bool
	printConfigured  bool
	printMismatched  bool
	deployedBuffer   []string
	configuredBuffer []string
	mismatchedBuffer []string
	message          []string
}
type auditType uint

const (
	Deployed auditType = iota
	Configured
	Mismatched
)

var freeFormAuditBuffer []string
var AuditBuffer map[string]*Audit

func init() {
	AuditBuffer = make(map[string]*Audit)
}

func NewAuditWithOptions(name string, deployed, configured, mismatched bool) error {
	if name == "" {
		return fmt.Errorf("No name given for the Audit")
	}
	a := &Audit{
		name:            name,
		printDeployed:   deployed,
		printConfigured: configured,
		printMismatched: mismatched,
	}
	a.PreAudit(os.Args)
	AuditBuffer[name] = a
	return nil
}

func NewAudit(name string) error {
	return NewAuditWithOptions(name, true, true, true)
}

func (a *Audit) PreAudit(args []string) {
	version := env.Lookup("VERSION")
	sshUser := env.Lookup("SSH_USER")
	userId := env.Lookup("USER")

	trimmedVersion := strings.Split(version, " ")
	command := strings.Join(args, " ")
	username := userId
	if sshUser != "" && sshUser != userId {
		username += "(" + sshUser + ")"
	}
	s := fmt.Sprintf("**%s Audit Start**: %s | %s | %s", a.name, username, trimmedVersion[0], command)
	a.message = append(a.message, s)
	s = fmt.Sprintf("**%s Audit Complete**", a.name)
	a.message = append(a.message, s)
}

func (a *Audit) Audit(t auditType, format string, b ...interface{}) {
	s := fmt.Sprintf(format, b...)
	log.Debug("%s Audit of %s", a.name, s)
	switch t {
	case Deployed:
		a.deployedBuffer = append(a.deployedBuffer, s)
	case Configured:
		a.configuredBuffer = append(a.configuredBuffer, s)
	case Mismatched:
		a.mismatchedBuffer = append(a.mismatchedBuffer, s)
	}
}

func (a *Audit) FreeFormAudit(format string, b ...interface{}) {
	freeFormAuditBuffer = append(freeFormAuditBuffer, fmt.Sprintf(format, b...))
}

func (a *Audit) auditFormatBuffer(t auditType, b []string) string {
	m := ""
	switch t {
	case Deployed:
		if !a.printDeployed {
			return m
		}
		m += fmt.Sprintf("**`Rogue %ss`**\r\n", a.name)
	case Configured:
		if !a.printConfigured {
			return m
		}
		m += "**`Configured but not created`**\r\n"
	case Mismatched:
		if !a.printMismatched {
			return m
		}
		m += "**`Mismatches`**\r\n"
	}
	if len(b) == 0 {
		m += "\n> "
		m += "No Differences Found\n"
		return m
	}
	for _, v := range b {
		m += "\t" + v + "\n"
	}
	m += "\n"
	return m
}

func (a *Audit) auditFormat(appName string) string {
	m := ""
	switch appName {
	case "arc":
		m += a.message[0] + "\n> "
		if len(a.deployedBuffer) == 0 && len(a.configuredBuffer) == 0 && len(a.mismatchedBuffer) == 0 {
			m += "No Differences Found\n\n"
			m += a.message[1]
			return m
		}
		m += a.auditFormatBuffer(Deployed, a.deployedBuffer) + "\n> "
		m += a.auditFormatBuffer(Configured, a.configuredBuffer) + "\n> "
		m += a.auditFormatBuffer(Mismatched, a.mismatchedBuffer) + "\n"
		if len(m) > 7000 {
			m = m[:7000] + "...\n\n"
		}
		m += a.message[1]
	case "audit":
		m += a.message[0] + "\n >"
		for _, v := range freeFormAuditBuffer {
			m += v
		}
		m += "\n> "
		m += a.message[1]
	}
	return m
}

func (a *Audit) auditMessage(b []string) {
	if len(b) == 0 {
		msg.Detail("No Differences Found")
		return
	}
	for _, v := range b {
		msg.Detail(v)
	}
}

func PostAudit(appName string) {
	if notification == nil {
		return
	}
	token := env.Lookup("SPARK_TOKEN")
	if token == "" {
		log.Warn("No spark token available")
		return
	}
	sparkClient, err := spark.New(token, notification.Spark.Rooms["audit"], spark.Html)
	if err != nil {
		msg.Error(err.Error())
		return
	}
	switch appName {
	case "arc":
		for _, v := range AuditBuffer {
			if v.printDeployed {
				msg.Info("Rogue %ss", v.name)
				v.auditMessage(v.deployedBuffer)
			}
			if v.printConfigured {
				msg.Info("Configured but not created")
				v.auditMessage(v.configuredBuffer)
			}
			if v.printMismatched {
				msg.Info("Mismatches")
				v.auditMessage(v.mismatchedBuffer)
			}
		}
		// Sends the formatted audit information to the audit spark room
		for _, v := range AuditBuffer {
			_, err = fmt.Fprintf(sparkClient, v.auditFormat(appName))
			if err != nil {
				msg.Error(err.Error())
			}
		}
	case "audit":
		// Sends the formatted audit information to the audit spark room
		for _, v := range AuditBuffer {
			_, err = fmt.Fprintf(sparkClient, v.auditFormat(appName))
			if err != nil {
				msg.Error(err.Error())
			}
		}
	}
}
