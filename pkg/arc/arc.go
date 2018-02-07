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
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type arc struct {
	*resource.Resources
	*config.Arc
	datacenter      *dataCenter
	databaseService *databaseService
	dns             *dns
}

// New is the constructor for an arc object. It returns a non-nil error upon failure.
func New(cfg *config.Arc) (*arc, error) {
	log.Info("Initializing Arc: %q", cfg.Name())

	a := &arc{
		Resources: resource.NewResources(),
		Arc:       cfg,
	}
	a.header()

	var err error
	a.datacenter, err = newDataCenter(a, cfg.DataCenter)
	if err != nil {
		return nil, err
	}
	if a.datacenter != nil {
		a.Append(a.datacenter)
	}

	a.databaseService, err = newDatabaseService(cfg.DatabaseService, a)
	if err != nil {
		return nil, err
	}
	if a.databaseService != nil {
		a.Append(a.databaseService)
	}

	a.dns, err = newDns(a, cfg.Dns)
	if err != nil {
		return nil, err
	}
	if a.dns != nil {
		a.Append(a.dns)
	}

	// Associate datacenter service to dns service, and vice versa.
	if a.datacenter != nil && a.dns != nil {
		a.datacenter.associate(a.dns)
		a.dns.associate(a.datacenter)
	}

	return a, nil
}

// Run starts arc processing. It returns 0 for success, 1 for failure.
// Upon failure err might be set to a non-nil value.
func (a *arc) Run() (int, error) {
	u, err := user.Current()
	if err != nil {
		return 1, err
	}

	// Create base request.
	req := route.NewRequest(a.Name(), u.Username, time.Now().UTC().String())

	// Parse the request from the command line.
	req.Parse(os.Args[2:])
	log.Info("Creating %s request for user %q", req, u.Username)

	// Load the data from the provider unless there is a Load, Help or Config command.
	switch req.Command() {
	case route.None, route.Load:
		// Invalid commands: issue a help command.
		req.SetCommand(route.Help)
		a.Route(req)
		return 1, nil
	case route.Help, route.Config:
		// Skip loading for help and config commands since we aren't going to
		// interact with the provider.
		break
	default:
		if req.TestFlag() {
			break
		}
		log.Info("Loading arc: %q", a.Name())
		if resp := a.Route(req.Clone(route.Load)); resp != route.OK {
			return 1, fmt.Errorf("Failed to load datacenter %s", a.Name())
		}
		log.Info("Loading complete")
	}

	log.Info("Routing request: %q", req)
	resp := a.Route(req)
	if resp != route.OK {
		log.Info("Exiting, %s request failed\n", req)
		return 1, nil
	}
	log.Info("Exiting successfully\n")
	return 0, nil
}

// DataCenter satisfies the resource.Arc interface and provides access
// to arc's datacenter service object.
func (a *arc) DataCenter() resource.DataCenter {
	if a.datacenter == nil {
		return nil
	}
	return a.datacenter
}

// Database satisfies the resource.Arc interface and provides access to
// arc's database service object.
func (a *arc) DatabaseService() resource.DatabaseService {
	if a.databaseService == nil {
		return nil
	}
	return a.databaseService
}

// Dns satisfies the resource.Arc interface and provides access
// to arc's dns service object.
func (a *arc) Dns() resource.Dns {
	if a.dns == nil {
		return nil
	}
	return a.dns
}

// Route satisfies the embedded resource.Resource interface in resource.Arc.
// All request routing is done via Route. Arc will terminate the load, help,
// config and info requests. All other commands are routed to arc's children.
func (a *arc) Route(req *route.Request) route.Response {
	log.Route(req, "Arc")

	// Route to the appropriate resource
	switch req.Top() {
	case "":
		break
	case "network", "subnet", "secgroup", "compute", "keypair", "cluster", "pod", "instance", "volume", "eip":
		if a.datacenter == nil {
			msg.Error("Datacenter not defined in the config file")
			return route.FAIL
		}
		return a.DataCenter().Route(req)
	case "database", "db":
		if a.databaseService == nil {
			msg.Error("DatabaseService not defined in the config file")
			return route.FAIL
		}
		return a.databaseService.Route(req.Pop())
	case "dns":
		if a.dns == nil {
			msg.Error("Dns not defined in the config file")
			return route.FAIL
		}
		return a.Dns().Route(req.Pop())
	default:
		Help()
		return route.FAIL
	}

	// Skip if the test flag is set.
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Load:
		return a.RouteInOrder(req)
	case route.Help:
		Help()
		return route.OK
	case route.Config:
		a.config()
		return route.OK
	case route.Info:
		a.info(req)
		return route.OK
	case route.Audit:
		return a.RouteInOrder(req)
	default:
		msg.Error("Unknown arc command %q.", req.Command().String())
	}
	return route.FAIL
}

// Help provides the command line help for the arc command.
func Help() {
	commands := []help.Command{
		{"network", "manage network"},
		{"subnet", "manage subnet groups"},
		{"subnet 'name'", "manage named subnet group"},
		{"secgroup", "manage security groups"},
		{"secgroup 'name'", "manage named security group"},
		{"compute", "manage compute"},
		{"keypair", "manage keypair"},
		{"cluster 'name'", "manage named cluster"},
		{"pod 'name'", "manage named pod"},
		{"instance 'name'", "manage named instance"},
		{"db", "manage database service"},
		{"db 'name'", "manage named database service"},
		{"dns", "manage dns"},
		{route.Config.String(), "show the arc configuration for the given datacenter"},
		{route.Info.String(), "show information about allocated arc resources"},
		{route.Help.String(), "show this help"},
	}
	help.Print("", commands)
}

func (a *arc) config() {
	a.Arc.Print()
}

func (a *arc) info(req *route.Request) {
	if a.Destroyed() {
		return
	}
	msg.IndentInc()
	a.Arc.PrintLocal()
	a.RouteInOrder(req)
	msg.IndentDec()
}

func (a *arc) header() {
	msg.Heading("arc, %s", env.Lookup("VERSION"))
}
