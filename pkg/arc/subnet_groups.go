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

type subnetGroups struct {
	*resource.Resources
	*config.SubnetGroups
	network      *network
	subnetGroups map[string]resource.SubnetGroup
}

// newSubnetGroups is the constructor for a subnetGroups object. It returns a non-nil error upon failure.
func newSubnetGroups(net *network, prov provider.DataCenter, cfg *config.SubnetGroups) (*subnetGroups, error) {
	log.Debug("Initializing SubnetGroups")

	s := &subnetGroups{
		Resources:    resource.NewResources(),
		SubnetGroups: cfg,
		network:      net,
		subnetGroups: map[string]resource.SubnetGroup{},
	}

	for _, conf := range *cfg {
		if s.Find(conf.Name()) != nil {
			return nil, fmt.Errorf("Subnet group name %q must be unique but is used multiple times", conf.Name())
		}
		subnetGroup, err := newSubnetGroup(net, prov, conf)
		if err != nil {
			return nil, err
		}
		s.subnetGroups[conf.Name()] = subnetGroup
		s.Append(subnetGroup)
	}
	return s, nil
}

// Find satisfies the resource.SubnetGroup interface and provides a way
// to search for a specific subnet group. This assumes subnet group
// names are unique.
func (s *subnetGroups) Find(name string) resource.SubnetGroup {
	return s.subnetGroups[name]
}

// Route satisfies the embedded resource.Resource interface in resource.SubnetGroups.
// SubnetGroups can terminate load, create, destroy, help, config and info requests
// in order to manage all subnet groups. All other commands are routed to
// a named subnet group.
func (s *subnetGroups) Route(req *route.Request) route.Response {
	log.Route(req, "SubnetGroups")

	group := s.Find(req.Top())
	if group != nil {
		return group.Route(req.Pop())
	}
	if req.Top() != "" {
		msg.Error("Unknown subnet %q.", req.Top())
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
	case route.Audit:
		if err := s.Audit(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	default:
		msg.Error("Unknown subnet command %q.", req.Command().String())
	}
	return route.FAIL
}

func (s *subnetGroups) help() {
	commands := []help.Command{
		{route.Create.String(), "create all subnet groups"},
		{route.Destroy.String(), "destroy all subnet groups"},
		{"'name'", "manage named subnet group"},
		{route.Config.String(), "show the subnet groups configuration"},
		{route.Info.String(), "show information about allocated subnet groups"},
		{route.Help.String(), "show this help"},
	}
	help.Print("subnet", commands)
}

func (s *subnetGroups) config() {
	s.SubnetGroups.Print()
}

func (s *subnetGroups) info(req *route.Request) {
	if s.Destroyed() {
		return
	}
	msg.Info("SubnetGroups")
	msg.IndentInc()
	s.RouteInOrder(req)
	msg.IndentDec()
}

func (s *subnetGroups) Audit(flags ...string) error {
	err := aaa.NewAudit("Subnet")
	if err != nil {
		return err
	}
	if err := s.network.AuditSubnets("Subnet"); err != nil {
		return err
	}
	for _, v := range s.subnetGroups {
		if err := v.Audit("Subnet"); err != nil {
			return err
		}
	}
	return nil
}
