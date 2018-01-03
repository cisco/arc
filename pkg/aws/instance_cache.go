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

type instanceCacheEntry struct {
	deployed   *ec2.Instance
	configured *instance
}

type instanceCache struct {
	cache   map[string]*instanceCacheEntry
	unnamed []*ec2.Instance
}

func newInstanceCache(c *compute) (*instanceCache, error) {
	log.Debug("Initializing AWS Instance Cache")

	i := &instanceCache{
		cache: map[string]*instanceCacheEntry{},
	}

	var token *string
	done := false

	for !done {
		params := &ec2.DescribeInstancesInput{
			MaxResults: aws.Int64(1000),
			NextToken:  token,
		}
		resp, err := c.ec2.DescribeInstances(params)
		if err != nil {
			return nil, err
		}

		log.Debug("Load Reservations: %d", len(resp.Reservations))
		for _, res := range resp.Reservations {
			if res == nil {
				log.Verbose("Skipping reservation %+v", res)
				continue
			}

			for _, inst := range res.Instances {
				if i == nil {
					log.Verbose("Skipping instance %+v", inst)
					continue
				}

				if inst.State != nil && inst.State.Name != nil && *inst.State.Name != "running" {
					log.Verbose("Non-running instance")
					if inst.InstanceId != nil {
						log.Verbose("\t\t%s", *inst.InstanceId)
					}
					continue
				}

				if inst.Tags == nil {
					log.Verbose("Unnamed instance")
					if inst.InstanceId != nil {
						log.Verbose("\t\t%s", *inst.InstanceId)
					}
					i.unnamed = append(i.unnamed, inst)
					continue
				}

				// Get the name and datacenter from the tags.
				name := ""
				dc := ""
				for _, t := range inst.Tags {
					if t.Key != nil && *t.Key == "Name" && t.Value != nil {
						name = *t.Value
					}
					if t.Key != nil && *t.Key == "DataCenter" && t.Value != nil {
						dc = *t.Value
					}
				}

				if name == "" {
					log.Verbose("Unnamed instance")
					if inst.InstanceId != nil {
						log.Verbose("\t\t%s", *inst.InstanceId)
					}
					continue
				}

				if dc == "" {
					log.Verbose("Instance %s not associated with a datacenter", name)
					if inst.InstanceId != nil {
						log.Verbose("\t\t%s", *inst.InstanceId)
					}
					continue
				}

				if dc != c.Compute.Name() {
					continue
				}

				log.Debug("Caching instance %s", name)
				i.cache[name] = &instanceCacheEntry{deployed: inst}
			}
		}

		if resp.NextToken != nil {
			token = resp.NextToken
		} else {
			done = true
		}
	}

	return i, nil
}

func (c *instanceCache) find(i *instance) *ec2.Instance {
	e := c.cache[i.Name()]
	if e == nil {
		return nil
	}
	e.configured = i
	return e.deployed
}

func (c *instanceCache) remove(i *instance) {
	log.Debug("Deleting %s from instanceCache", i.Name())
	delete(c.cache, i.Name())
}

func (c *instanceCache) audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("Name of audit object not given")
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit object doesn't exist")
	}
	for k, v := range c.cache {
		skip := false
		for _, tag := range v.deployed.Tags {
			if tag != nil && *tag.Key == "Name" && strings.HasPrefix(*tag.Value, "hedge-") {
				log.Debug("Audit ignoring: %s", *tag.Value)
				skip = true
			}
		}
		if !skip && v.configured == nil {
			a.Audit(aaa.Deployed, "%s", k)
		}
	}
	if c.unnamed != nil {
		a.Audit(aaa.Deployed, "\r")
		for i, v := range c.unnamed {
			u := "\t" + strings.Replace(fmt.Sprintf("%+v", v), "\n", "\n\t", -1)
			m := fmt.Sprintf("Unnamed Instance %d - InstanceId: %s", i+1, *v.InstanceId, u)
			a.Audit(aaa.Deployed, m)
		}
	}
	return nil
}
