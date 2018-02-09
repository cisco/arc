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

package aws

import (
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type routeTables struct {
	*resource.Resources
	routeTables map[string]*routeTable
}

func newRouteTables(c *ec2.EC2, n *network) (*routeTables, error) {
	log.Debug("Initializing AWS RouteTables")

	r := &routeTables{
		Resources:   resource.NewResources(),
		routeTables: map[string]*routeTable{},
	}

	routeTable, err := newRouteTable(c, n, "public", "public", "")
	if err != nil {
		return nil, err
	}
	r.routeTables["public"] = routeTable
	r.Append(routeTable)

	for _, availabilityZone := range n.AvailabilityZones() {
		name := "private-" + availabilityZone
		routeTable, err := newRouteTable(c, n, name, "private", availabilityZone)
		if err != nil {
			return nil, err
		}
		r.routeTables[name] = routeTable
		r.Append(routeTable)
	}

	// Create routetables for subnets that have "manage_route" set to true.
	for _, subnetGroup := range *n.SubnetGroups {
		if subnetGroup.ManageRoutes() == true {
			switch subnetGroup.Access() {
			case "public", "public_elastic", "local":
				routeTable, err := newRouteTable(c, n, subnetGroup.Name(), subnetGroup.Access(), "")
				if err != nil {
					return nil, err
				}
				r.routeTables[subnetGroup.Name()] = routeTable
				r.Append(routeTable)
			case "private":
				for _, availabilityZone := range n.AvailabilityZones() {
					name := subnetGroup.Name() + "-" + availabilityZone
					routeTable, err := newRouteTable(c, n, name, "private", availabilityZone)
					if err != nil {
						return nil, err
					}
					r.routeTables[name] = routeTable
					r.Append(routeTable)
				}
			}
		}
	}
	return r, nil
}

func (r *routeTables) Route(req *route.Request) route.Response {
	log.Route(req, "AWS RouteTables")

	if routeTable := r.find(req.Top()); routeTable != nil {
		return routeTable.Route(req.Pop())
	}
	if req.Top() != "" {
		msg.Error("Unknown routetable %q.", req.Top())
		return route.FAIL
	}

	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load, route.Create:
		return r.RouteInOrder(req)
	case route.Audit:
		return route.OK
	case route.Destroy:
		return r.RouteReverseOrder(req)
	case route.Help:
		r.help()
		return route.OK
	case route.Info:
		r.info(req)
		return route.OK
	}
	return route.FAIL
}

func (r *routeTables) find(s string) *routeTable {
	return r.routeTables[s]
}

func (r *routeTables) help() {
	commands := []help.Command{
		{Name: "'name'", Desc: "manage named routetable"},
		{Name: route.Create.String(), Desc: "create all routetables"},
		{Name: route.Destroy.String(), Desc: "destroy all routetables"},
		{Name: route.Info.String(), Desc: "show information about all allocated routetables"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("network routetable", commands)
}

func (r *routeTables) info(req *route.Request) {
	if r.Destroyed() {
		return
	}
	msg.Info("RouteTables")
	msg.IndentInc()
	r.RouteInOrder(req)
	msg.IndentDec()
}
