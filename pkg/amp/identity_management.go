//
// Copyright (c) 2018, Cisco Systems
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

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"

	_ "github.com/cisco/arc/pkg/aws"
)

type identityManagement struct {
	*resource.Resources
	*config.IdentityManagement
	amp                        *amp
	policies                   []*policy
	providerIdentityManagement resource.ProviderIdentityManagement
}

// newIdentityManagement is the constructor for a identityManagement object. It returns a non-nil error upon failure.
func newIdentityManagement(amp *amp, cfg *config.IdentityManagement) (*identityManagement, error) {
	log.Debug("Initializing Identity Management")

	// Validate the config.IdentityManagement object.
	if cfg.Policies == nil {
		return nil, fmt.Errorf("The Policies element is missing from the iam configuration")
	}

	i := &identityManagement{
		Resources:          resource.NewResources(),
		IdentityManagement: cfg,
		amp:                amp,
	}

	prov, err := provider.NewIdentityManagement(amp.Amp)
	if err != nil {
		return nil, err
	}
	i.providerIdentityManagement, err = prov.NewIdentityManagement(cfg)
	if err != nil {
		return nil, err
	}
	for _, conf := range cfg.Policies {
		policy, err := newPolicy(conf, i, prov)
		if err != nil {
			return nil, err
		}
		i.policies = append(i.policies, policy)
	}

	return i, nil
}

// Amp satisfies the resource.IdentityManagement interface and provides access
// to identityManagement's parent.
func (i *identityManagement) Amp() resource.Amp {
	return i.amp
}

// FindPolicy returns the policy with the given name.
func (i *identityManagement) FindPolicy(name string) resource.Policy {
	for _, policy := range i.policies {
		if name == policy.Name() {
			return policy
		}
	}
	return nil
}

func (i *identityManagement) ProviderIdentityManagement() resource.ProviderIdentityManagement {
	return i.providerIdentityManagement
}

// Route satisfies the embedded route.Router interface in resource.IdentityManagement.
// IdentityManagement does not directly terminate a request so only handles load, config, and info
// requests from it's parent.  All other commands are routed to identityManagement's children.
func (i *identityManagement) Route(req *route.Request) route.Response {
	log.Route(req, "Identity Management")

	// Route to the appropriate resource
	switch req.Top() {
	case "":
		break
	case "policy":
		req.Pop()
		if req.Top() == "" {
			i.Help()
			return route.FAIL
		}
		policy := i.FindPolicy(req.Top())
		if policy == nil {
			msg.Error("Unknown policy %q.", req.Top())
			return route.FAIL
		}
		if req.Command() == route.Audit {
			aaa.NewAudit("Policy")
		}
		req.Flags().Append("Policy")
		return policy.Route(req)
	default:
		i.Help()
		return route.FAIL
	}

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Info:
		i.Info()
		return route.OK
	case route.Config:
		i.Print()
		return route.OK
	case route.Load:
		if err := i.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Provision:
		// if err := i.Provision(req.Flags().Get()...); err != nil {
		// return route.FAIL
		// }
		return route.OK
	case route.Audit:
		if err := i.Audit("Policy"); err != nil {
			return route.FAIL
		}
		return route.OK
	case route.Help:
		i.Help()
		return route.OK
	default:
		msg.Error("Internal Error: amp/identity_management.go Unknown command " + req.Command().String())
		i.Help()
		return route.FAIL
	}
}

func (i *identityManagement) Load() error {
	for _, p := range i.policies {
		if err := p.Load(); err != nil {
			return err
		}
	}
	return nil
}

//This will be used for when roles are groups can be managed with amp
/*
func (i *identityManagement) Provision(flags ...string) error {
	return nil
}
*/

func (i *identityManagement) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	for _, v := range flags {
		err := aaa.NewAudit(v)
		if err != nil {
			return err
		}
	}
	if err := i.providerIdentityManagement.Audit("Policy"); err != nil {
		return err
	}
	for _, p := range i.policies {
		log.Debug("Audit of policy %q", p.Name())
		if err := p.Audit("Policy"); err != nil {
			return err
		}
	}
	return nil
}

func (i *identityManagement) Info() {
	msg.Info("Identity Management")
	msg.IndentInc()
	msg.Info("Policies")
	msg.IndentInc()
	for _, p := range i.policies {
		p.Info()
	}
	msg.IndentDec()
	msg.IndentDec()
	log.Debug("Info complete")
}

func (i *identityManagement) Help() {
	commands := []help.Command{
		// {Name: route.Provision.String(), Desc: "update identityManagement"},
		{Name: route.Audit.String(), Desc: "audit identityManagement"},
		{Name: route.Info.String(), Desc: "show information about allocated identityManagement"},
		{Name: route.Config.String(), Desc: "show the configuration for the given identityManagement"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("identity_management", commands)
}
