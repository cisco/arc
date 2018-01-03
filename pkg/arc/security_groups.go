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

type securityGroups struct {
	*resource.Resources
	*config.SecurityGroups
	securityGroups map[string]resource.SecurityGroup
	network        *network
}

// newSecurityGroups is the constructor for a securityGroups object. It returns a non-nil error upon failure.
func newSecurityGroups(net *network, prov provider.DataCenter, cfg *config.SecurityGroups) (*securityGroups, error) {
	log.Debug("Initializing SecurityGroups")

	s := &securityGroups{
		Resources:      resource.NewResources(),
		SecurityGroups: cfg,
		securityGroups: map[string]resource.SecurityGroup{},
		network:        net,
	}

	for _, conf := range *cfg {
		if s.Find(conf.Name()) != nil {
			return nil, fmt.Errorf("Security group name %q must be unique but is used multiple times", conf.Name())
		}
		securityGroup, err := newSecurityGroup(net, prov, conf)
		if err != nil {
			return nil, err
		}
		s.securityGroups[conf.Name()] = securityGroup
		s.Append(securityGroup)
	}
	return s, nil
}

// Find satisfies the resource.SecurityGroup interface and provides a way
// to search for a specific security group. This assumes security group
// names are unique.
func (s *securityGroups) Find(name string) resource.SecurityGroup {
	return s.securityGroups[name]
}

// Route satisfies the embedded resource.Resource interface in resource.SecurityGroups.
// SecurityGroups handled load, create, destroy, audit, help, config and info requests
// in order to manage all security groups. All other commands are routed to
// a named security group.
func (s *securityGroups) Route(req *route.Request) route.Response {
	log.Route(req, "SecurityGroups")

	group := s.Find(req.Top())
	if group != nil {
		return group.Route(req.Pop())
	}
	if req.Top() != "" {
		msg.Error("Unknown secgroup %q.", req.Top())
		return route.FAIL
	}

	// Skip if the test flag is set.
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load:
		return s.RouteInOrder(req)
	case route.Create:
		if s.Created() {
			msg.Detail("\nSecurityGroups exists, skipping...")
			return route.OK
		}
		// If the create has a flag we are not going to perform a two pass create
		// as described below. This is most likely an explicit "create norules".
		if !req.Flags().Empty() {
			return s.RouteInOrder(req)
		}
		// If we want to create all security groups at once, and since security
		// groups must exist before they can be referenced as a remote target in a
		// security rule, we are going to create the security groups in two
		// passes. The first pass is to create the security groups without the
		// security rules. The second pass is to update the rules, creating
		// them in the process.
		//
		// First pass - create norules
		create := req.Clone(route.Create)
		create.Flags().Append("norules")
		resp := s.RouteInOrder(create)
		if resp != route.OK {
			return resp
		}
		// Second pass - load then provision
		// TA322783: The load ensures that we are updating all of the security
		// rules including the security rules for default security group.
		load := req.Clone(route.Load)
		if resp := s.RouteInOrder(load); resp != route.OK {
			return resp
		}
		update := req.Clone(route.Provision)
		if resp := s.RouteInOrder(update); resp != route.OK {
			return resp
		}
		return route.OK
	case route.Provision:
		return s.RouteInOrder(req)
	case route.Audit:
		if err := s.Audit(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Destroy:
		if s.Destroyed() {
			msg.Detail("\nSecurityGroups does not exist, skipping...")
			return route.OK
		}
		if !req.Flags().Empty() {
			return s.RouteReverseOrder(req)
		}
		// Needs two passes to delete the security groups because the security rules
		// can be referencing a security group, so the security rules must be
		// deleted before the security groups.
		//
		// First pass - destroy rules_only
		destroyRules := req.Clone(route.Destroy)
		destroyRules.Flags().Append("rules_only")
		resp := s.RouteReverseOrder(destroyRules)
		if resp != route.OK {
			return resp
		}
		// Second pass - destroy secgroups
		destroyGroups := req.Clone(route.Destroy)
		return s.RouteReverseOrder(destroyGroups)
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
		msg.Error("Unkown secgroup command %q.", req.Command().String())
	}
	return route.FAIL
}

func (s *securityGroups) help() {
	commands := []help.Command{
		{route.Create.String(), "create all security groups"},
		{route.Audit.String(), "audit all security groups"},
		{route.Provision.String(), "update all security groups"},
		{route.Destroy.String(), "destroy all security groups"},
		{"'name'", "manage named security group"},
		{route.Config.String(), "show the security groups configuration"},
		{route.Info.String(), "show information about allocated security groups"},
		{route.Help.String(), "show this help"},
	}
	help.Print("secgroup", commands)
}

func (s *securityGroups) config() {
	s.SecurityGroups.Print()
}

func (s *securityGroups) info(req *route.Request) {
	if s.Destroyed() {
		return
	}
	msg.Info("SecurityGroups")
	msg.IndentInc()
	s.RouteInOrder(req)
	msg.IndentDec()
}

func (s *securityGroups) Audit(flags ...string) error {
	err := aaa.NewAudit("Secgroup")
	if err != nil {
		return err
	}
	if err := s.network.AuditSecgroups("Secgroup"); err != nil {
		return err
	}
	for _, v := range s.securityGroups {
		if err := v.Audit("Secgroup"); err != nil {
			return err
		}
	}

	return nil
}
