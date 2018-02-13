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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/log"
)

type bucketCacheEntry struct {
	deployed   *s3.Bucket
	configured *bucket
}

type bucketCache struct {
	cache   map[string]*bucketCacheEntry
	unnamed []*s3.Bucket
}

func newBucketCache(s *storage) (*bucketCache, error) {
	log.Debug("Initializing AWS Bucket Cache")

	c := &bucketCache{
		cache: map[string]*bucketCacheEntry{},
	}

	params := &s3.ListBucketsInput{}

	// us-east-1 is being used here because it is the default region but plays no role in
	// listing the buckets, it is dependent on the account that created the s.s3 object.
	region := "us-east-1"
	v := s.s3[region]
	resp, err := v.ListBuckets(params)
	if err != nil {
		return nil, err
	}

	for _, b := range resp.Buckets {
		if b.Name == nil {
			log.Verbose("Unnamed bucket")
			c.unnamed = append(c.unnamed, b)
			continue
		}
		log.Debug("Caching %s", aws.StringValue(b.Name))
		c.cache[aws.StringValue(b.Name)] = &bucketCacheEntry{deployed: b}
	}

	return c, nil
}

func (c *bucketCache) find(b *bucket) *s3.Bucket {
	e := c.cache[b.Name()]
	if e == nil {
		return nil
	}
	e.configured = b
	return e.deployed
}

func (c *bucketCache) remove(b *bucket) {
	log.Debug("Deleting %s from bucketCache", b.Name())
	delete(c.cache, b.Name())
}

func (c *bucketCache) audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit Object does not exist")
	}
	for k, v := range c.cache {
		if v.configured == nil {
			a.Audit(aaa.Deployed, "%s", k)
		}
	}
	if c.unnamed != nil {
		a.Audit(aaa.Deployed, "\r")
		for i, v := range c.unnamed {
			u := "\t" + strings.Replace(fmt.Sprintf("%+v", v), "\n", "\n\t", -1)
			m := fmt.Sprintf("Unnamed Bucket %d - Bucket Name: %q %s", i+1, *v.Name, u)
			a.Audit(aaa.Deployed, m)
		}
	}
	return nil
}
