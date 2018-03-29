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

type storageProvider struct {
	name   string
	number string
	s3     map[string]*s3.S3
}

func newStorageProvider(cfg *config.Amp) (provider.Storage, error) {
	log.Debug("Initializing AWS Storage")

	name := cfg.Provider.Data["account"]
	if name == "" {
		return nil, fmt.Errorf("AWS Storage provider/data config requires an 'account' field, being the aws account name.")
	}
	number := cfg.Provider.Data["number"]
	if number == "" {
		return nil, fmt.Errorf("AWS Storage provider/data config requires a 'number' field, being the aws account number.")
	}

	regions := map[string]string{}

	for _, bucket := range cfg.Storage.Buckets {
		if region := regions[bucket.Region()]; region == "" {
			log.Debug("Region %q", bucket.Region())
			regions[bucket.Region()] = cfg.Provider.Data["account"]
		}
	}

	p := &storageProvider{
		name:   name,
		number: number,
	}

	p.s3 = map[string]*s3.S3{}

	for region := range regions {
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

		p.s3[region] = s3.New(sess)
	}

	return p, nil
}

func (p *storageProvider) NewStorage(cfg *config.Storage) (resource.ProviderStorage, error) {
	return newStorage(cfg, p.s3)
}

func (p *storageProvider) NewBucket(b resource.Bucket, cfg *config.Bucket) (resource.ProviderBucket, error) {
	return newBucket(b, cfg, p)
}

func init() {
	provider.RegisterStorage("aws", newStorageProvider)
}
