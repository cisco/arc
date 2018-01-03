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

package config

import "github.com/cisco/arc/pkg/msg"

// The configuration of the network object. It has a name, a
// cidr block, a list of availability zones (one or more), a list of
// dns name server ip addresses, a subnet groups element and a
// security groups element.
//
// Note that the name is a convenience field and isn't part of the
// configuration file. It is set by the application at run time.
type Network struct {
	Name_              string
	CidrBlock_         string              `json:"cidr"`
	AvailabilityZones_ []string            `json:"availability_zones"`
	DnsNameServers_    []string            `json:"dns_name_servers"`
	CidrAliases_       map[string]string   `json:"cidr_aliases"`
	CidrGroups_        map[string][]string `json:"cidr_groups"`
	SubnetGroups       *SubnetGroups       `json:"subnet_groups"`
	SecurityGroups     *SecurityGroups     `json:"security_groups"`
}

// Name satisfies the resource.StaticNetwork interface.
func (n *Network) Name() string {
	return n.Name_
}

// SetName is a convenience function to set the name at run time.
func (n *Network) SetName(name string) {
	n.Name_ = name
}

// CidrBlock satisfies the resource.StaticNetwork interface.
func (n *Network) CidrBlock() string {
	return n.CidrBlock_
}

// AvailabilityZones satisfies the resource.StaticNetwork interface.
func (n *Network) AvailabilityZones() []string {
	return n.AvailabilityZones_
}

// DnsNameServers satisfies the resource.StaticNetwork interface.
func (n *Network) DnsNameServers() []string {
	return n.DnsNameServers_
}

// CidrAliases satisfies the resource.StaticNetwork interface.
func (n *Network) CidrAliases() map[string]string {
	return n.CidrAliases_
}

// CidrGroups satisfies the resource.StaticNetwork interface.
func (n *Network) CidrGroups() map[string][]string {
	return n.CidrGroups_
}

// PrintLocal provides a user friendly way to view the configuration local to the network object.
func (n *Network) PrintLocal() {
	msg.Info("Network Config")
	msg.Detail("%-20s\t%s", "name", n.Name())
	msg.Detail("%-20s\t%s", "cidr", n.CidrBlock())
	a, sep := "", ""
	for _, az := range n.AvailabilityZones() {
		a += sep + az
		sep = ", "
	}
	if a != "" {
		msg.Detail("%-20s\t%s", "availability zones", a)
	}
	d, sep := "", ""
	for _, dns := range n.DnsNameServers() {
		d += sep + dns
		sep = ", "
	}
	if d != "" {
		msg.Detail("%-20s\t%s", "dns name servers", d)
	}
	aliases := n.CidrAliases()
	if aliases != nil || len(aliases) > 0 {
		a := ""
		for k, v := range aliases {
			if a == "" {
				a += k + ":" + v
				continue
			}
			a += ", " + k + ":" + v
		}
		msg.Detail("%-20s\t%s", "cidr aliases", a)
	}
	groups := n.CidrGroups()
	if groups != nil && len(groups) > 0 {
		g := ""
		for k, v := range groups {
			if g == "" {
				g += k + ":["
			} else {
				g += ", " + k + ":["
			}
			for i, c := range v {
				if i == 0 {
					g += c
					continue
				}
				g += ", " + c
			}
			g += "]"
		}
		msg.Detail("%-20s\t%s", "cidr groups", g)
	}
}

// Print provides a user friendly way to view the entire network configuration.
func (n *Network) Print() {
	n.PrintLocal()
	msg.IndentInc()
	if n.SubnetGroups != nil {
		n.SubnetGroups.Print()
	}
	if n.SecurityGroups != nil {
		n.SecurityGroups.Print()
	}
	msg.IndentDec()
}
