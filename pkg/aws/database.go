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

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/resource"
)

type database struct {
	*config.Database
	dbs *databaseService
}

func newDatabase(cfg *config.Database, d resource.ProviderDatabaseService, p *databaseServiceProvider) (resource.ProviderDatabase, error) {
	dbs, ok := d.(*databaseService)
	if !ok {
		return nil, fmt.Errorf("Internal Error: aws/database.go, type assert for ProviderDatabaseService parameter failed.")
	}
	db := &database{
		Database: cfg,
		dbs:      dbs,
	}
	return db, nil
}

func (db *database) Load() error {
	return nil
}

func (db *database) Create(flags ...string) error {
	return nil
}

func (db *database) Created() bool {
	return false
}

func (db *database) Destroy(flags ...string) error {
	return nil
}

func (db *database) Provision(flags ...string) error {
	return nil
}

func (db *database) Destroyed() bool {
	return false
}

func (db *database) Audit(flags ...string) error {
	return nil
}

func (db *database) Info() {
}

func (db *database) Id() string {
	return ""
}
