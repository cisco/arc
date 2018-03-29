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

type roleCacheEntry struct {
	deployed   *iam.Role
	configured *role
}

type roleCache struct {
	cache   map[string]*roleCacheEntry
	unnamed []*iam.Role
}

func newRoleCache(i *identityManagement) (*roleCache, error) {
	log.Debug("Initializing AWS Role Cache")

	c := &roleCache{
		cache: map[string]*roleCacheEntry{},
	}

	next := ""
	for {
		params := &iam.ListRolesInput{}

		if next != "" {
			params.Marker = aws.String(next)
		}

		resp, err := i.iam.ListRoles(params)
		if err != nil {
			return nil, err
		}
		truncated := false
		if resp.IsTruncated != nil {
			truncated = *resp.IsTruncated
		}
		next = ""
		if resp.Marker != nil {
			next = *resp.Marker
		}

		for _, role := range resp.Roles {
			if role.RoleName == nil {
				log.Verbose("Unnamed role")
				c.unnamed = append(c.unnamed, role)
				continue
			}
			log.Debug("Caching %s", aws.StringValue(role.RoleName))
			c.cache[aws.StringValue(role.RoleName)] = &roleCacheEntry{deployed: role}
		}
		if truncated == false {
			break
		}
		for _, r := range c.cache {
			params := &iam.GetRoleInput{
				RoleName: r.deployed.RoleName,
			}
			resp, err := i.iam.GetRole(params)
			if err != nil {
				return nil, err
			}
			r.deployed = resp.Role
		}
	}

	return c, nil
}

func (c *roleCache) find(r *role) *iam.Role {
	e := c.cache[r.Name()]
	if e == nil {
		return nil
	}
	e.configured = r
	return e.deployed
}

func (c *roleCache) remove(r *role) {
	log.Debug("Deleting %s from roleCache", r.Name())
	delete(c.cache, r.Name())
}

func (c *roleCache) audit(flags ...string) error {
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
			m := fmt.Sprintf("Unnamed Policy %d - Policy ID: %q %s", i+1, *v.RoleId, u)
			a.Audit(aaa.Deployed, m)
		}
	}
	return nil
}
