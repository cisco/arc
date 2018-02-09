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

package arc

import (
	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type databaseService struct {
	*config.DatabaseService
	arc                     resource.Arc
	providerDatabaseService resource.ProviderDatabaseService
	databases               []resource.Database
}

// newDatabaseService is the constructor for a database service object. It returns a non-nil error upon failure.
func newDatabaseService(cfg *config.DatabaseService, arc resource.Arc) (*databaseService, error) {
	if cfg == nil {
		return nil, nil
	}
	log.Debug("Initializing Database Service")

	dbs := &databaseService{
		DatabaseService: cfg,
		arc:             arc,
	}

	p, err := provider.NewDatabaseService(cfg)
	if err != nil {
		return nil, err
	}

	dbs.providerDatabaseService, err = p.NewDatabaseService(cfg)
	if err != nil {
		return nil, err
	}

	for _, c := range cfg.Databases {
		db, err := newDatabase(c, dbs, p)
		if err != nil {
			return nil, err
		}
		dbs.databases = append(dbs.databases, db)
	}

	return dbs, nil
}

// Route satisfies the resource.DatabaseService interface.
func (dbs *databaseService) Route(req *route.Request) route.Response {
	log.Route(req, "DatabaseService")

	// Route to the appropriate resource
	switch req.Top() {
	case "":
		break
	default:
		db := dbs.Find(req.Top())
		if db == nil {
			msg.Error("Unknown database %q.", req.Top())
			return route.FAIL
		}
		return db.Route(req.Pop())
	}

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Load:
		if err := dbs.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Provision:
		if err := dbs.Provision(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Audit:
		if err := dbs.Audit(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Info:
		dbs.Info()
	case route.Config:
		dbs.Print()
	case route.Help:
		dbs.Help()
	default:
		dbs.Help()
		return route.FAIL
	}
	return route.OK
}

// Load satisfies the resource.DatabaseService interface.
func (dbs *databaseService) Load() error {
	log.Info("Loading Database Service")
	if err := dbs.providerDatabaseService.Load(); err != nil {
		return err
	}
	for _, db := range dbs.databases {
		if err := db.Load(); err != nil {
			return err
		}
	}
	return nil
}

// Provision satisfies the resource.DatabaseService interface.
func (dbs *databaseService) Provision(flags ...string) error {
	log.Info("Provisioning database service")
	for _, db := range dbs.databases {
		if err := db.Provision(flags...); err != nil {
			return err
		}
	}
	return nil
}

// Audit satisfies the resource.DatabaseService interface.
func (dbs *databaseService) Audit(flags ...string) error {
	auditSession := "Database"
	flags = append(flags, auditSession)

	err := aaa.NewAudit(auditSession)
	if err != nil {
		return err
	}
	if err := dbs.providerDatabaseService.Audit(flags...); err != nil {
		return err
	}
	for _, db := range dbs.databases {
		if err := db.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}

// Info satisfies the resource.DatabaseService interface.
func (dbs *databaseService) Info() {
	if dbs.Destroyed() {
		return
	}
	msg.Info("Database Service")
	dbs.providerDatabaseService.Info()
	msg.IndentInc()
	for _, db := range dbs.databases {
		db.Info()
	}
	msg.IndentDec()
}

// Help satisfies resource.DatabaseService.
func (dbs *databaseService) Help() {
	commands := []help.Command{
		{Name: "'name'", Desc: "manage named database instance"},
		{Name: route.Provision.String(), Desc: "update the database service"},
		{Name: route.Audit.String(), Desc: "audit the database service"},
		{Name: route.Info.String(), Desc: "show information about allocated database service"},
		{Name: route.Config.String(), Desc: "show the configuration for the given database service"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("db", commands)
}

// Arc satisfies the resource.DatabaseService interface and provides access to database's parent.
func (dbs *databaseService) Arc() resource.Arc {
	return dbs.arc
}

// Find satisfies the resource.DatabaseService interface. It returns the database with the given name.
func (dbs *databaseService) Find(name string) resource.Database {
	for _, db := range dbs.databases {
		if db.Name() == name {
			return db
		}
	}
	return nil
}

// ProviderDatabaseSerivces allows access to the provider's database service object.
func (dbs *databaseService) ProviderDatabaseService() resource.ProviderDatabaseService {
	return dbs.providerDatabaseService
}

// Created is required since the parent of this object, Arc, wants to treat it like a resource.Resource.
func (dbs *databaseService) Created() bool {
	// All database instances must be created to consider it created.
	for _, db := range dbs.databases {
		if !db.Created() {
			return false
		}
	}
	return true
}

// Destroyed is required since the parent of this object, Arc, wants to treat it like a resource.Resource.
func (dbs *databaseService) Destroyed() bool {
	// All database instances must be destroyed to consider it destroyed.
	for _, db := range dbs.databases {
		if !db.Destroyed() {
			return false
		}
	}
	return true
}
