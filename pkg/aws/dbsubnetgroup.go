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

type dbSubnetGroup struct {
	rds     *rds.RDS
	dbs     *databaseService
	subnets []*subnet
	sg      *rds.DBSubnetGroup
	name_   string
}

func newDBSubnetGroup(cfg *config.Database, params resource.DatabaseParams, p *databaseServiceProvider) (*dbSubnetGroup, error) {
	name := cfg.Name() + "-subnet-group"
	log.Debug("Initializing AWS DBSubnetGroup %q", name)

	sg := &dbSubnetGroup{
		rds:   p.rds,
		name_: name,
	}

	for _, sub := range params.Subnets {
		s, ok := sub.(*subnet)
		if !ok {
			return nil, fmt.Errorf("Internal Error: aws/database.go, type assert for Subnets parameter failed.")
		}
		sg.subnets = append(sg.subnets, s)
	}
	if sg.subnets == nil {
		return nil, fmt.Errorf("Creating AWS DBSubnetGroup for database %s, but no subnets found.", cfg.Name())
	}

	return sg, nil
}

func (sg *dbSubnetGroup) set(s *rds.DBSubnetGroup) {
	if s == nil || s.DBSubnetGroupName == nil {
		return
	}
	sg.sg = s
}

func (sg *dbSubnetGroup) clear() {
	sg.sg = nil
}

func (sg *dbSubnetGroup) load() error {
	sg.set(sg.dbs.dbSubnetGroupCache.find(sg))
	return nil
}

func (sg *dbSubnetGroup) create() error {
	msg.Info("DBSubnetGroup Creation: %s", sg.name())
	if sg.created() {
		msg.Detail("DBSubnetGroup exists, skipping...")
		return nil
	}

	subnetIds := []*string{}
	for _, s := range sg.subnets {
		subnetIds = append(subnetIds, aws.String(s.Id()))
	}

	params := &rds.CreateDBSubnetGroupInput{
		DBSubnetGroupDescription: aws.String(sg.subnets[0].GroupName()),
		DBSubnetGroupName:        aws.String(sg.subnets[0].GroupName()),
		SubnetIds:                subnetIds,
	}
	resp, err := sg.rds.CreateDBSubnetGroup(params)
	if err != nil {
		return err
	}
	sg.set(resp.DBSubnetGroup)
	sg.dbs.dbSubnetGroupCache.add(sg)

	return nil
}

func (sg *dbSubnetGroup) created() bool {
	return sg.sg != nil
}

func (sg *dbSubnetGroup) destroy(flags ...string) error {
	msg.Info("DBSubnetGroup Destruction: %s", sg.name())
	if sg.destroyed() {
		msg.Detail("DBSubnetGroup does not exist, skipping...")
		return nil
	}

	params := &rds.DeleteDBSubnetGroupInput{
		DBSubnetGroupName: aws.String(sg.name()),
	}
	if _, err := sg.rds.DeleteDBSubnetGroup(params); err != nil {
		return err
	}
	sg.clear()
	sg.dbs.dbSubnetGroupCache.remove(sg)

	return nil
}

func (sg *dbSubnetGroup) destroyed() bool {
	return sg.sg == nil
}

func (sg *dbSubnetGroup) info() {
	if sg.destroyed() {
		return
	}
	msg.Info("AWS DBSubnetGroup: %s", sg.name())
	if sg.sg.DBSubnetGroupName != nil {
		msg.Detail("%-20s\t%s", "name", *sg.sg.DBSubnetGroupName)
	}
	if sg.sg.DBSubnetGroupDescription != nil {
		msg.Detail("%-20s\t%s", "description", *sg.sg.DBSubnetGroupDescription)
	}
	if sg.sg.DBSubnetGroupArn != nil {
		msg.Detail("%-20s\t%s", "arn", *sg.sg.DBSubnetGroupArn)
	}
	if sg.sg.SubnetGroupStatus != nil {
		msg.Detail("%-20s\t%s", "status", *sg.sg.SubnetGroupStatus)
	}
	if sg.sg.VpcId != nil {
		msg.Detail("%-20s\t%s", "vpc", *sg.sg.VpcId)
	}
	msg.IndentInc()
	for _, s := range sg.sg.Subnets {
		if s != nil && s.SubnetIdentifier != nil {
			msg.Detail("%-20s\t%s", "subnet", *s.SubnetIdentifier)
		}
	}
	msg.IndentDec()
}

func (sg *dbSubnetGroup) name() string {
	return sg.name_
}
