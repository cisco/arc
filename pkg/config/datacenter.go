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

// The configuration of the datacenter object. It has a provider element,
// a network element and a compute element.
type DataCenter struct {
	Provider      *Provider    `json:"provider"`
	Network       *Network     `json:"network"`
	Compute       *Compute     `json:"compute"`
	SecurityTags_ SecurityTags `json:"security_tags"`
}

func (d *DataCenter) SecurityTags() SecurityTags {
	return d.SecurityTags_
}

// Print provides a user friendly way to view the entire datacenter configuration.
// This is a deep print.
func (d *DataCenter) Print() {
	msg.Info("DataCenter Config")
	msg.IndentInc()
	if d.Provider != nil {
		d.Provider.Print()
	}
	if d.SecurityTags_ != nil {
		d.SecurityTags_.Print()
	}
	if d.Network != nil {
		d.Network.Print()
	}
	if d.Compute != nil {
		d.Compute.Print()
	}
	msg.IndentDec()
}
