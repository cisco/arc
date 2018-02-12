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

type natGateways struct {
	*resource.Resources
	natGateways map[string]*natGateway
}

func newNatGateways(c *ec2.EC2, n resource.Network) (*natGateways, error) {
	log.Debug("Initializing AWS NatGateways")

	natNeeded := false
	for _, s := range n.SubnetGroups().Get() {
		if s.Access() == "private" {
			natNeeded = true
			break
		}
	}
	if !natNeeded {
		log.Debug("AWS NatGateways not needed. No subnets with private access.")
		return nil, nil
	}

	ngws := &natGateways{
		Resources:   resource.NewResources(),
		natGateways: map[string]*natGateway{},
	}
	for _, availabilityZone := range n.AvailabilityZones() {
		name := "private-" + availabilityZone
		ngw, err := newNatGateway(c, n, name, availabilityZone)
		if err != nil {
			return nil, err
		}
		ngws.natGateways[name] = ngw
		ngws.Append(ngw)
	}
	return ngws, nil
}

func (n *natGateways) Route(req *route.Request) route.Response {
	log.Route(req, "AWS NatGateways")

	if natGateway := n.find(req.Top()); natGateway != nil {
		return natGateway.Route(req.Pop())
	}
	if req.Top() != "" {
		msg.Error("Unknown natgateway %q.", req.Top())
		return route.FAIL
	}

	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load, route.Create:
		return n.RouteInOrder(req)
	case route.Destroy:
		return n.RouteReverseOrder(req)
	case route.Help:
		n.help()
		return route.OK
	case route.Info:
		n.info(req)
		return route.OK
	}
	return route.FAIL
}

func (n *natGateways) find(s string) *natGateway {
	return n.natGateways[s]
}

func (n *natGateways) help() {
	commands := []help.Command{
		{Name: "'name'", Desc: "manage named nat gateways"},
		{Name: route.Create.String(), Desc: "create all nat gateways"},
		{Name: route.Destroy.String(), Desc: "destroy all nat gateways"},
		{Name: route.Info.String(), Desc: "show information about all allocated nat gateways"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("network natgateway", commands)
}

func (n *natGateways) info(req *route.Request) {
	if n.Destroyed() {
		return
	}
	msg.Info("NatGateways")
	msg.IndentInc()
	n.RouteInOrder(req)
	msg.IndentDec()
}
