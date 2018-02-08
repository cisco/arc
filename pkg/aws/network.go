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

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

// network implements the resource.ProviderNetwork interface.
type network struct {
	*resource.Resources
	*config.Network
	ec2 *ec2.EC2

	vpc             *vpc
	routeTables     *routeTables
	internetGateway *internetGateway
	subnetCache     *subnetCache
	secgroupCache   *securityGroupCache
}

// newNetwork constructs the aws network.
func newNetwork(cfg *config.Network, c *ec2.EC2) (*network, error) {
	log.Debug("Initializing AWS Network")
	n := &network{
		Resources: resource.NewResources(),
		Network:   cfg,
		ec2:       c,
	}

	vpc, err := newVpc(c, n)
	if err != nil {
		return nil, err
	}
	n.vpc = vpc
	n.Append(vpc)

	routeTables, err := newRouteTables(c, n)
	if err != nil {
		return nil, err
	}
	n.routeTables = routeTables
	n.Append(routeTables)

	internetGateway, err := newInternetGateway(c, n, "public")
	if err != nil {
		return nil, err
	}
	n.internetGateway = internetGateway
	n.Append(internetGateway)

	// Load the vpc since it is needed for the caches.
	err = n.vpc.Load()
	if err != nil {
		return nil, err
	}

	n.subnetCache, err = newSubnetCache(n)
	if err != nil {
		return nil, err
	}
	n.secgroupCache, err = newSecurityGroupCache(n)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (n *network) Route(req *route.Request) route.Response {
	log.Route(req, "AWS Network")

	// Route to the appropriate resource.
	switch req.Top() {
	case "vpc":
		return n.vpc.Route(req.Pop())
	case "routetables", "routetable", "rt":
		return n.routeTables.Route(req.Pop())
	case "internetgateway", "igw":
		return n.internetGateway.Route(req.Pop())
	}

	// Handle commands
	switch req.Command() {
	case route.Load, route.Create, route.Info, route.Audit:
		return n.RouteInOrder(req)
	case route.Destroy:
		return n.RouteReverseOrder(req)
	}
	return route.FAIL
}

func (n *network) Id() string {
	if n.vpc == nil {
		return ""
	}
	return n.vpc.id()
}

func (n *network) State() string {
	if n.vpc == nil {
		return ""
	}
	return n.vpc.state()
}

func (n *network) AuditSubnets(flags ...string) error {
	return n.subnetCache.audit(flags...)
}

func (n *network) AuditSecgroups(flags ...string) error {
	return n.secgroupCache.audit(flags...)
}

func (n *network) CanRoute(req *route.Request) bool {
	switch req.Top() {
	case "vpc", "routetables", "routetable", "rt", "internetgateway", "igw":
		return true
	}
	return false
}

func (n *network) HelpCommands() []help.Command {
	return []help.Command{
		{Name: "vpc", Desc: "manage aws vpc"},
		{Name: "routetables", Desc: "manage aws routetables"},
		{Name: "routetable [name]", Desc: "manage named aws routetable"},
		{Name: "internetgateway", Desc: "manage aws internet gateway"},
	}
}
