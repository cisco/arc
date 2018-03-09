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

// IdentityManagement is an abstract factory. It provides the methods that will
// create the provider resources. Vendor implementations will provide the
// concrete implementations of these methods.
type IdentityManagement interface {
	NewIdentityManagement(cfg *config.IdentityManagement) (resource.ProviderIdentityManagement, error)
	NewPolicy(pol resource.Policy, cfg *config.Policy) (resource.ProviderPolicy, error)
}

// IdentityManagementCtor is the function signature for the provider's identityManagement constructor.
type IdentityManagementCtor func(*config.Amp) (IdentityManagement, error)

var identityManagementCtors map[string]IdentityManagementCtor = map[string]IdentityManagementCtor{}

// RegisterIdentityManagement is used by a provider implementation to make the provider package
// (i.e. pkg/aws or pkg/mock) available to the amp package. This function is called in the
// packages' init() function.
func RegisterIdentityManagement(vendor string, ctor IdentityManagementCtor) {
	identityManagementCtors[vendor] = ctor
}

// NewIdentityManagement is the provider agnostic constructor used by pkg/amp.
func NewIdentityManagement(cfg *config.Amp) (IdentityManagement, error) {
	vendor := cfg.Provider.Vendor
	ctor := identityManagementCtors[vendor]
	if ctor == nil {
		return nil, fmt.Errorf("Unknown vendor %q", vendor)
	}
	return ctor(cfg)
}
