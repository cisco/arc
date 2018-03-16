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

package amp

import (
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type amp struct {
	*resource.Resources
	*config.Amp
	identityManagement *identityManagement
	storage            *storage
	keyManagement      *keyManagement
}

func New(cfg *config.Amp) (*amp, error) {
	log.Info("Initializing Amp: %q", cfg.Name())

	if cfg.Storage == nil {
		return nil, fmt.Errorf("The storage element is missing from the amp configuration.")
	}

	a := &amp{
		Amp: cfg,
	}
	a.header()
	var err error

	a.identityManagement, err = newIdentityManagement(a, cfg.IdentityManagement)
	if err != nil {
		return nil, err
	}

	a.storage, err = newStorage(a, cfg.Storage)
	if err != nil {
		return nil, err
	}

	a.keyManagement, err = newKeyManagement(a, cfg.KeyManagement)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Run starts amp processing. It returns 0 for success, 1 for failure.
// Upon failure err might be set to a non-nil value.
func (a *amp) Run() (int, error) {
	u, err := user.Current()
	if err != nil {
		return 1, err
	}

	// Create base request
	req := route.NewRequest(a.Name(), u.Username, time.Now().UTC().String())

	// Parse the request from the command line.
	req.Parse(os.Args[2:])
	log.Info("Creating %s request for user %q", req, u.Username)

	// Load the data from the provider unless there is a Load, Help or Config command.
	switch req.Command() {
	case route.None, route.Load:
		// Invalid commands: issue a help command.
		req.SetCommand(route.Help)
		a.Route(req)
		return 1, nil
	case route.Help, route.Config:
		// Skip the loading for the help command since we aren't going to
		// interact with the provider.
		break
	default:
		if req.TestFlag() {
			break
		}
		log.Info("Loading amp: %q", a.Name())
		if resp := a.Route(req.Clone(route.Load)); resp != route.OK {
			return 1, fmt.Errorf("Failed to load account for %s", a.Name())
		}
		log.Info("Loading complete")
	}
	log.Info("Routing request: %q", req)
	if resp := a.Route(req); resp != route.OK {
		log.Info("Exiting, %s request failed\n", req)
		return 1, nil
	}
	log.Info("Exiting successfully\n")
	return 0, nil
}

func (a *amp) IdentityManagement() resource.IdentityManagement {
	return a.identityManagement
}

func (a *amp) Storage() resource.Storage {
	return a.storage
}

func (a *amp) KeyManagement() resource.KeyManagement {
	return a.keyManagement
}

func (a *amp) Route(req *route.Request) route.Response {
	log.Route(req, "Amp")

	// Route to the appropriate resource
	switch req.Top() {
	case "":
		break
	case "storage":
		return a.storage.Route(req.Pop())
	case "bucket", "bucket_set":
		return a.storage.Route(req)
	case "key_management":
		return a.keyManagement.Route(req.Pop())
	case "key", "encryption_key":
		return a.keyManagement.Route(req)
	case "identity_management":
		return a.identityManagement.Route(req.Pop())
	case "policy":
		return a.identityManagement.Route(req)
	}

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load:
		if resp := a.identityManagement.Route(req); resp != route.OK {
			return resp
		}
		if resp := a.storage.Route(req); resp != route.OK {
			return resp
		}
		if resp := a.keyManagement.Route(req); resp != route.OK {
			return resp
		}
		return route.OK
	case route.Info:
		a.identityManagement.Info()
		a.storage.Info()
		a.keyManagement.Info()
		return route.OK
	case route.Config:
		a.Print()
		return route.OK
	case route.Audit:
	case route.Help:
		Help()
	default:
		msg.Error("Error: amp/amp.go Unknown command " + req.Command().String())
		Help()
		return route.FAIL
	}
	return route.OK
}

func Help() {
	commands := []help.Command{
		{Name: "storage", Desc: "manage storage"},
		{Name: "bucket 'name'", Desc: "manage named bucket"},
		{Name: "bucket_set 'name'", Desc: "manage named bucket"},
		{Name: "key_management 'name'", Desc: "manage key management"},
		{Name: "encryption_key 'name'", Desc: "manage named key"},
		{Name: "key 'name'", Desc: "manage named key"},
		{Name: "identity_management", Desc: "manage identity management"},
		{Name: "policy 'name'", Desc: "manage named policy"},
		{Name: route.Info.String(), Desc: "show information about allocation amp resources"},
		{Name: route.Config.String(), Desc: "show the amp configuration for the given account"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("", commands)
}

func (a *amp) header() {
	msg.Heading("amp, %s", env.Lookup("VERSION"))
}
