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
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
)

type dataCenterProvider struct {
	*config.Provider
}

func NewDataCenterProvider(cfg *config.DataCenter) (provider.DataCenter, error) {
	log.Info("Initializing mock datacenter provider")

	return &dataCenterProvider{
		Provider: cfg.Provider,
	}, nil
}

func (p *dataCenterProvider) NewNetwork(cfg *config.Network) (resource.ProviderNetwork, error) {
	return newNetwork(cfg, p)
}

func (p *dataCenterProvider) NewSubnet(net resource.Network, cfg *config.Subnet) (resource.ProviderSubnet, error) {
	return newSubnet(cfg, p)
}

func (p *dataCenterProvider) NewSecurityGroup(net resource.Network, cfg *config.SecurityGroup) (resource.ProviderSecurityGroup, error) {
	return newSecurityGroup(cfg, p)
}

func (p *dataCenterProvider) NewNetworkPost(net resource.Network, cfg *config.Network) (resource.ProviderNetworkPost, error) {
	return newNetworkPost(cfg, p)
}

func (p *dataCenterProvider) NewCompute(cfg *config.Compute) (resource.ProviderCompute, error) {
	return newCompute(cfg, p)
}

func (p *dataCenterProvider) NewKeyPair(cfg *config.KeyPair) (resource.ProviderKeyPair, error) {
	return newKeyPair(cfg, p)
}

func (p *dataCenterProvider) NewInstance(instance resource.Instance, cfg *config.Instance) (resource.ProviderInstance, error) {
	return newInstance(cfg, p)
}

func (p *dataCenterProvider) NewVolume(c resource.Compute, cfg *config.Volume) (resource.ProviderVolume, error) {
	return newVolume(c, cfg, p)
}

func (p *dataCenterProvider) NewElasticIP(e resource.ElasticIP, i resource.Instance) (resource.ProviderElasticIP, error) {
	return newElasticIP(i, p)
}

func (p *dataCenterProvider) NewRole(r resource.Role, name string, in resource.Instance) (resource.ProviderRole, error) {
	return newRole(name, p, in)
}
