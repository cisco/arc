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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/cisco/arc/pkg/log"
)

type dbSubnetGroupCacheEntry struct {
	deployed   *rds.DBSubnetGroup
	configured *dbSubnetGroup
}

type dbSubnetGroupCache struct {
	rds     *rds.RDS
	cache   map[string]dbSubnetGroupCacheEntry
	unnamed []*rds.DBSubnetGroup
}

func newDBSubnetGroupCache(rds *rds.RDS) *dbSubnetGroupCache {
	log.Debug("Initializaing AWS DB Subnet Group Cache")
	return &dbSubnetGroupCache{
		rds:   rds,
		cache: map[string]dbSubnetGroupCacheEntry{},
	}
}

func (c *dbSubnetGroupCache) load() error {
	log.Debug("Loading AWS DB Subnet Group Cache")

	var marker *string
	done := false

	for !done {
		params := &rds.DescribeDBSubnetGroupsInput{
			Marker:     marker,
			MaxRecords: aws.Int64(100),
		}
		resp, err := c.rds.DescribeDBSubnetGroups(params)
		if err != nil {
			return err
		}

		log.Debug("Load DBSubnetGroups: %d", len(resp.DBSubnetGroups))
		for _, sg := range resp.DBSubnetGroups {
			if sg == nil {
				continue
			}
			if sg.DBSubnetGroupName == nil {
				log.Verbose("\t\t%+v", *sg)
				c.unnamed = append(c.unnamed, sg)
			}

			name := *sg.DBSubnetGroupName
			log.Debug("Caching database subnet group %s", name)
			c.cache[name] = dbSubnetGroupCacheEntry{deployed: sg}
		}
		if resp.Marker != nil {
			marker = resp.Marker
		} else {
			done = true
		}
	}
	return nil
}

func (c *dbSubnetGroupCache) find(sg *dbSubnetGroup) *rds.DBSubnetGroup {
	e, ok := c.cache[sg.name()]
	if !ok {
		return nil
	}
	e.configured = sg
	return e.deployed
}

func (c *dbSubnetGroupCache) add(sg *dbSubnetGroup) {
	log.Debug("Adding %s to database subnet group cache.", sg.name())
	c.cache[sg.name()] = dbSubnetGroupCacheEntry{deployed: sg.sg, configured: sg}
}

func (c *dbSubnetGroupCache) remove(sg *dbSubnetGroup) {
	log.Debug("Removing %s from database subnet group cache", sg.name())
	delete(c.cache, sg.name())
}

func (c *dbSubnetGroupCache) audit(flags ...string) error {
	// TODO
	return nil
}
