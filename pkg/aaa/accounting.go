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
	"strings"

	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/spark"
)

var accountingBuffer []string

func PreAccounting(args []string) {
	if notification == nil {
		return
	}
	version := env.Lookup("VERSION")
	sshUser := env.Lookup("SSH_USER")
	userId := env.Lookup("USER")

	trimmedVersion := strings.Split(version, " ")
	command := strings.Join(args, " ")
	username := userId
	if sshUser != "" && sshUser != userId {
		username += "(" + sshUser + ")"
	}
	s := fmt.Sprintf("**%s | %s | %s**\n> ", username, trimmedVersion[0], command)
	accountingBuffer = append(accountingBuffer, s)
}

func Accounting(format string, a ...interface{}) {
	if notification == nil {
		return
	}
	s := fmt.Sprintf(format, a...)
	m := fmt.Sprintf("%s\n", s)
	accountingBuffer = append(accountingBuffer, m)
}

func catErrors(errorList []string) string {
	if len(errorList) >= 5 {
		return "Too many errors, see log file"
	}
	m := ""
	for _, v := range errorList {
		m += v + "\n"
	}
	return m
}

func PostAccounting(result int) {
	if notification == nil {
		return
	}
	token := env.Lookup("SPARK_TOKEN")
	if token == "" {
		msg.Warn("No spark token available")
		return
	}
	sparkClient, err := spark.New(token, notification.Spark.Rooms["accounting"], spark.Html)
	if err != nil {
		msg.Error(err.Error())
		return
	}
	resultMessage := "\n\n**Success**"

	m := ""
	for _, b := range accountingBuffer {
		m += b + "\n> "
	}
	m += "\r\n"
	if result != 0 {
		m += catErrors(msg.LastError())
		resultMessage = "\n\n**Failure**"
	}
	if len(m) > 7200 {
		m = m[:7200]
	}
	m += resultMessage
	_, err = fmt.Fprintf(sparkClient, m)
	if err != nil {
		msg.Error(err.Error())
	}
}
