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

type securityGroup struct {
	*resource.Resources
	*config.SecurityGroup
	network               *network
	providerSecurityGroup resource.ProviderSecurityGroup
}

// newSecurityGroup is the constructor for a securityGroup object. It returns a non-nil error upon failure.
func newSecurityGroup(net *network, prov provider.DataCenter, cfg *config.SecurityGroup) (*securityGroup, error) {
	log.Debug("Initializing SecurityGroup %q", cfg.Name())

	// Validate the config.SecurityGroup object.
	if cfg.SecurityRules == nil {
		return nil, fmt.Errorf("The rules element is missing from the security_groups configuration")
	}
	s := &securityGroup{
		Resources:     resource.NewResources(),
		SecurityGroup: cfg,
		network:       net,
	}

	var err error
	s.providerSecurityGroup, err = prov.NewSecurityGroup(net, cfg)
	if err != nil {
		return nil, err
	}
	s.Append(s.providerSecurityGroup)

	return s, nil
}

// Id provides the id of the provider specific security group resource.
// This satisfies the resource.DynamicSecurityGroup interface.
func (s *securityGroup) Id() string {
	if s.providerSecurityGroup == nil {
		return ""
	}
	return s.providerSecurityGroup.Id()
}

// ProviderSecurityGroup satisfies the resource.SecurityGroup interface and provides access
// to the provider's security group.
func (s *securityGroup) ProviderSecurityGroup() resource.ProviderSecurityGroup {
	return s.providerSecurityGroup
}

// Network satisfies the resource.SecurityGroup interface and provides access
// to security group's parent.
func (s *securityGroup) Network() resource.Network {
	return s.network
}

// Route satisfies the embedded resource.Resource interface in resource.SecurityGroup.
// SecurityGroup handled load, create, destroy, audit, help, config and info requests
// to manage a named ssecurity group.
func (s *securityGroup) Route(req *route.Request) route.Response {
	log.Route(req, "SecurityGroup %q", s.Name())

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
	case route.Load:
		if err := s.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Info, route.Create, route.Provision:
		return s.RouteInOrder(req)
	case route.Audit:
		if err := s.Audit("Secgroup"); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Destroy:
		return s.RouteReverseOrder(req)
	case route.Help:
		s.help()
	case route.Config:
		s.config()
	default:
		msg.Error("Unknown secgroup command %q.", req.Command().String())
		return route.FAIL
	}
	return route.OK
}

func (s *securityGroup) Load() error {
	return s.providerSecurityGroup.Load()
}

func (s *securityGroup) help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: fmt.Sprintf("create %s security group", s.Name())},
		{Name: route.Audit.String(), Desc: fmt.Sprintf("audit %s security groups", s.Name())},
		{Name: route.Provision.String(), Desc: fmt.Sprintf("update %s security groups", s.Name())},
		{Name: route.Destroy.String(), Desc: fmt.Sprintf("destroy %s security group", s.Name())},
		{Name: route.Config.String(), Desc: fmt.Sprintf("show the %s security group configuration", s.Name())},
		{Name: route.Info.String(), Desc: fmt.Sprintf("show information about allocated i%s security group", s.Name())},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print(fmt.Sprintf("secgroup %s", s.Name()), commands)
}

func (s *securityGroup) config() {
	s.SecurityGroup.Print()
}

func (s *securityGroup) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	if aaa.AuditBuffer[flags[0]] == nil {
		err := aaa.NewAudit(flags[0])
		if err != nil {
			return err
		}
	}
	return s.providerSecurityGroup.Audit(flags...)
}
