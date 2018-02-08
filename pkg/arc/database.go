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
	"fmt"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type databaseParams struct {
}

type database struct {
	*config.Database
	databaseService  *databaseService
	providerDatabase resource.ProviderDatabase
}

func newDatabase(cfg *config.Database, dbs *databaseService, p provider.DatabaseService) (resource.Database, error) {
	if cfg.Name() == "" {
		return nil, fmt.Errorf("The 'database' element is missing from the database configuration")
	}
	if cfg.Engine() == "" {
		return nil, fmt.Errorf("The 'engine' element is missing from the database configuration")
	}
	if cfg.InstanceType() == "" {
		return nil, fmt.Errorf("The 'type' element is missing from the database configuration")
	}
	if cfg.SubnetGroup() == "" {
		return nil, fmt.Errorf("The 'subnet_group' element is missing from the database configuration")
	}
	if cfg.SecurityGroups() == nil {
		return nil, fmt.Errorf("The 'security_groups' element is missing from the database configuration")
	}

	log.Debug("Initializing Database %q", cfg.Name())

	db := &database{
		Database:        cfg,
		databaseService: dbs,
	}

	// Need to build up the database parameters. First we will gather the list of subnets
	// that the database instance will use.
	dc := db.databaseService.Arc().DataCenter()
	if dc == nil {
		return nil, fmt.Errorf("The database service requires the datacenter service to be defined in the configuration.")
	}
	net := dc.Network()
	if net == nil {
		return nil, fmt.Errorf("The database service requires the datacenter service to define a network in the configuration.")
	}
	subnetGroups := net.SubnetGroups()
	if subnetGroups == nil {
		return nil, fmt.Errorf("The database service requires the datacenter service to define subnet groups in the configuration.")
	}
	securityGroups := net.SecurityGroups()
	if securityGroups == nil {
		return nil, fmt.Errorf("The database service requires the datacenter service to define security groups in the configuration.")
	}

	subnets := []resource.ProviderSubnet{}
	subnetGroup := subnetGroups.Find(cfg.SubnetGroup())
	if subnetGroup == nil {
		return nil, fmt.Errorf("Creating database %s, unable to find subnet group %s.", db.Name(), cfg.SubnetGroup())
	}
	for _, s := range subnetGroup.Subnets() {
		subnets = append(subnets, s.ProviderSubnet())
	}

	// Need to build up the database parameters. Second we will gather the list of security groups
	// that the database instance will use.
	secgroups := []resource.ProviderSecurityGroup{}
	for _, secgroupName := range cfg.SecurityGroups() {
		secgroup := securityGroups.Find(secgroupName)
		if secgroup == nil {
			return nil, fmt.Errorf("Creating database %s, unable to find security group %s.", db.Name(), secgroupName)
		}
		secgroups = append(secgroups, secgroup.ProviderSecurityGroup())
	}

	params := resource.DatabaseParams{
		DatabaseService: db.databaseService.ProviderDatabaseService(),
		Subnets:         subnets,
		SecurityGroups:  secgroups,
	}

	var err error
	db.providerDatabase, err = p.NewDatabase(cfg, params)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Route satisfies the resource.Database interface.
func (db *database) Route(req *route.Request) route.Response {
	log.Route(req, "Database %q", db.Name())

	if req.Top() != "" {
		db.Help()
		return route.FAIL
	}

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Load:
		if err := db.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Create:
		if err := db.Create(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Destroy:
		if err := db.Destroy(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Provision:
		if err := db.Provision(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Audit:
		if err := db.Audit(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Info:
		db.Info()
	case route.Config:
		db.Print()
	case route.Help:
		db.Help()
	default:
		msg.Error("Internal Error: Unknown command " + req.Command().String())
		return route.FAIL
	}
	return route.OK
}

// Load satisfies the resource.Database interface.
func (db *database) Load() error {
	return db.providerDatabase.Load()
}

// Create satisfies the resource.Database interface.
func (db *database) Create(flags ...string) error {
	return db.providerDatabase.Create(flags...)
}

// Created satisfies the resource.Database interface.
func (db *database) Created() bool {
	return db.providerDatabase.Created()
}

// Destroy satisfies the resource.Database interface.
func (db *database) Destroy(flags ...string) error {
	return db.providerDatabase.Destroy(flags...)
}

// Destroyed satisfies the resource.Database interface.
func (db *database) Destroyed() bool {
	return db.providerDatabase.Destroyed()
}

// Provision satisfies the resource.Database interface.
func (db *database) Provision(flags ...string) error {
	return db.providerDatabase.Provision(flags...)
}

// Audit satisfies the resource.Database interface.
func (db *database) Audit(flags ...string) error {
	auditSession := "Database"
	found := false
	for _, v := range flags {
		if v == auditSession {
			found = true
			break
		}
	}
	if !found {
		flags = append(flags, auditSession)
	}
	err := aaa.NewAudit("Database")
	if err != nil {
		return err
	}
	return db.providerDatabase.Audit(flags...)
}

// Info satisfies the resource.Database interface.
func (db *database) Info() {
	if db.Destroyed() {
		return
	}
	msg.Info("Database Instance")
	db.providerDatabase.Info()
}

// Id satisfies the resource.Database interface.
func (db *database) Id() string {
	return db.providerDatabase.Id()
}

// Help satisfies the resource.Database interface.
func (db *database) Help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: fmt.Sprintf("create %s database instance", db.Name())},
		{Name: route.Destroy.String(), Desc: fmt.Sprintf("destroy %s database instance", db.Name())},
		{Name: route.Provision.String(), Desc: fmt.Sprintf("update %s database instance", db.Name())},
		{Name: route.Audit.String(), Desc: fmt.Sprintf("audit %s database instance", db.Name())},
		{Name: route.Info.String(), Desc: fmt.Sprintf("provide information about allocated %s database instance", db.Name())},
		{Name: route.Config.String(), Desc: fmt.Sprintf("provide the %s database instance configuration", db.Name())},
		{Name: route.Help.String(), Desc: "provide this help"},
	}
	help.Print(fmt.Sprintf("db %s", db.Name()), commands)
}

// DatabaseService satisfies the resource.Database interface.
func (db *database) DatabaseService() resource.DatabaseService {
	return db.databaseService
}

// ProviderDatabase allows access to the provider's database object.
func (db *database) ProviderDatabase() resource.ProviderDatabase {
	return db.providerDatabase
}
