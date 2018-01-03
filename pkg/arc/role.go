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

package arc

import (
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type role struct {
	name         string
	providerRole resource.ProviderRole
}

func newRole(in resource.Instance, prov provider.DataCenter, name string) (*role, error) {
	log.Debug("Initializing Role '%s'", name)

	r := &role{
		name: name,
	}

	var err error
	r.providerRole, err = prov.NewRole(r, name, in)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *role) Route(req *route.Request) route.Response {
	log.Route(req, "Role %q", r.Name())
	return route.FAIL
}

func (r *role) Load() error {
	return r.providerRole.Load()
}

func (r *role) Created() bool {
	return r.providerRole.Created()
}

func (r *role) Destroyed() bool {
	return r.providerRole.Destroyed()
}

func (r *role) Name() string {
	return r.name
}

func (r *role) Id() string {
	return r.providerRole.Id()
}

func (r *role) InstanceId() string {
	return r.providerRole.InstanceId()
}

func (r *role) Attached() bool {
	return r.providerRole.Attached()
}

func (r *role) Detached() bool {
	return r.providerRole.Detached()
}

func (r *role) Detach() error {
	if r.Name() == "" {
		return nil
	}
	msg.Info("Role Detach: %s", r.Name())
	return r.providerRole.Detach()
}

func (r *role) Attach() error {
	if r.Name() == "" {
		return nil
	}
	msg.Info("Role Attach: %s", r.Name())
	return r.providerRole.Attach()
}

func (r *role) Update() error {
	return r.providerRole.Update()
}

func (r *role) ProviderRole() resource.ProviderRole {
	return r.providerRole
}
