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

type databaseService struct {
	*resource.Resources
	*config.DatabaseService
	arc             *arc
	dc              resource.DataCenter
	databaseService resource.ProviderDatabaseService
	databases       *databases
}

// newDatabaseService is the constructor for a database service object. It returns a non-nil error upon failure.
func newDatabaseService(arc *arc, cfg *config.DatabaseService) (*databaseService, error) {
	if cfg == nil {
		return nil, nil
	}
	log.Debug("Initializing Database")

	// Validate the config.DatabaseService object.
	if cfg.Provider == nil {
		return nil, fmt.Errorf("The provider element is missing from the database_service configuration")
	}

	dbs := &databaseService{
		Resources:       resource.NewResources(),
		DatabaseService: cfg,
		arc:             arc,
	}

	p, err := provider.NewDatabaseService(cfg)
	if err != nil {
		return nil, err
	}

	db.databaseService, err = p.NewDatabaseService(cfg)
	if err != nil {
		return nil, err
	}
	n.Append(db.databaseService)

	db.databases, err = newDatabases(cfg.Databases, db.databaseService, p)
	if err != nil {
		return nil, err
	}
	n.Append(db.databases)

	return db, nil
}

// Arc satisfies the resource.DatabaseService interface and provides access
// to database's parent.
func (dbs *databaseService) Arc() resource.Arc {
	return dbs.arc
}

// Databases satisfies the resource.DatabaseService interface and provides access
// the the database service's children.
func (dbs *databaseService) Databases() resource.Databases {
	return dbs.databases
}

func (dbs *databaseService) Associate(dc resource.DataCenter) {
	dbs.dc = dc
}

func (dbs *databaseService) DataCenter() resource.DataCenter {
	return dbs.dc
}

// Route satisfies the embedded resource.Resource interface in resource.DatabaseService.
func (dbs *databaseService) Route(req *route.Request) route.Response {
	log.Route(req, "DatabaseService")

	// Route to the appropriate resource
	switch req.Top() {
	case "":
		break
	case "network":
		if d.Network() == nil {
			msg.Error("Network not defined in the config file")
			return route.OK
		}
		return d.Network().Route(req.Pop())
	case "subnet", "secgroup":
		if d.Network() == nil {
			msg.Error("Network not defined in the config file")
			return route.OK
		}
		return d.Network().Route(req)
	case "compute":
		if d.Compute() == nil {
			msg.Error("Compute not defined in the config file")
			return route.OK
		}
		return d.Compute().Route(req.Pop())
	case "keypair", "cluster", "pod", "instance", "volume", "eip":
		if d.Compute() == nil {
			msg.Error("Compute not defined in the config file")
			return route.OK
		}
		return d.Compute().Route(req)
	default:
		panic("Internal Error: Unknown path " + req.Top())
	}

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Load:
		return d.RouteInOrder(req)
	case route.Info:
		d.info(req)
		return route.OK
	case route.Audit:
		return d.RouteInOrder(req)
	default:
		panic("Internal Error: Unknown command " + req.Command().String())
	}
	return route.FAIL
}

func (dbs *databaseService) config() {
	dbs.DatabaseService.Print()
}

func (dbs *databaseService) info(req *route.Request) {
	if dbs.Destroyed() {
		return
	}
	msg.Info("Database Service")
	msg.IndentInc()
	dbs.RouteInOrder(req)
	msg.IndentDec()
}

func (dbs *databaseService) Audit(flags ...string) error {
	err := aaa.NewAudit("Database")
	if err != nil {
		return err
	}
	if err := dbs.databaseService.Audit("Database"); err != nil {
		return err
	}
	for _, d := range dbs.databases {
		if err := d.Audit("Database"); err != nil {
			return err
		}
	}
	return nil
}
