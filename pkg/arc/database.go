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

package arc

import (
	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	// "github.com/cisco/arc/pkg/help"
	// "github.com/cisco/arc/pkg/log"
	// "github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type database struct {
	*config.Database
	db resource.ProviderDatabase
}

func newDatabase(cfg *config.Database, dbs resource.ProviderDatabaseService, p provider.DatabaseService) (resource.Database, error) {
	return nil, nil
}

func (db *database) Route(req *route.Request) route.Response {
	return route.FAIL
}

func (db *database) Load() error {
	return db.db.Load()
}

func (db *database) Create(flags ...string) error {
	return db.db.Create(flags...)
}

func (db *database) Created() bool {
	return db.db.Created()
}

func (db *database) Destroy(flags ...string) error {
	return db.db.Destroy(flags...)
}

func (db *database) Destroyed() bool {
	return db.db.Destroyed()
}

func (db *database) Provision(flags ...string) error {
	return db.db.Provision(flags...)
}

func (db *database) Audit(flags ...string) error {
	err := aaa.NewAudit("Database")
	if err != nil {
		return err
	}
	return db.db.Audit("Database")
}

func (db *database) Info() {
	db.db.Info()
}

func (db *database) Id() string {
	return db.db.Id()
}
