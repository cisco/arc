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
	log.Debug("BP1")
	v := s.s3[region]
	for key, val := range s.s3 {
		log.Debug("key = %q | val = %+v", key, val)
	}
	log.Debug("BP2")
	resp, err := v.ListBuckets(params)
	log.Debug("BP3")
	if err != nil {
		return nil, err
	}
	log.Debug("BP4")

	for _, r := range resp.Buckets {
		if r.Name == nil {
			log.Verbose("Unnamed bucket")
			c.unnamed = append(c.unnamed, r)
			continue
		}
		log.Debug("Caching %s", aws.StringValue(r.Name))
		c.cache[aws.StringValue(r.Name)] = &bucketCacheEntry{deployed: r}
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

func (c *bucketCache) remove(d *dnsRecord) {
	log.Debug("Deleting %s from dnsCache", d.Id())
	delete(c.cache, d.Id())
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
