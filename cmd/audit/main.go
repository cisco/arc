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

package main

// Create the version in version.go
//go:generate ../../bin/generate_version

import (
	"fmt"
	"os"
	"path"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/users"
)

func main() {
	appname := path.Base(os.Args[0])

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("%s %s\n", appname, version)
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "help" || len(os.Args) == 1 {
		Help()
		return
	}

	if err := env.Init(appname, version); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	if err := log.Init(appname); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
	defer log.Fini()

	notifications := &config.Notifications{
		Spark: &config.Spark{
			Rooms: map[string]string{
				"accounting": "44a5ce00-6d69-11e7-ba3f-671fa0a6f5e1",
				"audit":      "e6bb8e90-881d-11e7-a4c0-8f2917788fe6",
			},
		},
	}
	if err := aaa.Init(notifications); err != nil {
		msg.Error(err.Error())
		os.Exit(1)
	}
	err := users.Init(env.Lookup("ROOT") + "/etc/arc/users.json")
	if err != nil {
		msg.Error(err.Error())
		os.Exit(1)
	}

	aaa.PreAccounting(os.Args)
	a := NewAudit()

	err = a.Run(os.Args[1])
	if err != nil {
		msg.Error(err.Error())
		os.Exit(1)
	}
	aaa.PostAccounting(0)
	aaa.PostAudit(appname)
}
