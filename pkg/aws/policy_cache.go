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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/log"
)

type policyCacheEntry struct {
	deployed   *iam.Policy
	configured *policy
}

type policyCache struct {
	cache   map[string]*policyCacheEntry
	unnamed []*iam.Policy
}

func newPolicyCache(i *identityManagement) (*policyCache, error) {
	log.Debug("Initializing AWS Policy Cache")

	c := &policyCache{
		cache: map[string]*policyCacheEntry{},
	}

	params := &iam.ListPoliciesInput{}

	resp, err := i.iam.ListPolicies(params)
	if err != nil {
		return nil, err
	}

	for _, policy := range resp.Policies {
		if policy.PolicyName == nil {
			log.Verbose("Unnamed bucket")
			c.unnamed = append(c.unnamed, policy)
			continue
		}
		log.Debug("Caching %s", aws.StringValue(policy.PolicyName))
		c.cache[aws.StringValue(policy.PolicyName)] = &policyCacheEntry{deployed: policy}
	}

	return c, nil
}

func (c *policyCache) find(p *policy) *iam.Policy {
	e := c.cache[p.Name()]
	if e == nil {
		return nil
	}
	e.configured = p
	return e.deployed
}

func (c *policyCache) remove(p *policy) {
	log.Debug("Deleting %s from policyCache", p.Name())
	delete(c.cache, p.Name())
}

func (c *policyCache) audit(flags ...string) error {
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
			m := fmt.Sprintf("Unnamed Policy %d - Policy ID: %q %s", i+1, *v.PolicyId, u)
			a.Audit(aaa.Deployed, m)
		}
	}
	return nil
}
