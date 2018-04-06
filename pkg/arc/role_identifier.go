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

type roleIdentifier struct {
	name                   string
	providerRoleIdentifier resource.ProviderRoleIdentifier
}

func newRoleIdentifier(in resource.Instance, prov provider.DataCenter, name string) (*roleIdentifier, error) {
	log.Debug("Initializing Role Identifier '%s'", name)

	r := &roleIdentifier{
		name: name,
	}

	var err error
	r.providerRoleIdentifier, err = prov.NewRoleIdentifier(r, name, in)
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *roleIdentifier) Route(req *route.Request) route.Response {
	log.Route(req, "Role %q", r.Name())
	return route.FAIL
}

func (r *roleIdentifier) Load() error {
	return r.providerRoleIdentifier.Load()
}

func (r *roleIdentifier) Created() bool {
	return r.providerRoleIdentifier.Created()
}

func (r *roleIdentifier) Destroyed() bool {
	return r.providerRoleIdentifier.Destroyed()
}

func (r *roleIdentifier) Name() string {
	return r.name
}

func (r *roleIdentifier) Id() string {
	return r.providerRoleIdentifier.Id()
}

func (r *roleIdentifier) InstanceId() string {
	return r.providerRoleIdentifier.InstanceId()
}

func (r *roleIdentifier) Attached() bool {
	return r.providerRoleIdentifier.Attached()
}

func (r *roleIdentifier) Detached() bool {
	return r.providerRoleIdentifier.Detached()
}

func (r *roleIdentifier) Detach() error {
	if r.Name() == "" {
		return nil
	}
	msg.Info("Role Identifier Detach: %s", r.Name())
	return r.providerRoleIdentifier.Detach()
}

func (r *roleIdentifier) Attach() error {
	if r.Name() == "" {
		return nil
	}
	msg.Info("Role Identifier Attach: %s", r.Name())
	return r.providerRoleIdentifier.Attach()
}

func (r *roleIdentifier) Update() error {
	return r.providerRoleIdentifier.Update()
}

func (r *roleIdentifier) ProviderRoleIdentifier() resource.ProviderRoleIdentifier {
	return r.providerRoleIdentifier
}
