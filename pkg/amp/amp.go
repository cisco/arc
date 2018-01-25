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
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type amp struct {
	*resource.Resources
	*config.Amp
	account *account
}

func New(cfg *config.Amp) (*amp, error) {
	log.Info("Initializing Amp: %q", cfg.Name())

	if cfg.Account == nil {
		return nil, fmt.Errorf("The account element is missing from the amp configuration.")
	}

	a := &amp{
		Amp: cfg,
	}
	a.header()
	var err error

	a.account, err = newAccount(a, cfg.Account)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *amp) Run() (int, error) {
	u, err := user.Current()
	if err != nil {
		return 1, err
	}

	// Create base request
	req := route.NewRequest(a.Name(), u.Username, time.Now().UTC().String())

	// Parse the request from thec ommand line.
	req.Parse(os.Args[2:])
	log.Info("Creating %s request for user %q", req, u.Username)

	// Load the data from the provider unless there is a Load, Help or Config command.
	switch req.Command() {
	case route.None:
		// Invalid commands: issue a help command.
		req.SetCommand(route.Help)
		a.Route(req)
		return 1, nil
	case route.Help:
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

func (a *amp) Account() resource.Account {
	return a.account
}

func (a *amp) Route(req *route.Request) route.Response {
	log.Route(req, "Amp")

	// Route to the appropriate resource
	switch req.Top() {
	case "":
		break
	case "storage":
		return a.account.storage.Route(req.Pop())
	case "bucket", "bucket_set":
		return a.account.storage.Route(req)
	}
	return route.OK
}

func Help() {
	help := `
  The following are all available usages of amp:
    amp <account> storage info
      prints out the information for all the buckets
      for the account.
    
    amp <account> storage config
      amp prints outs all the information for every bucket associated with
      the account from the corresponding json file.
    
    amp <account> storage audit
      amp checks for the following:
        -buckets that exist on the provider but don't exist in the corresponding json file.
        -buckets that exist in the json file but are not created with the provider.
    
    amp <account> bucket <bucket_name> info
      prints out the provider information for bucket_name.
    
    amp <account> storage bucket <bucket_name> info
      prints out the information for bucket_name.
    
    amp <account> bucket <bucket_name> config
      prints out all the information for bucket_name
    
    amp <account> storage bucket <bucket_name> config
      prints out all the information for bucket_name
    
    amp <account> bucket <bucket_name> create
      creates a bucket for bucket_name on the account using the
      configuration for it found in json file corresponding
      to the account.
    
    amp <account> storage bucket <bucket_name> create
      creates a bucket for bucket_name on the account using the
      configuration for it found in json file corresponding
      to the account.
    
    amp <account> bucket <bucket_name> delete
      deletes the bucket bucket_name.
    
    amp <account> storage bucket <bucket_name> delete
      deletes the bucket bucket_name.
    
    amp <account> storage update
      updates the tags on the buckets.

    amp <account> storage bucket <bucket_name> update
      updates the tags for bucket_name.
`
	fmt.Println(help)
}

func (a *amp) header() {
	msg.Heading("amp, %s", env.Lookup("VERSION"))
}
