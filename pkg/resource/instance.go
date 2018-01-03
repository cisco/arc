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

package resource

import "github.com/cisco/arc/pkg/route"

// StaticInstance provides the interface to the static portion of the
// instance. This information is provided via config file and is implemented
// by config.Instance.
type StaticInstance interface {
	Name() string
	Count() int
	ServerType() string
	Version() string
	Image() string
	InstanceType() string
	SubnetGroupName() string
	SecurityGroupNames() []string
	Teams() []string
}

// DynamicInstance provides access to the dynamic portion of the instance.
type DynamicInstance interface {

	// Id returns the id of the instance.
	Id() string

	// ImageId returns the imageId.
	ImageId() string

	// KeyName returns the name of the keypair.
	KeyName() string

	// State returns the state of the instance.
	State() string

	// Started returns true if the instance has been started.
	Started() bool

	// Stopped returns true if the instance is currently stopped.
	Stopped() bool

	// PrivateIPAddress returns the private IP address associated with the instance.
	PrivateIPAddress() string

	// PublicIPAddress returns the public IP address associated with the instance.
	PublicIPAddress() string

	// SetTags sets the tags for the instance such as who created it and the last person
	// that modified the instance.
	SetTags(map[string]string) error

	Auditor
}

// Instance provides the resource interface used for the common instance
// object implemented in the arc package.
type Instance interface {
	Resource
	StaticInstance
	DynamicInstance

	// Pod provides access to Instance's parent.
	Pod() Pod

	// Network provides access to the network to which instance is associated.
	Network() Network

	// Subnet provides access to the subnet to which instance is allocated.
	Subnet() Subnet

	// SecurityGroups provides access to the security groups to which instance is associated.
	SecurityGroups() []SecurityGroup

	// KeyPair provides access to the keypair that will be assigned to this instance.
	KeyPair() KeyPair

	// Dns provides access to the dns associated with the datacenter.
	Dns() Dns

	// PrivateHostname returns the hostname (without the domain name) associated with the
	// private ip address of the instance.
	PrivateHostname() string

	// PrivateFQDN returns the FQDN associated with the private ip address of the instance.
	PrivateFQDN() string

	// PublicHostname returns the hostname (without the domain name) associated with the
	// public ip address of the instance.
	PublicHostname() string

	// PublicFQDN returns the FQDN associated with the public ip address of the instance.
	// If this instance doesn't have a public ip address (most don't) this will return
	// an empty string "".
	PublicFQDN() string

	// PrivateDnsARecord returns the dns a record associated with this instance.
	PrivateDnsARecord() DnsRecord

	// PublicDnsARecord return the dns a record associated with the public ip address of this
	// instance. If this instance doesn't have a public ip address (most don't) this will
	// return nil.
	PublicDnsARecord() DnsRecord

	// FQDNMatch returns true if the given list of values contains a match to either the
	// private or public fqdn.
	FQDNMatch([]string) bool

	// RootUser returns the name of the root user for the given image type.
	RootUser() string

	// ProviderVolumes returns a slice of provider created volumes.
	ProviderVolumes() []ProviderVolume

	// ProviderRole returns the role of the instance.
	ProviderRole() ProviderRole

	// Creator
	PreCreate(req *route.Request) route.Response
	Create(req *route.Request) route.Response
	PostCreate(req *route.Request) route.Response

	// Destroyer
	PreDestroy(req *route.Request) route.Response
	Destroy(req *route.Request) route.Response
	PostDestroy(req *route.Request) route.Response

	// Provisioner
	PreProvision(req *route.Request) route.Response
	Provision(req *route.Request) route.Response
	PostProvision(req *route.Request) route.Response

	// Starter
	PreStart(req *route.Request) route.Response
	Start(req *route.Request) route.Response
	PostStart(req *route.Request) route.Response

	// Stopper
	PreStop(req *route.Request) route.Response
	Stop(req *route.Request) route.Response
	PostStop(req *route.Request) route.Response

	// Restarter
	PreRestart(req *route.Request) route.Response
	Restart(req *route.Request) route.Response
	PostRestart(req *route.Request) route.Response

	// Replacer
	PreReplace(req *route.Request) route.Response
	Replace(req *route.Request) route.Response
	PostReplace(req *route.Request) route.Response

	AuditOverride
}

// ProviderInstance provides a resource interface for the provider supplied
// instance. The common instance object delegates provider specific
// behavior to the raw instance.
type ProviderInstance interface {
	Resource
	DynamicInstance
}
