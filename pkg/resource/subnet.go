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

// StaticSubnet provides the interface to the static portion of the
// subnet.
type StaticSubnet interface {
	Name() string
	GroupName() string
	CidrBlock() string
	Access() string
	AvailabilityZone() string
	ManageRoutes() bool
}

// DyanmicSubnet provides the interface to the dynamic portion of the
// subnet. This information is provided by the resource allocated
// by the cloud provider.
type DynamicSubnet interface {
	Auditor
	Loader

	// Id returns the id of the subnet.
	Id() string

	// State returns the state of the subnet.
	State() string
}

// Subnet provides the resource interface used for the common subnet
// object implemented in the arc package.
type Subnet interface {
	Resource
	StaticSubnet
	DynamicSubnet

	// ProviderSubnet provides access to the provider specific subnet.
	ProviderSubnet() ProviderSubnet

	// Network provides access to Subnet's parent.
	Network() Network
}

// ProviderSubnet provides a resource interface for the provider supplied
// subnet. The common subnet object delegates provider specific
// behavior to the raw subnet.
type ProviderSubnet interface {
	Resource
	DynamicSubnet
}
