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

// SecurityGroups is a collection of SecurityGroup objects.
type SecurityGroups []*SecurityGroup

// Print provides a user friendly way to view the security groups configuration.
func (s *SecurityGroups) Print() {
	msg.Info("SecurityGroups Config")
	msg.IndentInc()
	for _, securityGroup := range *s {
		securityGroup.Print()
	}
	msg.IndentDec()
}

// The configuration of the security group object. It has a name, and
// a collection of security rules.
type SecurityGroup struct {
	Name_         string         `json:"security_group"`
	SecurityRules *SecurityRules `json:"rules"`
}

// Name satisfies the resource.StaticSecurityGroup interface.
func (s *SecurityGroup) Name() string {
	return s.Name_
}

// PrintLocal provides a user friendly way to view the configuration local to the security group object.
func (s *SecurityGroup) PrintLocal() {
	msg.Info("SecurityGroup Config")
	msg.Detail("%-20s\t%s", "name", s.Name())
}

// Print provides a user friendly way to view a security group configuration.
func (s *SecurityGroup) Print() {
	s.PrintLocal()
	msg.IndentInc()
	if s.SecurityRules != nil {
		s.SecurityRules.Print()
	}
	msg.IndentDec()
}
