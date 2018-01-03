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
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/resource"
)

// instance implements the resource.ProviderInstance interface.
type instance struct {
	*mock
	*config.Instance
	id               string
	imageId          string
	keyname          string
	state            string
	privateIPAddress string
	publicIPAddress  string
}

func (i *instance) Role() *resource.Role {
	return nil
}

func (i *instance) Id() string {
	return i.id
}

func (i *instance) ImageId() string {
	return i.imageId
}

func (i *instance) KeyName() string {
	return i.keyname
}

func (i *instance) State() string {
	return i.state
}

func (i *instance) PrivateIPAddress() string {
	return i.privateIPAddress
}

func (i *instance) PublicIPAddress() string {
	return i.publicIPAddress
}

func (i *instance) Started() bool {
	return true
}

func (i *instance) Stopped() bool {
	return false
}

func (i *instance) SetTags(t map[string]string) error {
	return nil
}

func (i *instance) Audit(flags ...string) error {
	return nil
}

// newInstance constructs the mock instance.
func newInstance(cfg *config.Instance, p *dataCenterProvider) (resource.ProviderInstance, error) {
	log.Info("Initializing mock instance")
	i := &instance{
		mock:             newMock("instance", p.Provider),
		Instance:         cfg,
		id:               "0xdeadbeef",
		imageId:          "0xcab01dab",
		keyname:          "id_rsa",
		state:            "available",
		privateIPAddress: "192.168.0.1",
		publicIPAddress:  "34.33.32.31",
	}
	return i, nil
}
