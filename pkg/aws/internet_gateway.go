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

type internetGateway struct {
	ec2     *ec2.EC2
	network *network
	name_   string
	id_     string
	igw     *ec2.InternetGateway
}

func newInternetGateway(c *ec2.EC2, n *network, name string) (*internetGateway, error) {
	log.Debug("Initializing AWS InternetGateway %q", name)

	i := &internetGateway{
		ec2:     c,
		network: n,
		name_:   name,
	}
	return i, nil
}

func (i *internetGateway) Route(req *route.Request) route.Response {
	log.Route(req, "AWS InternetGateway %q", i.name())

	if req.Top() != "" {
		i.help()
		return route.FAIL
	}

	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load:
		if err := i.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Create:
		return i.create(req)
	case route.Audit:
		return route.OK
	case route.Destroy:
		return i.destroy(req)
	case route.Help:
		i.help()
		return route.OK
	case route.Info:
		i.info()
		return route.OK
	}
	return route.FAIL
}

func (i *internetGateway) Created() bool {
	return i.igw != nil
}

func (i *internetGateway) Destroyed() bool {
	return !i.Created()
}

func (i *internetGateway) name() string {
	return i.name_
}

func (i *internetGateway) id() string {
	return i.id_
}

func (i *internetGateway) set(igw *ec2.InternetGateway) {
	if igw == nil || igw.InternetGatewayId == nil {
		return
	}
	i.igw = igw
	i.id_ = *igw.InternetGatewayId
}

func (i *internetGateway) clear() {
	i.igw = nil
	i.id_ = ""
}

func (i *internetGateway) Load() error {
	i.igw = nil

	params := &ec2.DescribeInternetGatewaysInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("attachment.vpc-id"),
				Values: []*string{
					aws.String(i.network.vpc.id()),
				},
			},
		},
	}
	if i.id() != "" {
		params.Filters = append(params.Filters, &ec2.Filter{
			Name: aws.String("internet-gateway-id"),
			Values: []*string{
				aws.String(i.id()),
			},
		})
	} else {
		params.Filters = append(params.Filters, &ec2.Filter{
			Name: aws.String("tag:Name"),
			Values: []*string{
				aws.String(i.name()),
			},
		})
	}
	resp, err := i.ec2.DescribeInternetGateways(params)
	if err != nil {
		return err
	}

	for _, igw := range resp.InternetGateways {
		if i.id() != "" && igw.InternetGatewayId != nil {
			if i.id() == *igw.InternetGatewayId {
				i.set(igw)
				return nil
			}
		}
		if igw.Tags == nil {
			break
		}
		for _, tag := range igw.Tags {
			if tag != nil && tag.Key != nil && *tag.Key == "Name" && tag.Value != nil && *tag.Value == i.name() {
				i.set(igw)
				return nil
			}
		}
	}
	return nil
}

func (i *internetGateway) reload() bool {
	return msg.Wait(
		fmt.Sprintf("Waiting for InternetGateway %s, %s to become available", i.name(), i.id()), // title
		fmt.Sprintf("InternetGateway %s, %s never became available", i.name(), i.id()),          // err
		60,        // duration
		i.Created, // test()
		func() bool { return i.Load() == nil }, // load()
	)
}

func (i *internetGateway) create(req *route.Request) route.Response {
	msg.Info("InternetGateway Creation: %s", i.name())
	if i.Created() {
		msg.Detail("InternetGateway exists, skipping...")
		return route.OK
	}
	resp, err := i.ec2.CreateInternetGateway(&ec2.CreateInternetGatewayInput{})
	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	i.set(resp.InternetGateway)
	if !i.reload() {
		return route.FAIL
	}

	if err := createTags(i.ec2, i.name(), i.id(), req); err != nil {
		msg.Error(err.Error())
		i.destroy(req)
		return route.FAIL
	}

	msg.Detail("Created %s", i.id())
	aaa.Accounting("InternetGateway created: %s", i.id())

	if resp := i.attach(req, i.network.vpc); resp != route.OK {
		return resp
	}

	for _, routeTable := range i.network.routeTables.routeTables {
		if routeTable.access() == "public" || routeTable.access() == "public_elastic" {
			if resp := routeTable.createRoute(req, "0.0.0.0/0", igw, i); resp != route.OK {
				return resp
			}
		}
	}

	return route.OK
}

