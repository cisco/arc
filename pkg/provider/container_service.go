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

package provider

import (
	"fmt"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/resource"
)

// ContainerServiceCtor is the function signature for the provider's container service constructor.
type ContainerServiceCtor func(*config.ContainerService) (ContainerService, error)

var csCtors map[string]ContainerServiceCtor = map[string]ContainerServiceCtor{}

// RegisterContainerService is used by a provider implementation to make the provider package
// (i.e. pkg/aws or pkg/mock) available to the arc package. This function is called in the
// packages' init() function.
func RegisterContainerService(vendor string, ctor ContainerServiceCtor) {
	csCtors[vendor] = ctor
}

// ContainerService is an abstract factory. It provides the methods that will
// create the provider resources. Vendor implementations will provide the
// concrete implementations of these methods.
type ContainerService interface {
	NewContainerService(*config.ContainerService) (resource.ProviderContainerService, error)
}

// NewContainerService is the provider agnostic constructor used by pkg/arc.
func NewContainerService(cfg *config.ContainerService) (ContainerService, error) {
	// Validate the config.ContainerService object.
	if cfg.Provider == nil {
		return nil, fmt.Errorf("The provider element is missing from the container_service configuration")
	}
	vendor := cfg.Provider.Vendor
	ctor := csCtors[vendor]
	if ctor == nil {
		return nil, fmt.Errorf("Unknown vendor %q", vendor)
	}
	return ctor(cfg)
}
