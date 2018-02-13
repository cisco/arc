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

func Init(appName string) {
	switch appName {
	case "arc":
		header = "\narc is a tool for managing datacenter resources.\n\n" +
			"Usage:\n\n" +
			"  arc [datacenter]%s [command]\n\n" +
			"The datacenter configuration files are found in /etc/arc/[datacenter].json.\n\n"
	case "amp":
		header = "\namp is a tool for managing account resources.\n\n" +
			"Usage:\n\n" +
			"  amp <account> %s <command>\n\n" +
			"The account configuration files are found in /etc/arc/[account].json.\n\n"
	}
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
	fmt.Printf("The commands are:\n\n")
	for _, cmd := range commands {
		fmt.Printf("  %-18s %s\n", cmd.Name, cmd.Desc)
	}
}

func PrintWithResources(request string, resources []Resource, commands []Command) {
	if request != "" {
		request = " " + request
	}
	fmt.Printf(header, request)
	fmt.Printf("The resources are:\n\n")
	for _, rsrc := range resources {
		fmt.Printf(" %-18s %s\n", rsrc.Name, rsrc.Desc)
	}
	fmt.Printf("\nThe commands are:\n\n")
	for _, cmd := range commands {
		fmt.Printf(" %-18s %s\n", cmd.Name, cmd.Desc)
	}
}

func Append(dest, src []Command) []Command {
	for _, c := range src {
		dest = append(dest, c)
	}
	return dest
}
