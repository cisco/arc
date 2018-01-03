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

package mock

import (
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/resource"
)

type role struct {
	*mock
	name       string
	id         string
	instanceId string
}

func newRole(name string, p *dataCenterProvider, in resource.Instance) (resource.ProviderRole, error) {
	log.Info("Initializing mock role")
	i := &role{
		mock:       newMock("role", p.Provider),
		name:       name,
		id:         "0x1127beef",
		instanceId: "0x1127beef",
	}
	return i, nil
}

func (r *role) Id() string {
	return r.id
}

func (r *role) InstanceId() string {
	return r.instanceId
}

func (r *role) Attached() bool {
	return true
}

func (r *role) Detached() bool {
	return true
}

func (r *role) Load() error {
	return nil
}

func (r *role) Attach() error {
	return nil
}

func (r *role) Detach() error {
	return nil
}

func (r *role) Destroy() error {
	return nil
}

func (r *role) Update() error {
	return nil
}
