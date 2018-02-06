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
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

type databaseService struct {
	databaseCache      *databaseCache
	dbSubnetGroupCache *dbSubnetGroupCache
}

func newDatabaseService(cfg *config.DatabaseService, p *databaseServiceProvider) (resource.ProviderDatabaseService, error) {
	return &databaseService{
		databaseCache:      newDatabaseCache(p.rds),
		dbSubnetGroupCache: newDBSubnetGroupCache(p.rds),
	}, nil
}

func (dbs *databaseService) Load() error {
	if err := dbs.databaseCache.load(); err != nil {
		return err
	}
	if err := dbs.dbSubnetGroupCache.load(); err != nil {
		return err
	}
	return nil
}

func (dbs *databaseService) Audit(flags ...string) error {
	if err := dbs.databaseCache.audit(flags...); err != nil {
		return err
	}
	if err := dbs.dbSubnetGroupCache.audit(flags...); err != nil {
		return err
	}
	return nil
}

func (dbs *databaseService) Info() {
	msg.Detail("%-20s\t%d", "database cache size", len(dbs.databaseCache.cache))
	msg.Detail("%-20s\t%d", "dbSubnetGroup cache size", len(dbs.dbSubnetGroupCache.cache))
}
