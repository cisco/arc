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

// SecurityRules is a collection of SecurityRule objects.
type SecurityRules []*SecurityRule

// Print provides a user friendly way to view the security rules configuration.
func (s *SecurityRules) Print() {
	msg.Info("SecurityRules Config")
	msg.IndentInc()
	for _, rule := range *s {
		rule.Print()
	}
	msg.IndentDec()
}

// The configuration of the security rule object. It has a description,
// a list of destinations being either a cidr block or a subnet group,
// a collection of directions being ingress or egress, a list of applicable
// protocols and a list of ports.
type SecurityRule struct {
	Description_ string   `json:"description"`
	Directions_  []string `json:"directions"`
	Remotes_     []string `json:"remotes"`
	Protocols_   []string `json:"protocols"`
	Ports_       []string `json:"ports"`
}

// The description can be free form text.
func (s *SecurityRule) Description() string {
	return s.Description_
}

// Remotes can be either a cidr block, a subnet group or a security group.
// A cidr block takes the form of "cidr:a.b.c.d/e", e.g. cidr:10.0.0.0/24.
// A subnet group takes the form of "subnet_group:name", e.g. subnet_group:bastion.
// A security group takes the form of "security_group:name", e.g. security_group:bastion.
// Avoid using security group remotes if at all possible.
func (s *SecurityRule) Remotes() []string {
	return s.Remotes_
}

// Direction values can either be ingress indicating that the rule applies to
// incoming traffic, or egress indication that the rule applies to
// outgoing traffic.
func (s *SecurityRule) Directions() []string {
	return s.Directions_
}

// Values can be "icmp", "udp", and "tcp".
func (s *SecurityRule) Protocols() []string {
	return s.Protocols_
}

// For tcp and udp protocols values can either be a scalar port
// number, e.g. 22, or a range of ports, e.g. 1:65535.
// For icmp values represent the ICMP type and code, e.g. 8:0.
func (s *SecurityRule) Ports() []string {
	return s.Ports_
}

// PrintLocal provides a user friendly way to view the configuration local to the security rule object.
func (s *SecurityRule) PrintLocal() {
	msg.Info("SecurityRule Config")
	msg.Detail("%-20s\t%s", "description", s.Description())
	d, sep := "", ""
	for _, direction := range s.Directions() {
		d += sep + direction
		sep = ", "
	}
	msg.Detail("%-20s\t%s", "directions", d)
	r, sep := "", ""
	for _, remote := range s.Remotes() {
		r += sep + remote
		sep = ", "
	}
	msg.Detail("%-20s\t%s", "remotes", r)
	p, sep := "", ""
	for _, protocol := range s.Protocols() {
		p += sep + protocol
		sep = ", "
	}
	msg.Detail("%-20s\t%s", "protocols", p)
	pt, sep := "", ""
	for _, port := range s.Ports() {
		pt += sep + port
		sep = ", "
	}
	msg.Detail("%-20s\t%s", "ports", pt)
}

// Print provides a user friendly way to view a security rule configuration.
func (s *SecurityRule) Print() {
	s.PrintLocal()
}
