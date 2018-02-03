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

type databaseCacheEntry struct {
	deployed   *rds.DBInstance
	configured *database
}

type databaseCache struct {
	rds   *rds.RDS
	cache map[string]*databaseCacheEntry
}

func newDatabaseCache(rds *rds.RDS) *databaseCache {
	log.Debug("Initializaing AWS Database Cache")
	return &databaseCache{
		rds:   rds,
		cache: map[string]*databaseCacheEntry{},
	}
}

func (c *databaseCache) load() error {
	log.Debug("Loading AWS Database Cache")

	var marker *string
	done := false

	for !done {
		params := &rds.DescribeDBInstancesInput{
			Marker:     marker,
			MaxRecords: aws.Int64(100),
		}
		resp, err := c.rds.DescribeDBInstances(params)
		if err != nil {
			return err
		}

		log.Debug("Load DBInstances: %d", len(resp.DBInstances))
		for _, db := range resp.DBInstances {
			if db == nil || db.DBInstanceIdentifier == nil {
				continue
			}
			id := *db.DBInstanceIdentifier
			log.Debug("Caching database instance %s", id)
			c.cache[id] = &databaseCacheEntry{deployed: db}

		}
		if resp.Marker != nil {
			marker = resp.Marker
		} else {
			done = true
		}
	}
	return nil
}

func (c *databaseCache) find(db *database) *rds.DBInstance {
	e := c.cache[db.Id()]
	if e == nil {
		return nil
	}
	e.configured = db
	return e.deployed
}

func (c *databaseCache) remove(db *database) {
	log.Debug("Deleting $s from database instance cache", db.Id())
	delete(c.cache, db.Id())
}

func (c *databaseCache) audit(flags ...string) error {
	// TODO
	return nil
}
