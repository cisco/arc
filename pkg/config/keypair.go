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

// The configuration of the keypair object. It has a name, a local name,
// a format, a comment and keymaterial. The name is used to identify the
// keypair in cloud provider, whereas the local name is the name used to
// identify the keypair locally. This only matters for the id_rsa key,
// where the name will be the username and the local name will be "id_rsa".
type KeyPair struct {
	Name_        string
	LocalName_   string
	Format_      string
	Comment_     string
	KeyMaterial_ string
}

// Name satisfies the resource.StaticKeyPair interface. Name known to
// the cloud provider.
func (k *KeyPair) Name() string {
	return k.Name_
}

// LocalName satisfies the resource.StaticKeyPair interface. Name known
// local to the execution of arc.
func (k *KeyPair) LocalName() string {
	return k.LocalName_
}

// Format satisfies the resource.StaticKeyPair interface.
func (k *KeyPair) Format() string {
	return k.Format_
}

// Comment satisfies the resource.StaticKeyPair interface. This will be populated with the filename of the key.
func (k *KeyPair) Comment() string {
	return k.Comment_
}

// KeyMaterial satisfies the resource.StaticKeyPair interface. Contains the base64 encoded serialized key.
func (k *KeyPair) KeyMaterial() string {
	return k.KeyMaterial_
}

// PrintLocal provides a user friendly way to view the configuration local to the subnet group object.
func (k *KeyPair) PrintLocal() {
	msg.Info("KeyPair Config")
	msg.Detail("%-20s\t%s", "name", k.Name())
	msg.Detail("%-20s\t%s", "local name", k.LocalName())
	msg.Detail("%-20s\t%s", "format", k.Format())
	msg.Detail("%-20s\t%s", "comment", k.Comment())
	msg.Detail("%-20s\t%s", "key material", k.KeyMaterial())
}

// Print provides a user friendly way to view a subnet group configuration.
func (k *KeyPair) Print() {
	k.PrintLocal()
}
