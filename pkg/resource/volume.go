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

// StaticVolume provides the interface to the static portion of the
// volume.
type StaticVolume interface {
	Device() string
	Size() int64
	Type() string
	Keep() bool
	Boot() bool
	Preserve() bool
	FsType() string
	Inodes() int
	MountPoint() string
}

// DynamicVolume provides the interface to the dynamic portion of the volume.
type DynamicVolume interface {
	Loader
	Attacher
	Detacher
	Auditor

	// Id returns the underlying volume id.
	Id() string

	// State returns the state of the volume.
	State() string

	// Destroys the underlying volume. Detach must be called before this
	// otherwise this could return an error.
	Destroy() error

	// SetTags sets the tags for the volume.
	SetTags(map[string]string) error

	// Info prints out volume information to the console.
	Info()

	// Reset the volume to an initialized state.
	Reset()
}

// Volume provides the resource interface used for the common volume
// object implemented in the arc package.
type Volume interface {
	Resource
	StaticVolume
	DynamicVolume

	ProviderVolume() ProviderVolume
}

// ProviderVolume provides a resource interface for the provider supplied
// volume. The common volume object delegates provider specific
// behavior to the raw volume.
type ProviderVolume interface {
	Resource
	DynamicVolume
}
