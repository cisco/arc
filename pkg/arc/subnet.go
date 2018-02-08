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
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type subnet struct {
	*config.Subnet
	network        *network
	providerSubnet resource.ProviderSubnet
}

// newSubnet is the constructor for a subnet object. It returns a non-nil error upon failure.
func newSubnet(net *network, prov provider.DataCenter, cfg *config.Subnet) (*subnet, error) {
	log.Debug("Initializing Subnet '%s'", cfg.Name())

	s := &subnet{
		Subnet:  cfg,
		network: net,
	}

	var err error

	s.providerSubnet, err = prov.NewSubnet(net, cfg)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// Id provides the id of the provider specific subnet resource.
// This satisfies the resource.DynamicSubnet interface.
func (s *subnet) Id() string {
	if s.providerSubnet == nil {
		return ""
	}
	return s.providerSubnet.Id()
}

// State provides the state of the provider specific subnet resource.
// This satisfies the resource.DynamicSubnet interface.
func (s *subnet) State() string {
	if s.providerSubnet == nil {
		return ""
	}
	return s.providerSubnet.State()
}

// ProviderSubnet satisfies the resource.Subnet interface and provides access
// to provider's subnet implementation.
func (s *subnet) ProviderSubnet() resource.ProviderSubnet {
	return s.providerSubnet
}

// Network satisfies the resource.Subnet interface and provides access
// to subnet's parent.
func (s *subnet) Network() resource.Network {
	return s.network
}

// Route satisfies the embedded resource.Resource interface in resource.Subnet.
// Subnet handles load, create, destroy, and info requests by delegating them
// to the providerSubnet.
func (s *subnet) Route(req *route.Request) route.Response {
	log.Route(req, "Subnet %q", s.Name())

	if req.Top() != "" {
		panic("Internal error: Unknown resource " + req.Top())
	}

	switch req.Command() {
	case route.Load:
		if err := s.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Create, route.Destroy, route.Info:
		return s.providerSubnet.Route(req)
	}
	msg.Error("Internal Error: arc/subnet.go. Unknown command %s", req.Command())
	return route.FAIL
}

func (s *subnet) Load() error {
	return s.providerSubnet.Load()
}

// Created satisfies the embedded resource.Resource interface in resource.Subnet.
// It delegates the call to the provider's subnet.
func (s *subnet) Created() bool {
	return s.providerSubnet.Created()
}

// Destroyed satisfies the embedded resource.Resource interface in resource.Subnet.
// It delegates the call to the provider's subnet.
func (s *subnet) Destroyed() bool {
	return s.providerSubnet.Destroyed()
}

func (s *subnet) Audit(flags ...string) error {
	return s.providerSubnet.Audit(flags...)
}
