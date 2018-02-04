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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

type database struct {
	*config.Database
	rds *rds.RDS
	dbs *databaseService
	// subnetGroup *dbSubnetGroup

	db *rds.DBInstance
	id string
}

func newDatabase(cfg *config.Database, d resource.ProviderDatabaseService, p *databaseServiceProvider) (resource.ProviderDatabase, error) {
	dbs, ok := d.(*databaseService)
	if !ok {
		return nil, fmt.Errorf("Internal Error: aws/database.go, type assert for ProviderDatabaseService parameter failed.")
	}

	db := &database{
		Database: cfg,
		rds:      p.rds,
		dbs:      dbs,
	}

	/*
		s, err := newDbSubGroup(dbs, cfg.SubnetGroup())
		if err != nil {
			return nil, err
		}
		db.subnetGroup = s
	*/

	return db, nil
}

func (db *database) set(dbi *rds.DBInstance) {
	if dbi == nil || dbi.DBInstanceIdentifier == nil {
		return
	}
	db.db = dbi
	db.id = *dbi.DBInstanceIdentifier
}

func (db *database) Load() error {
	db.set(db.dbs.databaseCache.find(db))
	return nil
}

func (db *database) Create(flags ...string) error {
	msg.Info("Database Creation: %s %s", db.Name())
	if db.Created() {
		msg.Detail("Database exists, skipping...")
		return nil
	}

	if db.dbs.network == nil {
		return fmt.Errorf("Network not associate with database service.")
	}

	param := &rds.CreateDBInstanceInput{
		CopyTagsToSnapshot:   aws.Bool(true),
		DBInstanceClass:      aws.String(db.InstanceType()),
		DBInstanceIdentifier: aws.String(db.Name()),
		DBName:               aws.String(db.Name()),
		// DBSubnetGroupName:    aws.String(db.subnetGroup),
		Engine:           aws.String(db.Engine()),
		MultiAZ:          aws.Bool(true),
		StorageEncrypted: aws.Bool(true),
	}
	if db.Version() != "" {
		param.EngineVersion = aws.String(db.Version())
	}
	if db.StorageIops() > 0 {
		param.Iops = aws.Int64(int64(db.StorageIops()))
	}
	if db.MasterUserName() != "" {
		param.MasterUsername = aws.String(db.MasterUserName())
	}
	if db.MasterPassword() != "" {
		param.MasterUserPassword = aws.String(db.MasterPassword())
	}
	if db.Port() > 0 {
		param.Port = aws.Int64(int64(db.Port()))
	}
	if db.StorageType() != "" {
		if db.StorageIops() > 0 {
			param.StorageType = aws.String("io1")
		} else {
			param.StorageType = aws.String(db.StorageType())
		}
	}

	// VpcSecurityGroupIds

	return db.Load()
}

func (db *database) Created() bool {
	return db.db != nil
}

func (db *database) Destroy(flags ...string) error {
	return nil
}

func (db *database) Provision(flags ...string) error {
	return nil
}

func (db *database) Destroyed() bool {
	return db.db == nil
}

func (db *database) Audit(flags ...string) error {
	// TODO
	return nil
}

func (db *database) Info() {
}

func (db *database) Id() string {
	return db.id
}
