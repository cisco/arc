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

// StaticKeyPair provides the interface to the static portion of the
// keypair.
type StaticKeyPair interface {
	Name() string
	LocalName() string
	Format() string
	Comment() string
	KeyMaterial() string
}

// DynamicKeyPair provides access to the dynamic portion of the keypair.
type DynamicKeyPair interface {

	// Loader is the interface that requires the Load function to be implmented.
	Loader

	// FingerPrint returns the finger print of the keypair from the provider.
	FingerPrint() string
}

// KeyPair provides the resource interface used for the common keypair
// object implemented in the arc package.
type KeyPair interface {
	Resource
	StaticKeyPair
	DynamicKeyPair
}

// ProviderKeyPair provides a resource interface for the provider supplied
// keypair. The common keypair object delegates provider specific
// behavior to the raw keypair.
type ProviderKeyPair interface {
	Resource
	DynamicKeyPair
}
