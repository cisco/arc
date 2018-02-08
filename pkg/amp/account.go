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

	// "github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"

	"github.com/cisco/arc/pkg/aws"
	//"github.com/cisco/arc/pkg/gcp"
	//"github.com/cisco/arc/pkg/azure"
)

type account struct {
	*resource.Resources
	*config.Account
	amp     *amp
	storage *storage
}

// newAccount is the constructor for a account object. It returns a non-nil error upon failure.
func newAccount(amp *amp, cfg *config.Account) (*account, error) {
	log.Debug("Initializing Account")

	// Validate the config.Account object.
	if cfg.Provider == nil {
		return nil, fmt.Errorf("The provider element is missing from the account configuration")
	}
	if cfg.SecurityTags() == nil {
		return nil, fmt.Errorf("The security tags element is missing from the account configuration")
	}
	if cfg.Storage == nil {
		return nil, fmt.Errorf("The storage element is missing from the account configuration")
	}

	a := &account{
		Resources: resource.NewResources(),
		Account:   cfg,
		amp:       amp,
	}

	vendor := cfg.Provider.Vendor
	var err error
	var p provider.Account

	switch vendor {
	case "aws":
		p, err = aws.NewAccountProvider(cfg)
	//case "azure":
	//	p, err = azure.NewAccountProvider(cfg)
	//case "gcp":
	//	p, err = gcp.NewAccountProvider(cfg)
	default:
		err = fmt.Errorf("Unknown vendor %q", vendor)
	}
	if err != nil {
		return nil, err
	}

	a.storage, err = newStorage(a, p, cfg.Storage)
	if err != nil {
		return nil, err
	}
	a.Append(a.storage)

	return a, nil
}

// Amp satisfies the resource.Storage interface and provides access
// to account's parent.
func (a *account) Amp() resource.Amp {
	return a.amp
}

// Storage satisfies the resource.Account interface and provides access
// to account's children.
func (a *account) Storage() resource.Storage {
	return a.storage
}

// Route satisfies the embedded resource.Resource interface in resource.Account.
// Account does not directly terminate a request so only handles load and info
// requests from it's parent.  All other commands are routed to account's children.
func (a *account) Route(req *route.Request) route.Response {
	log.Route(req, "Account")

	// Route to the appropriate resource
	switch req.Top() {
	case "":
		break
	case "storage", "bucket":
		return a.storage.Route(req)
	}

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Load:
		return a.RouteInOrder(req)
	case route.Info:
		a.info(req)
		return route.OK
	case route.Audit:
		return a.RouteInOrder(req)
	}
	msg.Error("Internal Error: amp/account.go. Unknown command %s", req.Command())
	return route.FAIL
}

func (a *account) info(req *route.Request) {
	msg.Info("Account")
	msg.IndentInc()
	a.RouteInOrder(req)
	msg.IndentDec()
}
