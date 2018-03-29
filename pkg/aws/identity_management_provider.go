//
// Copyright (c) 2018, Cisco Systems
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
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
)

type identityManagementProvider struct {
	name   string
	number string
	iam    *iam.IAM
}

func newIdentityManagementProvider(cfg *config.Amp) (provider.IdentityManagement, error) {
	log.Debug("Initializing AWS IdentityManagement Provider")

	name := cfg.Provider.Data["account"]
	if name == "" {
		return nil, fmt.Errorf("AWS Amp provider/data config requires an 'account' field, being the aws account name.")
	}
	number := cfg.Provider.Data["number"]
	if number == "" {
		return nil, fmt.Errorf("AWS Amp provider/data config requires a 'number' field, being the aws account number.")
	}
	region := cfg.IdentityManagement.Region()
	if region == "" {
		return nil, fmt.Errorf("AWS Amp identityManagement config requires a 'region' field, being the region for identityManagement to exist.")
	}

	p := &identityManagementProvider{
		name:   name,
		number: number,
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
	p.iam = iam.New(sess)

	return p, nil
}

func (p *identityManagementProvider) NewIdentityManagement(cfg *config.IdentityManagement) (resource.ProviderIdentityManagement, error) {
	return newIdentityManagement(cfg, p.iam)
}

func (p *identityManagementProvider) NewRole(rl resource.Role, cfg *config.Role) (resource.ProviderRole, error) {
	return newRole(rl, cfg, p)
}

func (p *identityManagementProvider) NewPolicy(pol resource.Policy, cfg *config.Policy) (resource.ProviderPolicy, error) {
	return newPolicy(pol, cfg, p)
}

func init() {
	provider.RegisterIdentityManagement("aws", newIdentityManagementProvider)
}
