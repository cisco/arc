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
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
)

func init() {
	provider.RegisterContainerService("aws", newContainerServiceProvider)
}

type containerServiceProvider struct {
	ecs     *ecs.ECS
	account string
	region  string
}

func newContainerServiceProvider(cfg *config.ContainerService) (provider.ContainerService, error) {
	log.Debug("Initializing AWS Container Service Provider")

	account := cfg.Provider.Data["account"]
	if account == "" {
		return nil, fmt.Errorf("AWS ContainerService provider/data config requires an 'account' field, being the aws account name.")
	}
	region := cfg.Provider.Data["region"]
	if region == "" {
		return nil, fmt.Errorf("AWS ContainerService provider/data config requires a 'region' field, being the aws region.")
	}

	opts := session.Options{
		Config: aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
			Region: aws.String(region),
		},
		Profile:           account,
		SharedConfigState: session.SharedConfigEnable,
	}

	sess, err := session.NewSessionWithOptions(opts)
	if err != nil {
		return nil, err
	}

	return &containerServiceProvider{
		ecs:     ecs.New(sess),
		account: account,
		region:  region,
	}, nil
}

func (p *containerServiceProvider) NewContainerService(cfg *config.ContainerService) (resource.ProviderContainerService, error) {
	return newContainerService(cfg, p)
}
