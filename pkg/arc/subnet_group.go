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
	"github.com/cisco/arc/pkg/net"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type subnetGroup struct {
	*resource.Resources
	*config.SubnetGroup
	network *network
	subnets map[string]resource.Subnet
}

// newSubnetGroup is the constructor for a subnetGroup object. It returns a non-nil error upon failure.
func newSubnetGroup(n *network, prov provider.DataCenter, cfg *config.SubnetGroup) (*subnetGroup, error) {
	log.Debug("Initializing SubnetGroup %q", cfg.Name())

	s := &subnetGroup{
		Resources:   resource.NewResources(),
		SubnetGroup: cfg,
		network:     n,
		subnets:     map[string]resource.Subnet{},
	}

	cidrBlock := cfg.CidrBlock()
	first := true
	var err error
	for _, az := range n.AvailabilityZones() {
		if first {
			first = false
		} else {
			cidrBlock, err = net.NextCidrBlock(cidrBlock)
			if err != nil {
				return nil, err
			}
		}
		name := cfg.Name() + "-" + az
		conf := &config.Subnet{}
		conf.SetName(name)
		conf.SetGroupName(cfg.Name())
		conf.SetCidrBlock(cidrBlock)
		conf.SetAccess(cfg.Access())
		conf.SetAvailabilityZone(az)
		conf.SetManageRoutes(cfg.ManageRoutes())

		subnet, err := newSubnet(n, prov, conf)
		if err != nil {
			return nil, err
		}
		s.subnets[name] = subnet
		s.Append(subnet)
	}
	return s, nil
}

// Subnets satifies the resource.SubnetGroup interface and provide
// access to the child subnets of the subnet group.
func (s *subnetGroup) Subnets() map[string]resource.Subnet {
	return s.subnets
}

// Find satisfies the resource.SubnetGroup interface and provides a way
// to search for a specific subnet. This assumes subnet names are unique.
func (s *subnetGroup) Find(name string) resource.Subnet {
	return s.subnets[name]
}

// Route satisfies the embedded resource.Resource interface in resource.SubnetGroup.
// SubnetGroup handled load, create, destroy, help, config and info requests
// to manage a named subnet group.
func (s *subnetGroup) Route(req *route.Request) route.Response {
	log.Route(req, "SubnetGroup %q", s.Name())

	if req.Top() != "" {
		s.help()
		return route.FAIL
	}

	// Skip if the test flag is set.
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load, route.Create:
		return s.RouteInOrder(req)
	case route.Audit:
		if err := s.Audit("Subnet"); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Destroy:
		return s.RouteReverseOrder(req)
	case route.Help:
		s.help()
		return route.OK
	case route.Config:
		s.config()
		return route.OK
	case route.Info:
		s.info(req)
		return route.OK
	default:
		msg.Error("Unknown subnet command %q.", req.Command())
	}
	return route.FAIL
}

func (s *subnetGroup) help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: fmt.Sprintf("create %s subnet group", s.Name())},
		{Name: route.Destroy.String(), Desc: fmt.Sprintf("destroy %s subnet group", s.Name())},
		{Name: route.Config.String(), Desc: fmt.Sprintf("show the %s subnet group configuration", s.Name())},
		{Name: route.Info.String(), Desc: fmt.Sprintf("show information about allocated %s subnet group", s.Name())},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print(fmt.Sprintf("subnet %s", s.Name()), commands)
}

func (s *subnetGroup) config() {
	s.SubnetGroup.Print()
}

func (s *subnetGroup) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	if aaa.AuditBuffer[flags[0]] == nil {
		err := aaa.NewAudit("Subnet")
		if err != nil {
			return err
		}
	}
	for _, v := range s.subnets {
		if err := v.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (s *subnetGroup) info(req *route.Request) {
	if !s.Created() {
		return
	}
	msg.Info("SubnetGroup")
	msg.IndentInc()
	s.SubnetGroup.PrintLocal()
	s.RouteInOrder(req)
	msg.IndentDec()
}
