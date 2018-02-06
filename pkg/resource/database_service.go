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

// StaticDatabaseService provides the interface to the static portion of the
// database service. This information is provided via config file and is implemented
// by config.DatabaseService.
type StaticDatabaseService interface {
	config.Printer
}

// DynamicDatabaseService provides access to the dynamic portion of the database service.
type DynamicDatabaseService interface {
	Loader
	Auditor
	Informer
}

// ProviderDatabaseService provides a resource interface for the provider supplied database service.
type ProviderDatabaseService interface {
	DynamicDatabaseService
}

// DatabaseService provides the resource interface used for the database service
// object implemented in the arc package.
type DatabaseService interface {
	route.Router
	StaticDatabaseService
	DynamicDatabaseService
	Provisioner
	Helper

	// Arc provides access to DataCenter's parent.
	Arc() Arc

	// Find returns the database with the given name.
	Find(string) Database

	// ProviderDatabaseSerivces allows access to the provider's database service object.
	ProviderDatabaseService() ProviderDatabaseService
}
