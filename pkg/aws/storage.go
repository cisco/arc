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
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type storage struct {
	*config.Storage
	s3 map[string]*s3.S3

	bucketCache *bucketCache
	buckets     []*bucket
}

func NewStorage(cfg *config.Storage) (provider.Storage, error) {
	log.Debug("Initializing AWS Storage Provider")

	name := cfg.Provider.Data["account"]
	if name == "" {
		return nil, fmt.Errorf("AWS Storage provider/data config requires an 'account' field, being the aws account name.")
	}
	number := cfg.Provider.Data["number"]
	if number == "" {
		return nil, fmt.Errorf("AWS Storage provider/data config requires a 'number' field, being the aws account number.")
	}

	regions := map[string]string{}

	for _, bucket := range cfg.Buckets {
		if region := regions[bucket.Region()]; region == "" {
			regions[bucket.Region()] = cfg.Provider.Data["account"]
		}
	}

	s := &storage{
		Storage: cfg,
	}

	s.s3 = make(map[string]*s3.S3)

	msg.Info("Creating S3 objects for account %q", cfg.Provider.Data["account"])
	for region, _ := range regions {
		msg.Info("Creating S3 object %q", region)
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

		s.s3[region] = s3.New(sess)
	}
	var err error

	s.bucketCache, err = newBucketCache(s)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *storage) NewBucket(cfg *config.Bucket) (resource.ProviderBucket, error) {
	b, err := newBucket(s, cfg, s.s3[cfg.Region()])
	s.buckets = append(s.buckets, b.(*bucket))
	return b, err
}

func (s *storage) Route(req *route.Request) route.Response {
	log.Route(req, "AWS Storage")

	switch req.Command() {
	case route.Info:
		s.Info()
		return route.OK
	}
	return route.FAIL
}

func (s *storage) Info() {
	for _, v := range s.buckets {
		v.Info()
	}
}
