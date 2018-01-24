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

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	// "github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"

	"github.com/cisco/arc/pkg/aws"
	"github.com/cisco/arc/pkg/mock"
	//"github.com/cisco/arc/pkg/gcp"
	//"github.com/cisco/arc/pkg/azure"
)

type database struct {
	*resource.Resources
	*config.Database
	arc *arc
}

// newDatabase is the constructor for a database service object. It returns a non-nil error upon failure.
func newDatabase(arc *arc, cfg *config.Database) (*database, error) {
	if cfg == nil {
		return nil, nil
	}
	log.Debug("Initializing Database")

	// Validate the config.Database object.
	if cfg.Provider == nil {
		return nil, fmt.Errorf("The provider element is missing from the database configuration")
	}

	db := &database{
		Resources: resource.NewResources(),
		Database:  cfg,
		arc:       arc,
	}

	vendor := cfg.Provider.Vendor
	var err error
	// var p provider.Database

	switch vendor {
	case "mock":
		_, err = mock.NewDatabaseProvider(cfg)
	case "aws":
		_, err = aws.NewDatabaseProvider(cfg)
	//case "azure":
	//	p, err = azure.NewDatabaseProvider(cfg)
	//case "gcp":
	//	p, err = gcp.NewDatabaseProvider(cfg)
	default:
		err = fmt.Errorf("Unknown vendor %q", vendor)
	}
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Arc satisfies the resource.Database interface and provides access
// to database's parent.
func (db *database) Arc() resource.Arc {
	return db.arc
}

// Route satisfies the embedded resource.Resource interface in resource.Database.
// Database does not directly terminate a request so only handles load and info
// requests from it's parent.  All other commands are routed to arc's children.
func (db *database) Route(req *route.Request) route.Response {
	log.Route(req, "Database")

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Load:
		return db.RouteInOrder(req)
	default:
		panic("Internal Error: Unknown command " + req.Command().String())
	}
	return route.FAIL
}

func (db *database) info(req *route.Request) {
	if db.Destroyed() {
		return
	}
	msg.Info("Database")
	msg.IndentInc()
	db.RouteInOrder(req)
	msg.IndentDec()
}
