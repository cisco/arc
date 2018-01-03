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

// StaticDnsRecord provides the interface to the static dns record.
// This information is provided via config file and is implemented
// config.DnsRecord.
type StaticDnsRecord interface {
	Name() string
	Ttl() int
	Values() []string
}

// DynamicDnsRecord provides access to the dynamic portion of the dns record.
type DynamicDnsRecord interface {
	Auditor
	Loader

	// Id returns the DnsRecord Id.
	Id() string

	// Type returns the type of Dns record.
	Type() string

	// DynamicValues returns a slice of the values associated with a dns record.
	DynamicValues() []string
}

// DnsRecord provides the resource interface used for the common dns record
// object implemented in the arc package.
type DnsRecord interface {
	Resource
	StaticDnsRecord
	DynamicDnsRecord

	// Dns provides access to the dns record's parent.
	Dns() Dns
}

// ProviderDnsRecord provides an interface for the provider supplied dns record resource.
type ProviderDnsRecord interface {
	Resource
	DynamicDnsRecord
}
