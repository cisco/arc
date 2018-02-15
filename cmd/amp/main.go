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
	"github.com/cisco/arc/pkg/amp"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
)

func main() {
	appName := path.Base(os.Args[0])

	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Printf("%s %s\n", appName, version)
		return
	}
	if (len(os.Args) > 1 && os.Args[1] == "help") || len(os.Args) < 3 {
		amp.Help()
		return
	}

	if err := env.Init(appName, version); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	if err := log.Init(appName); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}
	defer log.Fini()

	cfg, err := config.NewAmp(os.Args[1])
	if err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	if err := aaa.Init(cfg.Notifications); err != nil {
		fmt.Printf(err.Error())
		os.Exit(1)
	}

	aaa.PreAccounting(os.Args)
	a, err := amp.New(cfg)
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
	aaa.PostAudit(appName)
}

func exit(err error) {
	msg.Error(err.Error())
	aaa.PostAccounting(1)
	os.Exit(1)
}
