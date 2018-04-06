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

type role struct {
	*config.Role
	identityManagement *identityManagement
	providerRole       resource.ProviderRole
}

func newRole(cfg *config.Role, identityManagement *identityManagement, prov provider.IdentityManagement) (*role, error) {
	log.Debug("Initializing Role, %q", cfg.Name())
	r := &role{
		Role:               cfg,
		identityManagement: identityManagement,
	}

	var err error
	r.providerRole, err = prov.NewRole(r, cfg)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Route satisfies the embedded resource.Resource interface in resource.Role.
// Bucket handles load, create, destroy, config and info requests by delegating them
// to the providerRole.
func (r *role) Route(req *route.Request) route.Response {
	log.Route(req, "Role %q", r.Name())
	switch req.Command() {
	case route.Load:
		if err := r.Load(); err != nil {
			return route.FAIL
		}
		return route.OK
	case route.Create:
		if err := r.Create(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Destroy:
		if err := r.Destroy(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Provision:
		if err := r.Provision(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Audit:
		if err := r.Audit(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Info:
		r.Info()
		return route.OK
	case route.Config:
		r.Print()
		return route.OK
	case route.Help:
		r.Help()
		return route.OK
	default:
		msg.Error("Internal Error: Unknown command " + req.Command().String())
		r.Help()
		return route.FAIL
	}
}

// Created satisfies the embedded resource.Creator interface in resource.Role.
// It delegates the call to the provider's role.
func (r *role) Created() bool {
	return r.providerRole.Created()
}

// Destroyed satisfies the embedded resource.Destroyer interaface in resource.Role.
// It delegates the call to the provider's role.
func (r *role) Destroyed() bool {
	return r.providerRole.Destroyed()
}

func (r *role) IdentityManagement() resource.IdentityManagement {
	return r.identityManagement
}

func (r *role) ProviderRole() resource.ProviderRole {
	return r.providerRole
}

func (r *role) Provision(flags ...string) error {
	return r.providerRole.Provision(flags...)
}

func (r *role) Load() error {
	return r.providerRole.Load()
}

func (r *role) Create(flags ...string) error {
	if r.Created() {
		msg.Detail("Role exists, skipping...")
		return nil
	}
	return r.ProviderRole().Create(flags...)
}

func (r *role) Destroy(flags ...string) error {
	if r.Destroyed() {
		msg.Detail("Role does not exist, skipping...")
		return nil
	}
	return r.ProviderRole().Destroy(flags...)
}

func (r *role) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	return r.ProviderRole().Audit(flags...)
}

func (r *role) Info() {
	if r.Destroyed() {
		return
	}
	r.ProviderRole().Info()
}

func (r *role) Help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: fmt.Sprintf("create role %s", r.Name())},
		{Name: route.Destroy.String(), Desc: fmt.Sprintf("destroy role %s", r.Name())},
		{Name: route.Audit.String(), Desc: fmt.Sprintf("audit role %s", r.Name())},
		{Name: route.Info.String(), Desc: "show information about allocated role"},
		{Name: route.Config.String(), Desc: "show the configuration for the given role"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("role "+r.Name(), commands)
}
