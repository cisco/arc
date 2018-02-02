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

package mock

import (
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

type database struct {
	*config.Database
	opt options
}

func newDatabase(cfg *config.Database, dbs resource.ProviderDatabaseService, p *databaseServiceProvider) (resource.ProviderDatabase, error) {
	log.Info("Initializing Mock Database %q", cfg.Name())
	db := &database{
		Database: cfg,
		opt:      options{p.Provider.Data},
	}
	if db.opt.err("db.New") {
		return nil, dberr{"db.New"}
	}
	return db, nil
}

func (db *database) Load() error {
	log.Info("Loading Mock Database")
	if db.opt.err("db.Load") {
		return dberr{"db.Load"}
	}
	return nil
}

func (db *database) Create(flags ...string) error {
	log.Info("Creating Mock Database")
	if db.opt.err("db.Create") {
		return dberr{"db.Create"}
	}
	return nil
}

func (db *database) Created() bool {
	if db.opt.err("db.Created") {
		return false
	}
	return true
}

func (db *database) Destroy(flags ...string) error {
	log.Info("Destroying Mock Database")
	if db.opt.err("db.Destroy") {
		return dberr{"db.Destroy"}
	}
	return nil
}

func (db *database) Provision(flags ...string) error {
	log.Info("Provisioning Mock Database")
	if db.opt.err("db.Provision") {
		return dberr{"db.Provision"}
	}
	return nil
}

func (db *database) Destroyed() bool {
	return false
}

func (db *database) Audit(flags ...string) error {
	log.Info("Auditing Mock Database")
	if db.opt.err("db.Audit") {
		return dberr{"db.Audit"}
	}
	return nil
}

func (db *database) Info() {
	msg.Info("Mock Database")
	msg.Detail("%-20s\t%s", "id", db.Id())
}

func (db *database) Id() string {
	return "db-mock-" + db.Name()
}
