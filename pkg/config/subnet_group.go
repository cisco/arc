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

// SubnetGroups is a collection of SubnetGroup objects.
type SubnetGroups []*SubnetGroup

// Print provides a user friendly way to view the subnet groups configuration.
func (s *SubnetGroups) Print() {
	msg.Info("SubnetGroups Config")
	msg.IndentInc()
	for _, subnetGroup := range *s {
		subnetGroup.Print()
	}
	msg.IndentDec()
}

// The configuration of the subnet group object. It has a name, the
// starting cidr block of the subnet group, and the type of access
// (public, private, local) for the subnet group.
type SubnetGroup struct {
	Name_         string       `json:"subnet"`
	CidrBlock_    string       `json:"cidr"`
	CidrBlocks_   []*CidrBlock `json:"cidrs"`
	Access_       string       `json:"access"`
	ManageRoutes_ bool         `json:"manage_routes"`
}

// Name satisfies the resource.StaticSubnetGroup interface.
func (s *SubnetGroup) Name() string {
	return s.Name_
}

// CidrBlock satisfies the resource.StaticSubnetGroup interface.
// This is the original functionality and should not be used with CidrBlocks
func (s *SubnetGroup) CidrBlock() string {
	return s.CidrBlock_
}

// CidrBlocks is for additional functionality and should not be used with CidrBlock
func (s *SubnetGroup) CidrBlocks() []*CidrBlock {
	return s.CidrBlocks_
}

// ManageRoutes satisfies the resource.StaticSubnetGroup interface.
func (s *SubnetGroup) ManageRoutes() bool {
	return s.ManageRoutes_
}

// Access satisfies the resource.StaticSubnetGroup interface.
//
// Access return values and meanings.
//  public
//    Instances on the subnet can see the internet and can be seen
//    by the internet. Both public and private ip addresses are
//    allocated for an instance when it is created.
//
//  public_elastic
//    Instances on the subnet can see the internet and can be seen
//    by the internet. Only a private ip address is allocated for
//    an instance when it is created. A public elastic ip address
//    will be associated with the instance after creation.
//
//  private
//    Instances on the subnet can see the internet and cannot be
//    seen by the internet. Only a private ip address is allocated
//    for an instance when it is created.
//
//  local
//    Instances one the subnet cannot see or be seen by the
//    internet. Only a private ip address is allocated for an instance
//    when it is created.
func (s *SubnetGroup) Access() string {
	return s.Access_
}

// PrintLocal provides a user friendly way to view the configuration local to the subnet group object.
func (s *SubnetGroup) PrintLocal() {
	msg.Info("SubnetGroup Config")
	msg.Detail("%-20s\t%s", "name", s.Name())
	msg.Detail("%-20s\t%s", "cidr", s.CidrBlock())
	msg.Detail("%-20s\t%s", "access", s.Access())
	msg.Detail("%-20s\t%t", "manage routes", s.ManageRoutes())
}

// Print provides a user friendly way to view a subnet group configuration.
func (s *SubnetGroup) Print() {
	s.PrintLocal()
}

type CidrBlock struct {
	Cidr_             string `json:"cidr"`
	AvailibilityZone_ string `json:"availibility_zone"`
}

func (c *CidrBlock) Cidr() string {
	return c.Cidr_
}

func (c *CidrBlock) AvailibilityZone() string {
	return c.AvailibilityZone_
}
