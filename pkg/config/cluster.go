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

// Clusters is a collection of Cluster objects.
type Clusters []*Cluster

// Print provides a user friendly way to view the clusters configuration.
func (c *Clusters) Print() {
	msg.Info("Clusters Config")
	msg.IndentInc()
	for _, cluster := range *c {
		cluster.Print()
	}
	msg.IndentDec()
}

// The configuration of the clutser object. It has a name and a
// pods element.
type Cluster struct {
	Name_         string       `json:"cluster"`
	Pods          *Pods        `json:"pods"`
	SecurityTags_ SecurityTags `json:"security_tags"`
	AuditIgnore_  bool         `json:"audit_ignore"`
}

// Name satisfies the resource.StaticCluster interface.
func (c *Cluster) Name() string {
	return c.Name_
}

func (c *Cluster) SecurityTags() SecurityTags {
	return c.SecurityTags_
}

func (c *Cluster) AuditIgnore() bool {
	return c.AuditIgnore_
}

// Print provides a user friendly way to view the cluster configuration.
func (c *Cluster) Print() {
	msg.Info("Cluster Config")
	msg.Detail("%-20s\t%s", "name", c.Name())
	msg.IndentInc()
	if c.Pods != nil {
		c.Pods.Print()
	}
	if c.SecurityTags_ != nil {
		c.SecurityTags_.Print()
	}
	msg.IndentDec()
}
