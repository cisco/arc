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
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type elasticIP struct {
	providerElasticIP resource.ProviderElasticIP
}

// newElasticIP is the constructor for a elasticIP object. It returns a non-nil error upon failure.
func newElasticIP(i resource.Instance, prov provider.DataCenter) (*elasticIP, error) {
	log.Debug("Initializing ElasticIP for '%s'", i.Name())

	p := &elasticIP{}

	var err error
	p.providerElasticIP, err = prov.NewElasticIP(p, i)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Route satisfies the embedded resource.Resource interface in resource.ElasticIP.
// ElasticIP handles load, create, destroy, and info requests by delegating them
// to the providerElasticIP.
func (e *elasticIP) Route(req *route.Request) route.Response {
	log.Route(req, "ElasticIP")
	return route.FAIL
}

// Id provides the id of the provider specific elasticIP resource.
// This satisfies the resource.DynamicElasticIP interface.
func (e *elasticIP) Id() string {
	return e.providerElasticIP.Id()
}

func (e *elasticIP) Instance() resource.Instance {
	return e.providerElasticIP.Instance()
}

func (e *elasticIP) IpAddress() string {
	return e.providerElasticIP.IpAddress()
}

// Created satisfies the embedded resource.Resource interface in resource.ElasticIP.
// It delegates the call to the provider's elasticIP.
func (e *elasticIP) Created() bool {
	return e.providerElasticIP.Created()
}

// Destroyed satisfies the embedded resource.Resource interface in resource.ElasticIP.
// It delegates the call to the provider's elasticIP.
func (e *elasticIP) Destroyed() bool {
	return e.providerElasticIP.Destroyed()
}

func (e *elasticIP) Attached() bool {
	return e.providerElasticIP.Attached()
}

func (e *elasticIP) Detached() bool {
	return e.providerElasticIP.Detached()
}

func (e *elasticIP) Load() error {
	return e.providerElasticIP.Load()
}

func (e *elasticIP) Create() error {
	return e.providerElasticIP.Create()
}

func (e *elasticIP) Attach() error {
	return e.providerElasticIP.Attach()
}

func (e *elasticIP) Detach() error {
	return e.providerElasticIP.Detach()
}

func (e *elasticIP) Destroy() error {
	return e.providerElasticIP.Destroy()
}
