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

// The configuration of the subnet object. It has a name, the
// starting cidr block of the subnet group, the type of access
// (public, private, local) for the subnet group, the availability
// zone where the subet is located and the flag indicating whether
// this subnet needs to have a separate routetable..
//
// A subnet is not part of the configuration file, so the existence
// of this structure is a convenience meant to hold per subnet data
// derived from the subnet group.
type Subnet struct {
	Name_             string
	GroupName_        string
	CidrBlock_        string
	Access_           string
	AvailabilityZone_ string
	ManageRoutes_     bool
}

// Name satisfies the resource.StaticSubnet interface.
// The name is derived from the name of the subnet group
// and the availability zone. I.e. bation-us-west-2a.
func (s *Subnet) Name() string {
	return s.Name_
}

func (s *Subnet) SetName(name string) {
	s.Name_ = name
}

// Name satisfies the resource.StaticSubnet interface.
// The group name is the name of the subnet group.
func (s *Subnet) GroupName() string {
	return s.GroupName_
}

// SetGroupName provides a way to set the subnet's group name at runtime.
func (s *Subnet) SetGroupName(groupName string) {
	s.GroupName_ = groupName
}

// CidrBlock satisfies the resource.StaticSubnet interface.
// The cidr block is derived from the starting cidr block
// of the subnet group that owns this subnet.
func (s *Subnet) CidrBlock() string {
	return s.CidrBlock_
}

// SetCidrBlock provides a way to set the cidr block at runtime.
func (s *Subnet) SetCidrBlock(cidrBlock string) {
	s.CidrBlock_ = cidrBlock
}

// Access satisfies the resource.StaticSubnet interface.
func (s *Subnet) Access() string {
	return s.Access_
}

// SetAccess provides a way to set the access value at runtime.
func (s *Subnet) SetAccess(access string) {
	s.Access_ = access
}

// AvailabilityZone satisfies the resource.StaticSubnet interface.
func (s *Subnet) AvailabilityZone() string {
	return s.AvailabilityZone_
}

// SetAvailabilityZone provides a way to set the availabilit zone at runtime.
func (s *Subnet) SetAvailabilityZone(availabilityZone string) {
	s.AvailabilityZone_ = availabilityZone
}

// ManageRoutes satisfies the resource.StaticSubnet interface.
func (s *Subnet) ManageRoutes() bool {
	return s.ManageRoutes_
}

// SetManageRoutes provides a way to set the manage routes value at runtime.
func (s *Subnet) SetManageRoutes(manageRoutes bool) {
	s.ManageRoutes_ = manageRoutes
}

// PrintLocal provides a user friendly way to view the configuration local to the arc object.
func (s *Subnet) PrintLocal() {
	msg.Info("Subnet Config")
	msg.Detail("%-20s\t%s", "name", s.Name())
	msg.Detail("%-20s\t%s", "group name", s.GroupName())
	msg.Detail("%-20s\t%s", "cidr", s.CidrBlock())
	msg.Detail("%-20s\t%s", "access", s.Access())
	msg.Detail("%-20s\t%s", "availability zone", s.AvailabilityZone())
	msg.Detail("%-20s\t%s", "manage routes", s.ManageRoutes())
}

// Print provides a user friendly way to view the subnet configuration.
func (s *Subnet) Print() {
	s.PrintLocal()
}
