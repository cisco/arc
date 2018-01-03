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

package route

type Command uint

const (
	None Command = iota
	Load
	Help
	Config
	Info
	Create
	Provision
	Start
	Stop
	Restart
	Replace
	Destroy
	Audit
)

var c2s = map[Command][]string{
	Load:      []string{"load"},
	Help:      []string{"help"},
	Config:    []string{"config"},
	Info:      []string{"info", "show", "list"},
	Create:    []string{"create"},
	Provision: []string{"provision", "refresh", "update"},
	Start:     []string{"start"},
	Stop:      []string{"stop"},
	Restart:   []string{"restart", "reboot"},
	Replace:   []string{"replace", "upgrade"},
	Destroy:   []string{"destroy", "delete", "nuke"},
	Audit:     []string{"audit"},
}

var s2c = map[string]Command{
	"":          None,
	"load":      Load,
	"help":      Help,
	"config":    Config,
	"info":      Info,
	"show":      Info,
	"list":      Info,
	"create":    Create,
	"provision": Provision,
	"refresh":   Provision,
	"update":    Provision,
	"start":     Start,
	"stop":      Stop,
	"restart":   Restart,
	"reboot":    Restart,
	"replace":   Replace,
	"upgrade":   Replace,
	"destroy":   Destroy,
	"delete":    Destroy,
	"nuke":      Destroy,
	"audit":     Audit,
}

func (c Command) String() string {
	cmdList := c2s[c]
	if cmdList == nil {
		return ""
	}
	return cmdList[0]
}
