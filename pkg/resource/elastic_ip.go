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

// DynamicElasticIP provides the interface to the dynamic portion of the elastic IP
type DynamicElasticIP interface {

	// Loader is the interface that requires the Load function to be implmented.
	Loader

	// Attacher provides the ability to attach the resource with the cloud provider and to see
	// if the resource is attached.
	Attacher

	// Detacher provides the ability to detach the resource with the cloud provider and to see
	// if the resource is detached.
	Detacher

	// Id returns the identifier of the elastic IP.
	Id() string

	// Instance returns the instance that the elastic IP is or will be associated with.
	Instance() Instance

	// IpAddress returns the elastic IP address.
	IpAddress() string

	// Create allocates the elastic IP.
	Create() error

	// Destroy releases the elastic IP.
	Destroy() error
}

// ElasticIP provides the resource interface used for the common elastic IP
// object implemented in the arc package.
type ElasticIP interface {
	Resource
	DynamicElasticIP
}

// ProviderElasticIP provides a resource interface for the provider supplied
// elastic IP. The common elastic IP object delegates provider specific
// behavior to the provider elastic IP.
type ProviderElasticIP interface {
	Resource
	DynamicElasticIP
}
