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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
)

type dataCenterProvider struct {
	ec2    *ec2.EC2
	name   string
	number string
	region string
	images map[string]string
}

func NewDataCenterProvider(cfg *config.DataCenter) (provider.DataCenter, error) {
	log.Debug("Initializing AWS Datacenter Provider")

	name := cfg.Provider.Data["account"]
	if name == "" {
		return nil, fmt.Errorf("AWS DataCenter provider/data config requires an 'account' field, being the aws account name.")
	}
	region := cfg.Provider.Data["region"]
	if region == "" {
		return nil, fmt.Errorf("AWS DataCenter provider/data config requires a 'region' field, being the aws region.")
	}
	number := cfg.Provider.Data["number"]
	if number == "" {
		return nil, fmt.Errorf("AWS DataCenter provider/data config requires a 'number' field, being the aws account number.")
	}

	opts := session.Options{
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
			Region: aws.String(region),
		},
		Profile:           name,
		SharedConfigState: session.SharedConfigEnable,
	}

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return nil, err
	}

	return &dataCenterProvider{
		ec2:    ec2.New(sess),
		name:   name,
		number: number,
		region: region,
		images: cfg.Provider.Images,
	}, nil
}

func (p *dataCenterProvider) NewNetwork(cfg *config.Network) (resource.ProviderNetwork, error) {
	return newNetwork(cfg, p.ec2)
}

func (p *dataCenterProvider) NewSubnet(net resource.Network, cfg *config.Subnet) (resource.ProviderSubnet, error) {
	return newSubnet(net, cfg, p.ec2)
}

func (p *dataCenterProvider) NewSecurityGroup(net resource.Network, cfg *config.SecurityGroup) (resource.ProviderSecurityGroup, error) {
	return newSecurityGroup(net, cfg, p.ec2)
}

func (p *dataCenterProvider) NewNetworkPost(net resource.Network, cfg *config.Network) (resource.ProviderNetworkPost, error) {
	return newNetworkPost(net, cfg, p.ec2)
}

func (p *dataCenterProvider) NewCompute(cfg *config.Compute) (resource.ProviderCompute, error) {
	return newCompute(cfg, p.ec2)
}

func (p *dataCenterProvider) NewKeyPair(cfg *config.KeyPair) (resource.ProviderKeyPair, error) {
	return newKeyPair(cfg, p.ec2)
}

func (p *dataCenterProvider) NewInstance(i resource.Instance, cfg *config.Instance) (resource.ProviderInstance, error) {
	return newInstance(i, cfg, p)
}

func (p *dataCenterProvider) NewVolume(c resource.Compute, cfg *config.Volume) (resource.ProviderVolume, error) {
	return newVolume(c, cfg, p)
}

func (p *dataCenterProvider) NewElasticIP(e resource.ElasticIP, i resource.Instance) (resource.ProviderElasticIP, error) {
	return newElasticIP(i, p)
}

func (p *dataCenterProvider) NewRoleIdentifier(r resource.RoleIdentifier, name string, i resource.Instance) (resource.ProviderRoleIdentifier, error) {
	return newRoleIdentifier(name, p, i)
}

func (p *dataCenterProvider) findById(s string) string {
	for k, v := range p.images {
		if v == s {
			return k
		}
	}
	return ""
}
