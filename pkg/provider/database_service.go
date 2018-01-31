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
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/resource"
)

// DatabaseService is an abstract factory. It provides the methods that will
// create the provider resources. Vendor implementations will provide the
// concrete implementations of these methods.
type DatabaseService interface {
	NewDatabaseService(*config.DatabaseService) (resource.DatabaseService, error)
	NewDatabase(*config.Database, resource.DatabaseService) (resource.Database, error)
}

// DatabaseServiceCtor is the function signature for the provider's database service constructor.
type DatabaseServiceCtor func(*config.DatabaseService) (DatabaseService, error)

var dbsCtors map[string]DatabaseServiceCtor = map[string]DatabaseServiceCtor{}

// RegisterDatabaseService is used by a provider implementation to make the provider package
// (i.e. pkg/aws or pkg/mock) available to the arc package. This function is called in the
// packages' init() function.
func RegisterDatabaseService(vendor string, ctor DatabaseServiceCtor) {
	dbsCtors[vendor] = ctor
}

// NewDatabaseService is the provider agnostic constructor used by pkg/arc.
func NewDatabaseService(cfg *config.DatabaseService) (DatabaseService, error) {
	ctor := dbsCtors[cfg.Provider.Vendor]
	if ctor == nil {
		return nil, fmt.Errorf("Unknown vendor %q", vendor)
	}
	return ctor(cfg)
}
