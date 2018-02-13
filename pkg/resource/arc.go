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

// StaticArc provides the interface to the static portion of the
// arc resource tree. This information is provided via config file
// and is implemented config.Arc.
type StaticArc interface {
	Name() string
	Title() string
}

// Arc provides the resource interface used for the common arc object
// implemented in the arc package. It contains an Run method used to
// start application processing. It also contains DataCenter and Dns
// methods used to access it's children.
type Arc interface {
	Resource
	StaticArc

	// Run is the entry point for arc.
	Run() (int, error)

	// DataCenter provides access to Arc's child datacenter service.
	DataCenter() DataCenter

	// DatabaseService provides access to Arc's child database service.
	DatabaseService() DatabaseService

	// ContainerService provides access to Arc's child container service.
	ContainerService() ContainerService

	// Dns provides access to Arc's child dns service.
	Dns() Dns
}
