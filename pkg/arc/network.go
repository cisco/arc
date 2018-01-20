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

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type network struct {
	*resource.Resources
	*config.Network
	dc                  *dataCenter
	providerNetwork     resource.ProviderNetwork
	subnetGroups        *subnetGroups
	providerNetworkPost resource.ProviderNetworkPost
	securityGroups      *securityGroups
}

// newNetwork is the constructor for a network object. It returns a non-nil error upon failure.
func newNetwork(dc *dataCenter, prov provider.DataCenter, cfg *config.Network) (*network, error) {
	if cfg == nil {
		return nil, nil
	}
	log.Debug("Initializing Network")

	// Validate the config.Network object.
	if cfg.SubnetGroups == nil {
		return nil, fmt.Errorf("The subnet_groups element is missing from the network configuration")
	}
	if cfg.SecurityGroups == nil {
		return nil, fmt.Errorf("The security_groups element is missing from the network configuration")
	}

	n := &network{
		Resources: resource.NewResources(),
		Network:   cfg,
		dc:        dc,
	}

	var err error

	// Delegate the provider specific network behavior to the resource.ProviderNetwork object.
	n.providerNetwork, err = prov.NewNetwork(cfg)
	if err != nil {
		return nil, err
	}
	n.Append(n.providerNetwork)

	n.subnetGroups, err = newSubnetGroups(n, prov, cfg.SubnetGroups)
	if err != nil {
		return nil, err
	}
	n.Append(n.subnetGroups)

	n.providerNetworkPost, err = prov.NewNetworkPost(n, cfg)
	if err != nil {
		return nil, err
	}
	n.Append(n.providerNetworkPost)

	n.securityGroups, err = newSecurityGroups(n, prov, cfg.SecurityGroups)
	if err != nil {
		return nil, err
	}
	n.Append(n.securityGroups)

	return n, nil
}

// Id provides the id of the provider specific network resource.
// This satisfies the resource.DynamicNetwork interface.
func (n *network) Id() string {
	if n.providerNetwork == nil {
		return ""
	}
	return n.providerNetwork.Id()
}

// State provides the state of the provider specific network resource.
// This satisfies the resource.DynamicNetwork interface.
func (n *network) State() string {
	if n.providerNetwork == nil {
		return ""
	}
	return n.providerNetwork.State()
}

func (n *network) AuditSubnets(flags ...string) error {
	return n.providerNetwork.AuditSubnets(flags...)
}

func (n *network) AuditSecgroups(flags ...string) error {
	return n.providerNetwork.AuditSecgroups(flags...)
}

// ProviderNetwork satisfies the resource.Network interface and provides access
// to the provider's network.
func (n *network) ProviderNetwork() resource.ProviderNetwork {
	return n.providerNetwork
}

// DataCenter satisfies the resource.Network interface and provides access
// to network's parent.
func (n *network) DataCenter() resource.DataCenter {
	return n.dc
}

// DataCenter satisfies the resource.Network interface and provides access
// to network's subnet groups.
func (n *network) SubnetGroups() resource.SubnetGroups {
	return n.subnetGroups
}

// DataCenter satisfies the resource.Network interface and provides access
// to network's security groups.
func (n *network) SecurityGroups() resource.SecurityGroups {
	return n.securityGroups
}

// CidrAlias returns the alias for the given name.
func (n *network) CidrAlias(s string) string {
	return n.CidrAliases_[s]
}

// CidrGroup returns the cidrs for the given name.
func (n *network) CidrGroup(s string) []string {
	return n.CidrGroups_[s]
}

// Route satisfies the embedded resource.Resource interface in resource.Network.
// Network handles load, create, destroy, help, config and info requests
// in order to manage the entire network. All other commands are routed to
// network's children.
func (n *network) Route(req *route.Request) route.Response {
	log.Route(req, "Network")

	// Is the user associated with the request allowed to do network commands?
	if err := aaa.Authorized(req, "Network", n.Name()); err != nil {
		msg.Error(err.Error())
		return route.UNAUTHORIZED
	}

	switch req.Top() {
	case "subnet":
		return n.SubnetGroups().Route(req.Pop())
	case "secgroup":
		return n.SecurityGroups().Route(req.Pop())
	}
	if n.providerNetwork.CanRoute(req) {
		return n.providerNetwork.Route(req)
	}
	if n.providerNetworkPost.CanRoute(req) {
		return n.providerNetworkPost.Route(req)
	}
	if req.Top() != "" {
		n.help()
		return route.FAIL
	}

	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load, route.Create, route.Audit:
		return n.RouteInOrder(req)
	case route.Destroy:
		return n.RouteReverseOrder(req)
	case route.Help:
		n.help()
		return route.OK
	case route.Config:
		n.config()
		return route.OK
	case route.Info:
		n.info(req)
		return route.OK
	default:
		msg.Error("Unknown network command %q.", req.Command().String())
	}
	return route.FAIL
}

func (n *network) help() {
	providerCommands := help.Append(n.providerNetwork.HelpCommands(), n.providerNetworkPost.HelpCommands())
	commands := []help.Command{
		{route.Create.String(), "create all network resources"},
		{route.Destroy.String(), "destroy all network resources"},
		{route.Config.String(), "show the network configuration"},
		{route.Info.String(), "show information about allocated network resource"},
		{route.Help.String(), "show this help"},
	}
	commands = help.Append(providerCommands, commands)
	help.Print("network", commands)
}

func (n *network) config() {
	n.Network.Print()
}

func (n *network) info(req *route.Request) {
	if n.Destroyed() {
		return
	}
	msg.Info("Network")
	msg.IndentInc()
	n.Network.PrintLocal()
	n.RouteInOrder(req)
	msg.IndentDec()
}
