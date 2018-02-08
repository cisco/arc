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
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

type database struct {
	*config.Database
	rds         *rds.RDS
	dbs         *databaseService
	subnetGroup *dbSubnetGroup
	secgroups   []*string
	db          *rds.DBInstance
}

func newDatabase(cfg *config.Database, params resource.DatabaseParams, p *databaseServiceProvider) (resource.ProviderDatabase, error) {
	log.Debug("Initializing AWS Database Instance %q", cfg.Name())

	db := &database{
		Database: cfg,
		rds:      p.rds,
	}

	dbs, ok := params.DatabaseService.(*databaseService)
	if !ok {
		return nil, fmt.Errorf("Internal Error: aws/database.go, type assert for DatabaseService parameter failed.")
	}
	db.dbs = dbs

	for _, secgroup := range params.SecurityGroups {
		s, ok := secgroup.(*securityGroup)
		if !ok {
			return nil, fmt.Errorf("Internal Error: aws/database.go, type assert for SecurityGroups parameter failed.")
		}
		db.secgroups = append(db.secgroups, aws.String(s.Id()))
	}

	dbsg, err := newDBSubnetGroup(cfg, params, p)
	if err != nil {
		return nil, err
	}
	db.subnetGroup = dbsg

	return db, nil
}

func (db *database) set(dbi *rds.DBInstance) {
	if dbi == nil || dbi.DBInstanceIdentifier == nil {
		return
	}
	db.db = dbi
}

func (db *database) clear() {
	db.db = nil
}

func (db *database) Load() error {
	db.set(db.dbs.databaseCache.find(db))
	if err := db.subnetGroup.load(); err != nil {
		return err
	}
	return nil
}

func (db *database) Create(flags ...string) error {
	if err := db.subnetGroup.create(); err != nil {
		return err
	}

	params := &rds.CreateDBInstanceInput{
		DBInstanceClass:      aws.String(db.InstanceType()),
		DBInstanceIdentifier: aws.String(db.Name()),
		DBName:               aws.String(db.Name()),
		DBSubnetGroupName:    aws.String(db.subnetGroup.name()),
		Engine:               aws.String(db.Engine()),
		MultiAZ:              aws.Bool(true),
		PubliclyAccessible:   aws.Bool(false),
		StorageEncrypted:     aws.Bool(true),
		VpcSecurityGroupIds:  db.secgroups,
	}
	if db.DBName() != "" {
		params.DBName = aws.String(db.DBName())
	}
	if db.Version() != "" {
		params.EngineVersion = aws.String(db.Version())
	}
	if db.StorageIops() > 0 {
		params.Iops = aws.Int64(int64(db.StorageIops()))
	}
	if db.MasterUserName() != "" {
		params.MasterUsername = aws.String(db.MasterUserName())
	}
	if db.MasterPassword() != "" {
		params.MasterUserPassword = aws.String(db.MasterPassword())
	}
	if db.Port() > 0 {
		params.Port = aws.Int64(int64(db.Port()))
	}
	if db.StorageType() != "" {
		if db.StorageSize() > 0 {
			params.AllocatedStorage = aws.Int64(int64(db.StorageSize()))
		}
		if db.StorageIops() > 0 {
			params.StorageType = aws.String("io1")
			params.Iops = aws.Int64(int64(db.StorageIops()))
		} else {
			params.StorageType = aws.String(db.StorageType())
		}
	}
	resp, err := db.rds.CreateDBInstance(params)
	if err != nil {
		return err
	}
	db.set(resp.DBInstance)
	db.dbs.databaseCache.add(db)

	return nil
}

func (db *database) Created() bool {
	return db.db != nil
}

func (db *database) Destroy(flags ...string) error {
	if db.Destroyed() {
		if err := db.subnetGroup.destroy(); err != nil {
			return err
		}
		return nil
	}

	params := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier: aws.String(db.Name()),
		SkipFinalSnapshot:    aws.Bool(true),
	}
	_, err := db.rds.DeleteDBInstance(params)
	if err != nil {
		return err
	}
	db.clear()
	db.dbs.databaseCache.remove(db)

	if err := db.subnetGroup.destroy(); err != nil {
		return err
	}

	return nil
}

func (db *database) Destroyed() bool {
	return db.db == nil
}

func (db *database) Provision(flags ...string) error {
	// TODO
	return nil
}

func (db *database) Audit(flags ...string) error {
	// TODO
	return nil
}

func (db *database) Info() {
	if db.Destroyed() {
		return
	}
	if db.db.DBInstanceIdentifier != nil {
		msg.Detail("%-20s\t%s", "database", *db.db.DBInstanceIdentifier)
	}
	if db.db.DBName != nil {
		msg.Detail("%-20s\t%s", "db name", *db.db.DBName)
	}
	if db.db.DBInstanceArn != nil {
		msg.Detail("%-20s\t%s", "arn", *db.db.DBInstanceArn)
	}
	if db.db.DbiResourceId != nil {
		msg.Detail("%-20s\t%s", "resource id", *db.db.DbiResourceId)
	}
	if db.db.DBInstanceStatus != nil {
		msg.Detail("%-20s\t%s", "status", *db.db.DBInstanceStatus)
	}
	if db.db.Engine != nil {
		msg.Detail("%-20s\t%s", "engine", *db.db.Engine)
	}
	if db.db.EngineVersion != nil {
		msg.Detail("%-20s\t%s", "version", *db.db.EngineVersion)
	}
	if db.db.DBInstanceClass != nil {
		msg.Detail("%-20s\t%s", "type", *db.db.DBInstanceClass)
	}
	if db.db.DbInstancePort != nil && *db.db.DbInstancePort > 0 {
		msg.Detail("%-20s\t%d", "port", *db.db.DbInstancePort)
	}
	if db.db.InstanceCreateTime != nil {
		msg.Detail("%-20s\t%s", "created", *db.db.InstanceCreateTime)
	}
	if db.db.StorageType != nil {
		msg.Detail("%-20s\t%s", "storage type", *db.db.StorageType)
	}
	if db.db.AllocatedStorage != nil && *db.db.AllocatedStorage > 0 {
		msg.Detail("%-20s\t%d", "storage size", *db.db.AllocatedStorage)
	}
	if db.db.Iops != nil && *db.db.Iops > 0 {
		msg.Detail("%-20s\t%d", "storage iops", *db.db.Iops)
	}
	if db.db.DBSubnetGroup != nil && db.db.DBSubnetGroup.DBSubnetGroupName != nil {
		msg.Detail("%-20s\t%s", "subnet group", *db.db.DBSubnetGroup.DBSubnetGroupName)
	}
	for _, sg := range db.db.VpcSecurityGroups {
		if sg.VpcSecurityGroupId != nil {
			msg.Detail("%-20s\t%s", "security group", *sg.VpcSecurityGroupId)
		}
	}
}

func (db *database) Id() string {
	if db.Destroyed() || db.db.DBInstanceIdentifier == nil {
		return ""
	}
	return *db.db.DBInstanceIdentifier
}
