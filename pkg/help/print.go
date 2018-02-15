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

package help

import (
	"fmt"
)

var header string

func Init(appName string, configType string) {
	header = "\n%s is a tool for managing %s resources.\n\n" +
		"Usage:\n\n" +
		"  %s [%s]%%s [command]\n\n" +
		"The %s configuration files are found in /etc/arc/[%s].json.\n\n" +
		"The commands are:\n\n"
	header = fmt.Sprintf(header, appName, configType, appName, configType, configType, configType)
}

type Resource struct {
	Name string
	Desc string
}

type Command struct {
	Name string
	Desc string
}

func Print(request string, commands []Command) {
	if request != "" {
		request = " " + request
	}
	fmt.Printf(header, request)
	for _, cmd := range commands {
		fmt.Printf("  %-18s %s\n", cmd.Name, cmd.Desc)
	}
}

func Append(dest, src []Command) []Command {
	for _, c := range src {
		dest = append(dest, c)
	}
	return dest
}
