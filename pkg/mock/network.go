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

// network implements the resource.ProviderNetwork interface.
type network struct {
	*mock
	*config.Network
	id    string
	state string
}

// newNetwork constructs the mock network.
func newNetwork(cfg *config.Network, p *dataCenterProvider) (resource.ProviderNetwork, error) {
	log.Info("Initializing mock network")
	n := &network{
		mock:    newMock("network", p.Provider),
		Network: cfg,
		id:      "0xdeadbeef",
		state:   "available",
	}
	return n, nil
}

func (n *network) Id() string {
	return n.id
}

func (n *network) State() string {
	return n.state
}

func (n *network) AuditSubnets(flags ...string) error {
	return nil
}

func (n *network) AuditSecgroups(flags ...string) error {
	return nil
}

// networkPost implements the resource.ProviderNetworkPost interface.
type networkPost struct {
	*mock
	*config.Network
}

// newNetworkPost constructs the mock network.
func newNetworkPost(cfg *config.Network, p *dataCenterProvider) (resource.ProviderNetworkPost, error) {
	log.Info("Initializing mock network post")
	n := &networkPost{
		mock:    newMock("networkPost", p.Provider),
		Network: cfg,
	}
	return n, nil
}
