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
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/log"
)

type subnetCacheEntry struct {
	deployed   *ec2.Subnet
	configured *subnet
}

type subnetCache struct {
	cache   map[string]*subnetCacheEntry
	unnamed []*ec2.Subnet
}

func newSubnetCache(n *network) (*subnetCache, error) {
	log.Debug("Initializing AWS Subnet Cache")

	c := &subnetCache{
		cache: map[string]*subnetCacheEntry{},
	}

	params := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(n.vpc.id()),
				},
			},
		},
	}
	resp, err := n.ec2.DescribeSubnets(params)
	if err != nil {
		return nil, err
	}
	log.Debug("Load Subnets: %d", len(resp.Subnets))

	for _, s := range resp.Subnets {

		if s.Tags == nil {
			log.Verbose("Unnamed subnet")
			if s.SubnetId != nil {
				log.Verbose("\t\t%s", *s.SubnetId)
			}
			c.unnamed = append(c.unnamed, s)
			continue
		}

		// Get the name from the tags.
		name := ""
		for _, t := range s.Tags {
			if t.Key != nil && *t.Key == "Name" && t.Value != nil {
				name = *t.Value
				break
			}
		}
		if name == "" {
			log.Verbose("Unamed subnet")
			if s.SubnetId != nil {
				log.Verbose("\t\t&s", &s.SubnetId)
			}
			c.unnamed = append(c.unnamed, s)
			continue
		}

		// Remember the subnet.
		log.Debug("Caching %s", name)
		c.cache[name] = &subnetCacheEntry{deployed: s}
	}
	return c, nil
}

func (c *subnetCache) find(s *subnet) *ec2.Subnet {
	e := c.cache[s.Name()]
	if e == nil {
		return nil
	}
	e.configured = s
	return e.deployed
}

func (c *subnetCache) findById(s string) *ec2.Subnet {
	for _, v := range c.cache {
		if *v.deployed.SubnetId == s {
			return v.deployed
		}
	}
	return nil
}

func (c *subnetCache) remove(s *subnet) {
	log.Debug("Deleting %s from subnetCache", s.Name())
	delete(c.cache, s.Name())
}

func (c *subnetCache) audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	for k, v := range c.cache {
		if v.configured == nil {
			a.Audit(aaa.Deployed, "%s", k)
		}
	}
	if c.unnamed != nil {
		a.Audit(aaa.Deployed, "\r")
		for i, v := range c.unnamed {
			u := "\t" + strings.Replace(fmt.Sprintf("%+v", v), "\n", "\n\t", -1)
			m := fmt.Sprintf("Unnamed Subnet %d - SubnetId: %s CidrBlock: %s\n%s", i+1, *v.SubnetId, *v.CidrBlock, u)
			a.Audit(aaa.Deployed, m)
		}
	}
	return nil
}
