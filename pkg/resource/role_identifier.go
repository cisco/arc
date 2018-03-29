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

// StaticRoleIdentifier provides the interface to the static portion of the role identifier.
type StaticRoleIdentifier interface {
	Name() string
}

// DynamicRoleIdentifier provides the interface to the dynamic portion of the role identifier.
type DynamicRoleIdentifier interface {
	Loader
	Attacher
	Detacher

	// Id returns the underlying role id.
	Id() string

	// InstanceId returns the id of the role's instance.
	InstanceId() string

	// Update changes the role to be what is currently in the config file.
	Update() error
}

// RoleIdentifier provides the resource interface used for the common role identifier
// object implemented in the arc package.
type RoleIdentifier interface {
	Resource
	StaticRoleIdentifier
	DynamicRoleIdentifier

	ProviderRoleIdentifier() ProviderRoleIdentifier
}

// ProviderRoleIdentifier provides a resource interface for the provider supplied
// role identifier. The common role identifier object delegates provider specific
// behavior to the raw role.
type ProviderRoleIdentifier interface {
	Resource
	DynamicRoleIdentifier
}
