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

type securityGroupCacheEntry struct {
	deployed   *ec2.SecurityGroup
	configured *securityGroup
}

type securityGroupCache struct {
	cache   map[string]*securityGroupCacheEntry
	unnamed []*ec2.SecurityGroup
}

func newSecurityGroupCache(n *network) (*securityGroupCache, error) {
	log.Debug("Initializing AWS SecurityGroup Cache")

	c := &securityGroupCache{
		cache: map[string]*securityGroupCacheEntry{},
	}

	params := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(n.vpc.id()),
				},
			},
		},
	}
	resp, err := n.ec2.DescribeSecurityGroups(params)
	if err != nil {
		return nil, err
	}
	log.Debug("Load Security Groups: %d", len(resp.SecurityGroups))

	for _, s := range resp.SecurityGroups {

		// We don't have a named security group, so cache in the unnamed slice.
		if s.Tags == nil {
			log.Verbose("Unamed Security Group")
			if s.GroupId != nil {
				log.Verbose("\t\t&s", &s.GroupId)
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
			log.Verbose("Unamed Security Group")
			if s.GroupId != nil {
				log.Verbose("\t\t&s", &s.GroupId)
			}
			c.unnamed = append(c.unnamed, s)
			continue
		}

		// Remember the security group.
		log.Debug("Caching %s", name)
		c.cache[name] = &securityGroupCacheEntry{deployed: s}
	}
	return c, nil
}

func (c *securityGroupCache) find(s *securityGroup) *ec2.SecurityGroup {
	e := c.cache[s.Name()]
	if e == nil {
		return nil
	}
	e.configured = s
	return e.deployed
}

func (c *securityGroupCache) remove(name string) {
	log.Debug("Deleting %s from securityGroupCache", name)
	delete(c.cache, name)
}

func (c *securityGroupCache) audit(flags ...string) error {
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
			m := fmt.Sprintf("Unnamed Security Group %d - GroupId: %s\n%s", i+1, *v.GroupId, u)
			a.Audit(aaa.Deployed, m)
		}
	}
	return nil
}
