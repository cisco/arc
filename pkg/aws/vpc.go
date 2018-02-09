//
// Copyright (c) 2017, Cisco Systems
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
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/route"
)

type vpc struct {
	ec2     *ec2.EC2
	network *network
	name_   string

	vpc *ec2.Vpc
	id_ string
}

func newVpc(c *ec2.EC2, n *network) (*vpc, error) {
	log.Debug("Initializing AWS Vpc")

	v := &vpc{
		ec2:     c,
		network: n,
		name_:   "vpc-" + n.Name(),
	}
	return v, nil
}

func (v *vpc) Route(req *route.Request) route.Response {
	log.Route(req, "AWS Vpc")

	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load:
		if err := v.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Create:
		return v.create(req)
	case route.Audit:
		return route.OK
	case route.Destroy:
		return v.destroy(req)
	case route.Help:
		v.help()
		return route.OK
	case route.Info:
		v.info()
		return route.OK
	}
	return route.FAIL
}

func (v *vpc) Created() bool {
	return v.vpc != nil
}

func (v *vpc) Destroyed() bool {
	return !v.Created()
}

func (v *vpc) name() string {
	return v.name_
}

func (v *vpc) id() string {
	return v.id_
}

func (v *vpc) state() string {
	if v.vpc == nil || v.vpc.State == nil {
		return ""
	}
	return *v.vpc.State
}

func (v *vpc) cidrBlock() string {
	if v.vpc == nil || v.vpc.CidrBlock == nil {
		return ""
	}
	return *v.vpc.CidrBlock
}

func (v *vpc) set(vpc *ec2.Vpc) {
	if vpc == nil || vpc.VpcId == nil {
		return
	}
	v.vpc = vpc
	v.id_ = *vpc.VpcId
}

func (v *vpc) clear() {
	v.vpc = nil
	v.id_ = ""
}

func (v *vpc) Load() error {
	v.vpc = nil

	params := &ec2.DescribeVpcsInput{
		Filters: []*ec2.Filter{},
	}
	if v.id() != "" {
		params.Filters = append(params.Filters, &ec2.Filter{
			Name: aws.String("vpc-id"),
			Values: []*string{
				aws.String(v.id()),
			},
		})
	} else if v.name() != "" {
		params.Filters = append(params.Filters, &ec2.Filter{
			Name: aws.String("tag:Name"),
			Values: []*string{
				aws.String(v.name()),
			},
		})
	}
	if len(params.Filters) == 0 {
		return fmt.Errorf("Internal error, cannot load vpc. name or id not present")
	}
	resp, err := v.ec2.DescribeVpcs(params)
	if err != nil {
		return err
	}
	for _, vpc := range resp.Vpcs {
		if v.id() != "" && vpc.VpcId != nil {
			if v.id() == *vpc.VpcId {
				v.set(vpc)
				return err
			}
		}
		if vpc.Tags == nil {
			break
		}
		for _, tag := range vpc.Tags {
			if tag != nil && tag.Key != nil && *tag.Key == "Name" && tag.Value != nil && *tag.Value == v.name() {
				v.set(vpc)
				return nil
			}
		}
	}
	return nil
}

func (v *vpc) reload() bool {
	return msg.Wait(
		fmt.Sprintf("Waiting for Vpc %s to become available", v.id()), //title
		fmt.Sprintf("Vpc %s never became available", v.id()),          // err
		600, // duration
		func() bool { return v.state() == "available" }, // test()
		func() bool { // load()
			if err := v.Load(); err != nil {
				msg.Error(err.Error())
				return false
			}
			return true
		},
	)
}

func (v *vpc) create(req *route.Request) route.Response {
	msg.Info("Vpc Creation: %s", v.name())
	if v.Created() {
		msg.Detail("Vpc exists, skipping...")
		return route.OK
	}

	params := &ec2.CreateVpcInput{
		CidrBlock: aws.String(v.network.CidrBlock()),
	}
	resp, err := v.ec2.CreateVpc(params)
	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	v.set(resp.Vpc)
	if !v.reload() {
		return route.FAIL
	}

	if err := createTags(v.ec2, v.name(), v.id(), req); err != nil {
		msg.Error(err.Error())
		v.destroy(req)
		return route.FAIL
	}

	msg.Detail("Created %s", v.id())
	aaa.Accounting("Vpc created: %s", v.id())
	return route.OK
}

func (v *vpc) destroy(req *route.Request) route.Response {
	msg.Info("Vpc Destruction: %s", v.name())
	if v.Destroyed() {
		msg.Detail("Vpc does not exist, skipping...")
		return route.OK
	}

	params := &ec2.DeleteVpcInput{
		VpcId: v.vpc.VpcId,
	}
	if _, err := v.ec2.DeleteVpc(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	msg.Detail("Destroyed %s", v.id())
	aaa.Accounting("Vpc destroyed: %s", v.id())

	v.clear()

	return route.OK
}

func (v *vpc) help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: "create vpc"},
		{Name: route.Destroy.String(), Desc: "destroy vpc"},
		{Name: route.Info.String(), Desc: "show information about allocated vpc"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("network vpc", commands)
}

func (v *vpc) info() {
	if v.Destroyed() {
		return
	}
	msg.Info("Vpc")
	msg.Detail("%-20s\t%s", "name", v.name())
	msg.Detail("%-20s\t%s", "id", v.id())
	msg.Detail("%-20s\t%s", "state", v.state())
	msg.Detail("%-20s\t%s", "cidr", v.cidrBlock())
	printTags(v.vpc.Tags)
}
