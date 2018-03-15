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

type StaticPolicy interface {
	Name() string
	Description() string
	PolicyDocument() string
	config.Printer
}

// DynamicPolicy provides the interface to the dynamic portion of the policy.
type DynamicPolicy interface {
	Loader
	Creator
	Destroyer
	Auditor
	Informer
}

// Policy provides the resource interface used for the common policy
// object implemented in the amp package. It contains an IdentityManagement method used to
// access its parent object.
type Policy interface {
	route.Router
	StaticPolicy
	DynamicPolicy
	Helper

	// IdentityManagement provides access to Policy's parent object.
	IdentityManagement() IdentityManagement

	// ProviderPolicy provides access to the provider policy object.
	ProviderPolicy() ProviderPolicy
}

type ProviderPolicy interface {
	DynamicPolicy
}
