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

package config

import "github.com/cisco/arc/pkg/msg"

// Database represents the configuration of a database instance resource.
type Database struct {
	Name_           string   `json:"database"`
	Engine_         string   `json:"engine"`
	Version_        string   `json:"version"`
	Type_           string   `json:"type"`
	Port_           int      `json:"port"`
	SubnetGroup_    string   `json:"subnet_group"`
	SecurityGroups_ []string `json:"security_groups"`
	Storage_        struct {
		Type_ string `json:"type"`
		Size_ int    `json:"size"`
		Iops_ int    `json:"iops"`
	} `json:"storage"`
	Master_ struct {
		UserName_ string `json:"username"`
		Password_ string `json:"password"`
	} `json:"master"`
}

func (db *Database) Name() string {
	return db.Name_
}

func (db *Database) Engine() string {
	return db.Engine_
}

func (db *Database) Version() string {
	return db.Version_
}

func (db *Database) InstanceType() string {
	return db.Type_
}

func (db *Database) Port() int {
	return db.Port_
}

func (db *Database) SubnetGroup() string {
	return db.SubnetGroup_
}

func (db *Database) SecurityGroups() []string {
	return db.SecurityGroups_
}

func (db *Database) StorageType() string {
	return db.Storage_.Type_
}

func (db *Database) StorageSize() int {
	return db.Storage_.Size_
}

func (db *Database) StorageIops() int {
	return db.Storage_.Iops_
}

func (db *Database) MasterUserName() string {
	return db.Master_.UserName_
}

func (db *Database) MasterPassword() string {
	return db.Master_.Password_
}

// PrintLocal provides a user friendly way to view the configuration local to the database object.
func (db *Database) PrintLocal() {
	msg.Info("Database Config")
	msg.Detail("%-20s\t%s", "name", db.Name())
	msg.Detail("%-20s\t%s", "engine", db.Engine())
	if db.Version() != "" {
		msg.Detail("%-20s\t%s", db.Version())
	}
	msg.Detail("%-20s\t%s", "type", db.InstanceType())
	if db.Port() > 0 {
		msg.Detail("%-20s\t%d", "type", db.Port())
	}
	msg.Detail("%-20s\t%s", "subnet", db.SubnetGroup())
	groups, sep := "", ""
	for _, group := range db.SecurityGroups() {
		groups += sep + group
		sep = ", "
	}
	msg.Detail("%-20s\t%s", "security_groups", groups)
	msg.Detail("%-20s\t%s", "storage type", db.StorageType())
	msg.Detail("%-20s\t%s", "storage size", db.StorageSize())
	if db.StorageIops() > 0 {
		msg.Detail("%-20s\t%s", "storage iops", db.StorageIops())
	}
	msg.Detail("%-20s\t%s", "master username", db.MasterUserName())
	msg.Detail("%-20s\t%s", "master password", db.MasterPassword())
}

// Print provides a user friendly way to view a subnet group configuration.
func (db *Database) Print() {
	db.PrintLocal()
}
