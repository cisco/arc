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

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type policy struct {
	*config.Policy
	identityManagement *identityManagement
	providerPolicy     resource.ProviderPolicy
}

func newPolicy(cfg *config.Policy, identityManagement *identityManagement, prov provider.IdentityManagement) (*policy, error) {
	log.Debug("Initializing Policy, %q", cfg.Name())
	p := &policy{
		Policy:             cfg,
		identityManagement: identityManagement,
	}

	var err error
	p.providerPolicy, err = prov.NewPolicy(p, cfg)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// Route satisfies the embedded resource.Resource interface in resource.Bucket.
// Bucket handles load, create, destroy, config and info requests by delegating them
// to the providerBucket.
func (p *policy) Route(req *route.Request) route.Response {
	log.Route(req, "Policy %q", p.Name())
	switch req.Command() {
	case route.Create:
		if err := p.Create(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Destroy:
		if err := p.Destroy(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Provision:
		if err := p.Provision(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Audit:
		if err := p.Audit(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Info:
		p.Info()
		return route.OK
	case route.Config:
		p.Print()
		return route.OK
	case route.Help:
		p.Help()
		return route.OK
	default:
		msg.Error("Internal Error: Unknown command " + req.Command().String())
		p.Help()
		return route.FAIL
	}
}

// Created satisfies the embedded resource.Resource interface in resource.Bucket.
// It delegates the call to the provider's bucket.
func (p *policy) Created() bool {
	return p.providerPolicy.Created()
}

// Destroyed satisfies the embedded resource.Resource interaface in resource.Bucket.
// It delegates the call to the provider's bucket.
func (p *policy) Destroyed() bool {
	return p.providerPolicy.Destroyed()
}

func (p *policy) IdentityManagement() resource.IdentityManagement {
	return p.identityManagement
}

func (p *policy) ProviderPolicy() resource.ProviderPolicy {
	return p.providerPolicy
}

func (p *policy) Load() error {
	return p.providerPolicy.Load()
}

func (p *policy) Create(flags ...string) error {
	if p.Created() {
		msg.Detail("Policy exists, skipping...")
		return nil
	}
	return p.ProviderPolicy().Create(flags...)
}

func (p *policy) Destroy(flags ...string) error {
	if p.Destroyed() {
		msg.Detail("Policy does not exist, skipping...")
		return nil
	}
	return p.ProviderPolicy().Destroy(flags...)
}

func (p *policy) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	return p.ProviderPolicy().Audit(flags...)
}

func (p *policy) Provision(flags ...string) error {
	if p.Destroyed() {
		msg.Detail("Bucket does not exist, skipping...")
		return nil
	}
	provisionFlags := map[string]bool{}
	if len(flags) != 0 {
		for _, v := range flags {
			if provisionFlags[v] == false {
				provisionFlags[v] = true
			}
		}
	}
	return nil
}

func (p *policy) Info() {
	log.Debug("Amp Policy Info start")
	if p.Destroyed() {
		return
	}
	p.ProviderPolicy().Info()
	log.Debug("Amp Policy Info complete")
}

func (p *policy) Help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: fmt.Sprintf("create policy %s", p.Name())},
		{Name: route.Destroy.String(), Desc: fmt.Sprintf("destroy policy %s", p.Name())},
		{Name: route.Provision.String(), Desc: fmt.Sprintf("update the policy for %s", p.Name())},
		{Name: route.Audit.String(), Desc: fmt.Sprintf("audit policy %s", p.Name())},
		{Name: route.Info.String(), Desc: "show information about allocated policy"},
		{Name: route.Config.String(), Desc: "show the configuration for the given policy"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("policy", commands)
}
