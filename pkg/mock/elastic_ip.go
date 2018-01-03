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

// elasticIP implements the resource.ProviderElasticIP interface.
type elasticIP struct {
	*mock
	instance resource.Instance
	id       string
}

// newElasticIP constructs the mock elastic IP.
func newElasticIP(i resource.Instance, p *dataCenterProvider) (resource.ProviderElasticIP, error) {
	log.Info("Initializing mock elasticIP")
	return &elasticIP{
		mock:     newMock("elasticIP", p.Provider),
		instance: i,
		id:       "0xdeadc0de",
	}, nil
}

// Id returns the allocationId of the elastic IP
func (e *elasticIP) Id() string {
	return ""
}

// InstanceId returns the id of the instance the elastic IP is or will
// be associated with.
func (e *elasticIP) Instance() resource.Instance {
	return e.instance
}

// IpAddress returns the elastic IP address.
func (e *elasticIP) IpAddress() string {
	return ""
}

// Attached returns true if the elastic IP is associated with an instance.
func (e *elasticIP) Attached() bool {
	return true
}

// Detached returns true if the elastic IP is not associated with an instance.
func (e *elasticIP) Detached() bool {
	return true
}

// Load elastic IP from the provider.
func (e *elasticIP) Load() error {
	return nil
}

// Create allocates the elastic IP.
func (e *elasticIP) Create() error {
	return nil
}

// Attach associates the allocated elastic IP to the instance.
func (e *elasticIP) Attach() error {
	return nil
}

// Detach disassocates the allocated elastic IP from the instance.
func (e *elasticIP) Detach() error {
	return nil
}

// Destroy releases the elastic IP.
func (e *elasticIP) Destroy() error {
	return nil
}
