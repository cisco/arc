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

// StaticDns provides the interface to the static portion of dns.
// This information is provided via config file and is implemented
// config.Dns.
type StaticDns interface {
	DomainName() string
	Subdomain() string
	Domain() string
}

// DyanmicDns provides the interface to the dynamic portion of dns.
// This information is provided by the resource allocated
// by the cloud provider.
type DynamicDns interface {

	// Id returns the Dns id.
	Id() string

	// AuditDnsRecords checks for any dns records that are unnamedo or deployed but not configured
	AuditDnsRecords(flags ...string) error
}

// Dns provides the resource interface used for the common dns object
// implemented in the arc package. It contains an Arc method used to
// access it's parent object, and ARecords and CNameRecords methods used
//  to access it's children.
type Dns interface {
	Resource
	StaticDns
	DynamicDns

	// Arc provides access to Dns' parent.
	Arc() Arc

	// ProviderDns provides access to the provider specific dns implementation.
	ProviderDns() ProviderDns

	// ARecords provides access to Dns' A records.
	ARecords() DnsRecords

	// CNameRecords provides access to Dns' CNAME records.
	CNameRecords() DnsRecords

	// DataCenter provides access to the DataCenter resource.
	DataCenter() DataCenter
}

// ProviderDns provides an interface for the provider supplied dns resource.
type ProviderDns interface {
	DynamicDns
}
