//
// Copyright (c) 2018, Cisco Systems
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

type StaticEncryptionKey interface {
	Name() string
}

// DynamicEncryptionKey provides the interface to the dynamic portion of the encryptionKey.
type DynamicEncryptionKey interface {
	Loader
	Creator
	Destroyer
	Provisioner
	Auditor
	Informer

	// SetTags sets the tags for the bucket.
	SetTags(map[string]string) error
}

// Bucket provides the resource interface used for the common storage
// object implemented in the amp package. It contains an Storage method used to
// access its parent object.
type EncryptionKey interface {
	route.Router
	StaticEncryptionKey
	DynamicEncryptionKey
	Helper

	// KeyManagement provides access to EncryptionKey's parent object.
	KeyManagement() KeyManagement

	//ProviderEncryptionKey provides access to the provider EncryptionKey object.
	ProviderEncryptionKey() ProviderEncryptionKey
}

type ProviderEncryptionKey interface {
	DynamicEncryptionKey
}