func (i *internetGateway) destroy(req *route.Request) route.Response {
	msg.Info("InternetGateway Destruction: %s", i.name())
	if i.Destroyed() {
		msg.Detail("InternetGateway does not exist, skipping...")
		return route.OK
	}

	for _, routeTable := range i.network.routeTables.routeTables {
		if routeTable.access() == "public" || routeTable.access() == "public_elastic" {
			if resp := routeTable.deleteRoute(req, "0.0.0.0/0"); resp != route.OK {
				return resp
			}
		}
	}

	if resp := i.detach(req, i.network.vpc); resp != route.OK {
		return resp
	}

	params := &ec2.DeleteInternetGatewayInput{
		InternetGatewayId: aws.String(i.id()),
	}
	if _, err := i.ec2.DeleteInternetGateway(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	msg.Detail("Destroyed %s", i.id())
	aaa.Accounting("InternetGateway destroyed: %s", i.id())

	i.clear()

	return route.OK
}

func (i *internetGateway) help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: "create internet gateway"},
		{Name: route.Destroy.String(), Desc: "destroy internet gateway"},
		{Name: route.Info.String(), Desc: "show information about allocated internet gateway"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("network internetgateway", commands)
}

func (i *internetGateway) info() {
	if i.Destroyed() {
		return
	}
	msg.Info("InternetGateway")
	msg.Detail("%-20s\t%s", "name", i.name())
	msg.Detail("%-20s\t%s", "id", i.id())
	msg.IndentInc()
	for _, igw := range i.igw.Attachments {
		msg.Info("Attachment")
		if igw.VpcId != nil {
			msg.Detail("%-20s\t%s", "vpc", *igw.VpcId)
		}
		if igw.State != nil {
			msg.Detail("%-20s\t%s", "state", *igw.State)
		}
	}
	msg.IndentDec()
	printTags(i.igw.Tags)
}

func (i *internetGateway) attached(v *vpc) bool {
	for _, igw := range i.igw.Attachments {
		if igw.VpcId != nil {
			if *igw.VpcId == v.id() {
				return true
			}
		}
	}
	return false
}

func (i *internetGateway) attach(req *route.Request, v *vpc) route.Response {
	if v.Destroyed() || i.Destroyed() {
		return route.OK
	}
	msg.Info("Attach Internet Gateway: %s to Vpc: %s", i.id(), v.id())
	if i.attached(v) {
		msg.Detail("Gateway already attached. Skipping...")
		return route.OK
	}
	params := &ec2.AttachInternetGatewayInput{
		InternetGatewayId: aws.String(i.id()),
		VpcId:             aws.String(v.id()),
	}
	if _, err := i.ec2.AttachInternetGateway(params); err != nil {
		msg.Warn(err.Error())
	}
	aaa.Accounting("InternetGateway %s attached to Vpc %s", i.id(), v.id())
	return route.OK
}

func (i *internetGateway) detach(req *route.Request, v *vpc) route.Response {
	if !v.Created() || !i.Created() {
		return route.OK
	}
	msg.Info("Detach Internet Gateway %s from Vpc: %s", i.id(), v.id())
	if !i.attached(v) {
		msg.Detail("Gateway already detached, skipping...")
		return route.OK
	}
	params := &ec2.DetachInternetGatewayInput{
		InternetGatewayId: aws.String(i.id()),
		VpcId:             aws.String(v.id()),
	}
	if _, err := i.ec2.DetachInternetGateway(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	aaa.Accounting("InternetGateway %s detached from Vpc %s", i.id(), v.id())
	return route.OK
}
