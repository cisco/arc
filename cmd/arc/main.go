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
	"github.com/cisco/arc/pkg/arc"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/servertypes"
	"github.com/cisco/arc/pkg/users"
)

func main() {
	appname := path.Base(os.Args[0])

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("%s %s\n", appname, version)
		return
	}

	help.Init(appname)

	if (len(os.Args) > 1 && os.Args[1] == "help") || len(os.Args) < 3 {
		arc.Help()
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

	cfg, err := config.NewArc(os.Args[1])
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	if err := aaa.Init(cfg.Notifications); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	err = users.Init(env.Lookup("ROOT") + "/etc/arc/users.json")
	if err != nil {
		exit(err)
	}

	err = servertypes.Init()
	if err != nil {
		exit(err)
	}

	aaa.PreAccounting(os.Args)
	a, err := arc.New(cfg)
	if err != nil {
		exit(err)
	}

	result, err := a.Run()
	if err != nil {
		exit(err)
	}
	aaa.PostAccounting(result)
	if result != 0 {
		os.Exit(result)
	}
	aaa.PostAudit(appname)
}

func exit(err error) {
	msg.Error(err.Error())
	aaa.PostAccounting(1)
	os.Exit(1)
}
