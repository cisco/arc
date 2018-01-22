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

package msg

import (
	"fmt"
	"os"

	"github.com/cisco/arc/pkg/log"
)

var lastError []string

var err string
var warn string
var heading string
var info string
var clear string

var quiet bool

func init() {
	if os.Getenv("color") != "no" {
		err = "\033[33;31m"
		warn = "\033[33;35m"
		heading = "\033[33;33m"
		info = "\033[33;32m"
		clear = "\033[33;0m"
	}
}

const tab = "  "

var indent string = ""

func Quiet(q bool) {
	quiet = q
}

func GetQuiet() bool {
	return quiet
}

func Error(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	log.Error("%s%s", indent, s)
	if !quiet {
		fmt.Printf("\n%s%sError:%s %s\n", indent, err, clear, s)
	}

	t := removeExtraSpaces(s)
	lastError = append(lastError, fmt.Sprintf("\n> `Error:` %s\n\n", t))
}

func Warn(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	log.Warn("%s%s%s", indent, tab, s)
	if !quiet {
		fmt.Printf("\n%s%s%sWarning:%s %s\n", indent, warn, tab, clear, s)
	}
}

func Heading(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	log.Debug("%s%s", indent, s)
	if !quiet {
		fmt.Printf("\n%s%s%s%s\n", indent, heading, s, clear)
	}
}

func Info(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	log.Debug("%s%s", indent, s)
	if !quiet {
		fmt.Printf("\n%s%s%s%s\n", indent, info, s, clear)
	}
}

func Detail(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	log.Debug("%s%s%s", tab, indent, s)
	if !quiet {
		fmt.Printf("%s%s%s\n", tab, indent, s)
	}
}

func Raw(format string, a ...interface{}) {
	if quiet {
		return
	}
	fmt.Printf(format, a...)
}

func Indent() string {
	return indent
}

func Tab() string {
	return tab
}

func IndentInc() {
	indent += tab
}

func IndentDec() {
	if len(indent) >= len(tab) {
		end := len(indent) - len(tab)
		indent = indent[:end]
	}
}

func LastError() []string {
	return lastError
}

// removeExtraSpaces goes through the given string s and removes any
// extra spaces from the string as well as any tabs indicated with a \t
func removeExtraSpaces(s string) string {
	newString := ""
	var prev byte
	for i, v := range s {
		if v == ' ' && prev == ' ' {
			prev = s[i]
			continue
		}
		prev = s[i]
		if v == '\t' {
			continue
		}
		// If the value is \n then we need to add \n> to keep the formatting
		// of the spark message the same.
		if v == '\n' {
			newString += "\n> "
		}
		newString += string(v)
	}
	if len(newString) > 1000 {
		substring := newString[:999]
		substring += "..."
		return substring
	}
	return newString
}
