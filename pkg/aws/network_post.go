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
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

// networkPost implements the resource.ProviderNetworkPost interface.
type networkPost struct {
	*resource.Resources
	*config.Network
	natGateways *natGateways
}

// newNetwork constructs the aws network.
func newNetworkPost(net resource.Network, cfg *config.Network, c *ec2.EC2) (*networkPost, error) {
	log.Debug("Initializing AWS Network Post")
	np := &networkPost{
		Resources: resource.NewResources(),
		Network:   cfg,
	}

	natGateways, err := newNatGateways(c, net)
	if err != nil {
		return nil, err
	}
	np.natGateways = natGateways
	np.Append(natGateways)

	return np, nil
}

func (n *networkPost) Route(req *route.Request) route.Response {
	log.Route(req, "AWS Network Post")

	// Do sub-object routing first.
	switch req.Top() {
	case "natgateways", "natgateway", "nat", "ngw":
		return n.natGateways.Route(req.Pop())
	}

	// Handle commands
	switch req.Command() {
	case route.Load, route.Create, route.Info:
		return n.RouteInOrder(req)
	case route.Audit:
		return route.OK
	case route.Destroy:
		return n.RouteReverseOrder(req)
	}

	msg.Error("Internal Error: aws/network_post.go, Unknown network command %q.", req.Top())
	return route.FAIL
}

func (n *networkPost) CanRoute(req *route.Request) bool {
	switch req.Top() {
	case "natgateways", "natgateway", "nat", "ngw":
		return true
	}
	return false
}

func (n *networkPost) HelpCommands() []help.Command {
	return []help.Command{
		{Name: "natgateways", Desc: "manage aws nat gateways"},
		{Name: "natgateway [name]", Desc: "manage named aws natgateway"},
	}
}
