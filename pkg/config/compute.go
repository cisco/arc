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

// The configuration of the compute object. It has the version
// of the bootstrap repo used to create the hiera tarball,
// a keypair element and a clusters groups element.
//
// Note that the keypair is a convenience field and isn't part of the
// configuration file. It is set by the application at run time.
type Compute struct {
	Name_             string
	BootstrapVersion_ int `json:"bootstrap_version"`
	DeployVersion_    int `json:"deploy_version"`
	SecretsVersion_   int `json:"secrets_version"`
	AideVersion_      int `json:"aide_version"`
	KeyPair           *KeyPair
	Clusters          *Clusters `json:"clusters"`
}

// Name satisfies the resource.StaticCompute interface.
func (c *Compute) Name() string {
	return c.Name_
}

// SetName is a convenience function to set the name at run time.
func (c *Compute) SetName(name string) {
	c.Name_ = name
}

// BootstrapVersion satisfies the resource.StaticCompute interface.
func (c *Compute) BootstrapVersion() int {
	return c.BootstrapVersion_
}

// DeployVersion satisfies the resource.StaticCompute interface.
func (c *Compute) DeployVersion() int {
	return c.DeployVersion_
}

// SecretsVersion satisfies the resource.StaticCompute interface.
func (c *Compute) SecretsVersion() int {
	return c.SecretsVersion_
}

// AideVersion satisfies the resource.StaticCompute interface.
func (c *Compute) AideVersion() int {
	return c.AideVersion_
}

// PrintLocal provides a user friendly way to view the configuration local to the network object.
func (c *Compute) PrintLocal() {
	msg.Info("Compute Config")
	msg.Detail("%-20s\t%d", "bootstrap version", c.BootstrapVersion())
	msg.Detail("%-20s\t%d", "deploy version", c.DeployVersion())
	msg.Detail("%-20s\t%d", "secrets version", c.SecretsVersion())
}

// Print provides a user friendly way to view the compute configuration.
func (c *Compute) Print() {
	c.PrintLocal()
	msg.IndentInc()
	if c.KeyPair != nil {
		c.KeyPair.Print()
	}
	if c.Clusters != nil {
		c.Clusters.Print()
	}
	msg.IndentDec()
}
