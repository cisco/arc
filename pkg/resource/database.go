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

// StaticDatabase provides the interface to the static portion of the
// database. This information is provided via config file and is implemented
// by config.Database.
type StaticDatabase interface {
	config.Printer

	// Name of the configured database instance. Required.
	Name() string

	// Engine used by the database instance. Required.
	Engine() string

	// Version of the engine. Optional.
	Version() string

	// InstanceType of the database instance. Required.
	InstanceType() string

	// Port used by the database instance. Optional.
	Port() int

	// Subnet the database instance will use. Required.
	SubnetGroup() string

	// SecurityGroups the database instance will use. Required.
	SecurityGroups() []string

	// StorageType associated with the database instance. Optional.
	StorageType() string

	// StorageSize is the configured size of the storage attached to the database instance. Optional.
	StorageSize() int

	// StorageIops is the configured tops of the storage attached to the database instance. Optional.
	StorageIops() int

	// MasterUserName is the name for the master user. Optional
	MasterUserName() string

	// MasterPassword is the password for the master user. Optional.
	MasterPassword() string
}

// DynamicDatabase provides access to the dynamic portion of the database.
type DynamicDatabase interface {
	Loader
	Creator
	Destroyer
	Provisioner
	Auditor
	Informer

	// Id returns the id of the instance.
	Id() string
}

// Database provides the resource interface used for the common subnet group
// object implemented in the arc package.
type Database interface {
	route.Router
	StaticDatabase
	DynamicDatabase

	// ProviderDatabase provides access tot he provider specific database.
	ProviderDatabase() ProviderDatabase

	// DatabaseService provides access to the database's parent.
	DataBaseService() DatabaseService
}

// ProviderDatabase provides a resource interface for the provider supplied database instance.
type ProviderDatabase interface {
	DynamicDatabase
}
