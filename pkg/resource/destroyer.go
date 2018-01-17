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

// Destroyer provides the ability to destroy the resource with the cloud provider and to see
// if the resource has been destroyed.
type Destroyer interface {

	// Destroy asks the provider to deallocate this resource.
	Destroy(flags ...string) error

	// Destroyed indicated that the underlying resource has not been created
	// with the cloud provider. With a composite resource, destruction may
	// only be true if all the contained resources have been destroyed.
	// Destroyed() bool
}

// DestroyOverride allows the destroy methods of the class to be overridden by a derived class.
type DestroyOverride interface {

	// PreDestroy executes before the object being destroyed with the cloud provider.
	PreDestroy(flags ...string) error

	// MidDestroy asks the provider to deallocate the resource.
	MidDestroy(flags ...string) error

	// PostDestroy executes after to the object being destroyed with the cloud provider.
	PostDestroy(flags ...string) error
}
