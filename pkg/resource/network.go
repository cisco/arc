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

package resource

import (
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/route"
)

// StaticNetwork provides the interface to the static portion of the
// network. This information is provided via config file and is implemented
// by config.Network.
type StaticNetwork interface {
	Name() string
	CidrBlock() string
	AvailabilityZones() []string
	DnsNameServers() []string
	CidrAliases() map[string]string
	CidrGroups() map[string][]string
}

// DyanmicNetwork provides the interface to the dynamic portion of the
// network. This information is provided by the resource allocated
// by the cloud provider.
type DynamicNetwork interface {

	// Id returns the id of the network.
	Id() string

	// State returns the state of the network.
	State() string

	// AuditSubnets identifies any subnets that have been deployed but are not in the configuration.
	AuditSubnets(flags ...string) error

	// AuditSecgroups indentifies any secgroups that have been deployed but are not configured.
	AuditSecgroups(flags ...string) error
}

// Network provides the resource interface used for the common network
// object implemented in the arc package. It contains a DataCenter method
// to access it's parent, and the SubnetGroups and SecurityGroups methods
// to access it's children.
type Network interface {
	Resource
	StaticNetwork
	DynamicNetwork

	// ProviderNetwork provides access to the provider specific network.
	ProviderNetwork() ProviderNetwork

	// DataCenter provides access to Network's parent.
	DataCenter() DataCenter

	// SubnetGroups provides access to Network's child subnet groups.
	SubnetGroups() SubnetGroups

	// SecurityGroups provides access to Network's child security groups.
	SecurityGroups() SecurityGroups

	// CidrAlias return the alias for the given name.
	CidrAlias(string) string

	// CidrGroup returns the slice of cidr values for the given name.
	CidrGroup(string) []string
}

// ProviderNetwork provides a resource interface for the provider supplied
// resources. The common network object, Network, delegates provider specific
// behavior to the raw network.
type ProviderNetwork interface {
	Resource
	DynamicNetwork

	// Does the ProviderNetwork support provider specific requests?
	CanRoute(*route.Request) bool

	// Additional help commands for provider specific requests.
	HelpCommands() []help.Command
}

type NetworkPost interface {
	Resource
	StaticNetwork
	ProviderNetworkPost() ProviderNetworkPost
}

type ProviderNetworkPost interface {
	Resource

	// Does the ProviderNetwork support provider specific requests?
	CanRoute(*route.Request) bool

	// Additional help commands for provider specific requests.
	HelpCommands() []help.Command
}
