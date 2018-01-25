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
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
)

type accountProvider struct {
	s3     map[string]*s3.S3
	name   string
	number string
}

func NewAccountProvider(cfg *config.Account) (provider.Account, error) {
	log.Debug("Initializing AWS Account")

	name := cfg.Provider.Data["account"]
	if name == "" {
		return nil, fmt.Errorf("AWS Account provider/data config requires an 'account' field, being the aws account name.")
	}
	number := cfg.Provider.Data["number"]
	if number == "" {
		return nil, fmt.Errorf("AWS Account provider/data config requires a 'number' field, being the aws account number.")
	}

	regions := map[string]string{}

	for _, bucket := range *cfg.Storage.Buckets {
		if region := regions[bucket.Region()]; region == "" {
			log.Debug("Region %q", region)
			regions[bucket.Region()] = cfg.Provider.Data["account"]
		}
	}

	a := &accountProvider{
		name:   name,
		number: number,
	}

	a.s3 = map[string]*s3.S3{}

	for region, _ := range regions {
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

		a.s3[region] = s3.New(sess)
	}

	return a, nil
}

func (a *accountProvider) NewStorage(cfg *config.Storage) (resource.ProviderStorage, error) {
	return newStorage(cfg, a.s3)
}

func (a *accountProvider) NewBucket(b resource.Bucket, cfg *config.Bucket) (resource.ProviderBucket, error) {
	return newBucket(b, cfg, a)
}
