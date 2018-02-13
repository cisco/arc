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

import (
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/route"
)

// StaticContainerService provides the interface to the static portion of the
// container service. This information is provided via config file and is implemented
// by config.ContainerService.
type StaticContainerService interface {
	config.Printer

	// Name of the container service.
	Name() string
}

// DynamicContainerService provides access to the dynamic portion of the container service.
type DynamicContainerService interface {
	Loader
	Creator
	Destroyer
	Provisioner
	Auditor
	Informer
}

// ProviderContainerService provides a resource interface for the provider supplied container service.
type ProviderContainerService interface {
	DynamicContainerService
}

// ContainerService provides the resource interface used for the container service
// object implemented in the arc package.
type ContainerService interface {
	route.Router
	StaticContainerService
	DynamicContainerService
	Helper

	// Arc provides access to DataCenter's parent.
	Arc() Arc

	// ProviderContainerSerivces allows access to the provider's container service object.
	ProviderContainerService() ProviderContainerService
}
