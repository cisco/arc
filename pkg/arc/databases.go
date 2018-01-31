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

type databases struct {
	*resource.Resources
	*config.Databases
	databases map[string]resource.Database
}

// newDatabases is the constructor for a database object. It returns a non-nil error upon failure.
func newDatabases(cfg *config.Databases, dbs resource.DatabaseService, prov provider.DatabaseService) (*databases, error) {
	log.Debug("Initializing Databases")

	d := &databases{
		Resources: resource.NewResources(),
		Databases: cfg,
		database:  map[string]resource.Database{},
	}

	for _, conf := range *cfg {
		if d.Find(conf.Name()) != nil {
			return nil, fmt.Errorf("Database name %q must be unique but is used multiple times", conf.Name())
		}
		database, err := newDatabase(conf, dbs, prov)
		if err != nil {
			return nil, err
		}
		d.databases[conf.Name()] = database
		d.Append(database)
	}
	return d, nil
}

// Find satisfies the resource.Database interface and provides a way
// to search for a specific database. This assumes databse names are unique.
func (d *databases) Find(name string) resource.Database {
	return d.databases[name]
}

// Route satisfies the embedded resource.Resource interface in resource.Databased.
// Databases can terminate load, create, destroy, help, config and info requests
// in order to manage all databases. All other commands are routed to a named database.
func (d *databases) Route(req *route.Request) route.Response {
	log.Route(req, "Databases")

	db := d.Find(req.Top())
	if db != nil {
		return db.Route(req.Pop())
	}
	if req.Top() != "" {
		msg.Error("Unknown database %q.", req.Top())
		return route.FAIL
	}

	// Skip if the test flag is set.
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load, route.Create:
		return d.RouteInOrder(req)
	case route.Destroy:
		return d.RouteReverseOrder(req)
	case route.Help:
		d.help()
		return route.OK
	case route.Config:
		d.config()
		return route.OK
	case route.Info:
		d.info(req)
		return route.OK
	default:
		msg.Error("Unknown database command %q.", req.Command().String())
	}
	return route.FAIL
}

func (d *databases) help() {
	commands := []help.Command{
		{route.Create.String(), "create all databases"},
		{route.Destroy.String(), "destroy all databases"},
		{"'name'", "manage named database"},
		{route.Config.String(), "show the databases configuration"},
		{route.Info.String(), "show information about allocated databases"},
		{route.Help.String(), "show this help"},
	}
	help.Print("db", commands)
}

func (d *databases) config() {
	d.Databases.Print()
}

func (d *databases) info(req *route.Request) {
	if d.Destroyed() {
		return
	}
	msg.Info("Databases")
	msg.IndentInc()
	d.RouteInOrder(req)
	msg.IndentDec()
}

func (d *databases) Audit(flags ...string) error {
	err := aaa.NewAudit("Database")
	if err != nil {
		return err
	}
	for _, db := range d.databases {
		if err := db.Audit("Database"); err != nil {
			return err
		}
	}
	return nil
}
