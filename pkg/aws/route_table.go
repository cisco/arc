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
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type routeTable struct {
	ec2               *ec2.EC2
	network           *network
	name_             string
	access_           string
	availabilityZone_ string

	routeTable *ec2.RouteTable
	id_        string
}

func newRouteTable(c *ec2.EC2, n *network, name, access, availabilityZone string) (*routeTable, error) {
	log.Debug("Initializing AWS RouteTable %q", name)

	r := &routeTable{
		ec2:               c,
		network:           n,
		name_:             name,
		access_:           access,
		availabilityZone_: availabilityZone,
	}
	return r, nil
}

func (r *routeTable) Route(req *route.Request) route.Response {
	log.Route(req, "AWS RouteTable %q", r.name())

	if req.Top() != "" {
		r.help()
		return route.FAIL
	}

	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load:
		if err := r.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Create:
		return r.create(req)
	case route.Destroy:
		return r.destroy(req)
	case route.Help:
		r.help()
		return route.OK
	case route.Info:
		r.info()
		return route.OK
	}
	return route.FAIL
}

func (r *routeTable) Created() bool {
	return r != nil && r.routeTable != nil
}

func (r *routeTable) Destroyed() bool {
	return !r.Created()
}

func (r *routeTable) name() string {
	return r.name_
}

func (r *routeTable) access() string {
	return r.access_
}

func (r *routeTable) availabilityZone() string {
	return r.availabilityZone_
}

func (r *routeTable) id() string {
	return r.id_
}

func (r *routeTable) vpcId() string {
	if r.routeTable == nil || r.routeTable.VpcId == nil {
		return ""
	}
	return *r.routeTable.VpcId
}

func (r *routeTable) set(routeTable *ec2.RouteTable) {
	if routeTable == nil || routeTable.RouteTableId == nil {
		msg.Info("set fail")
		return
	}
	r.routeTable = routeTable
	r.id_ = *r.routeTable.RouteTableId
}

func (r *routeTable) clear() {
	r.routeTable = nil
	r.id_ = ""
}

func (r *routeTable) Load() error {
	r.routeTable = nil

	params := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(r.network.vpc.id()),
				},
			},
		},
	}
	if r.id() != "" {
		params.Filters = append(params.Filters, &ec2.Filter{
			Name: aws.String("route-table-id"),
			Values: []*string{
				aws.String(r.id()),
			},
		})
	} else {
		params.Filters = append(params.Filters, &ec2.Filter{
			Name: aws.String("tag:Name"),
			Values: []*string{
				aws.String(r.name()),
			},
		})
	}
	resp, err := r.ec2.DescribeRouteTables(params)
	if err != nil {
		return err
	}

	for _, routeTable := range resp.RouteTables {
		if r.id() != "" && routeTable.RouteTableId != nil {
			if r.id() == *routeTable.RouteTableId {
				r.set(routeTable)
				return nil
			}
		}
		if routeTable.Tags == nil {
			break
		}
		for _, tag := range routeTable.Tags {
			if tag != nil && tag.Key != nil && *tag.Key == "Name" && tag.Value != nil && *tag.Value == r.name() {
				r.set(routeTable)
				return nil
			}
		}
	}
	return nil
}

func (r *routeTable) reload() bool {
	return msg.Wait(
		fmt.Sprintf("Waiting for RouteTable %s, %s to become available", r.name(), r.id()), // title
		fmt.Sprintf("RouteTable %s, %s never became available", r.name(), r.id()),          // err
		60,        // duration
		r.Created, // test()
		func() bool { //load()
			if err := r.Load(); err != nil {
				msg.Error(err.Error())
				return false
			}
			return true
		},
	)
}

func (r *routeTable) create(req *route.Request) route.Response {
	msg.Info("RouteTable Creation: %s", r.name())
	if r.Created() {
		msg.Detail("RouteTable exists, skipping...")
		return route.OK
	}

	params := &ec2.CreateRouteTableInput{
		VpcId: aws.String(r.network.vpc.id()),
	}

	resp, err := r.ec2.CreateRouteTable(params)
	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	r.set(resp.RouteTable)
	if !r.reload() {
		return route.FAIL
	}

	if err := createTags(r.ec2, r.name(), r.id(), req); err != nil {
		msg.Error(err.Error())
		r.destroy(req)
		return route.FAIL
	}

	msg.Detail("Created %s", r.id())
	aaa.Accounting("RouteTable created: %s", r.id())
	return route.OK
}

func (r *routeTable) destroy(req *route.Request) route.Response {
	msg.Info("RouteTable Destruction: %s", r.name())
	if r.Destroyed() {
		msg.Detail("RouteTable does not exist, skipping...")
		return route.OK
	}

	params := &ec2.DeleteRouteTableInput{
		RouteTableId: r.routeTable.RouteTableId,
	}
	if _, err := r.ec2.DeleteRouteTable(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	msg.Detail("Destroyed %s", r.id())
	aaa.Accounting("RouteTable destroyed: %s", r.id())

	r.clear()

	return route.OK
}

func (r *routeTable) help() {
	commands := []help.Command{
		{route.Create.String(), fmt.Sprintf("create %s routetable", r.name())},
		{route.Destroy.String(), fmt.Sprintf("destroy %s routetable", r.name())},
		{route.Info.String(), fmt.Sprintf("show information about allocated %s routetable", r.name())},
		{route.Help.String(), "show this help"},
	}
	help.Print(fmt.Sprintf("network routetable %s", r.name()), commands)
}

func (r *routeTable) info() {
	if r.Destroyed() {
		return
	}
	msg.Info("RouteTable")
	msg.Detail("%-20s\t%s", "name", r.name())
	msg.Detail("%-20s\t%s", "access", r.access())
	if r.availabilityZone() != "" {
		msg.Detail("%-20s\t%s", "availabilityZone", r.availabilityZone())
	}
	msg.Detail("%-20s\t%s", "id", r.id())
	msg.Detail("%-20s\t%s", "vpc id", r.vpcId())

	msg.IndentInc()
	for _, route := range r.routeTable.Routes {
		msg.Info("Route")
		if route.DestinationCidrBlock != nil {
			msg.Detail("%-20s\t%s", "destination", *route.DestinationCidrBlock)
		}
		if route.GatewayId != nil {
			msg.Detail("%-20s\t%s", "target", *route.GatewayId)
		}
		if route.NatGatewayId != nil {
			msg.Detail("%-20s\t%s", "target", *route.NatGatewayId)
		}
	}
	msg.IndentDec()

	msg.IndentInc()
	msg.Info("Associations")
	for _, association := range r.routeTable.Associations {
		if association.SubnetId != nil {
			msg.Detail("%-20s\t%s", "subnet", *association.SubnetId)
		}
	}
	msg.IndentDec()
	printTags(r.routeTable.Tags)
}

func (r *routeTable) routeCreated(cidrBlock string) bool {
	for _, route := range r.routeTable.Routes {
		if route.DestinationCidrBlock != nil {
			if *route.DestinationCidrBlock == cidrBlock {
				return true
			}
		}
	}
	return false
}

type gateway int

const (
	igw gateway = iota
	ngw
)

func (r *routeTable) createRoute(req *route.Request, cidrBlock string, gw gateway, s resource.Resource) route.Response {
	if r.Destroyed() || s.Destroyed() {
		return route.OK
	}

	gwPrefix := ""
	gwName := ""
	gwId := ""

	if gw == igw {
		gwPrefix = "Internet"
		gwName = s.(*internetGateway).name()
		gwId = s.(*internetGateway).id()
	} else if gw == ngw {
		gwPrefix = "Nat"
		gwName = s.(*natGateway).name()
		gwId = s.(*natGateway).id()
	}

	msg.Info("Creating Route in RouteTable %s for %s to %sGateway %s", r.name(), cidrBlock, gwPrefix, gwName)
	if r.routeCreated(cidrBlock) {
		msg.Detail("Route exists, skipping...")
		return route.OK
	}

	params := &ec2.CreateRouteInput{
		DestinationCidrBlock: aws.String(cidrBlock),
		RouteTableId:         aws.String(r.id()),
	}

	if gw == igw {
		params.GatewayId = aws.String(gwId)
	} else if gw == ngw {
		params.NatGatewayId = aws.String(gwId)
	}

	if _, err := r.ec2.CreateRoute(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if err := r.Load(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	aaa.Accounting("Creating Route in RouteTable %s for %s to %sGateway %s", r.id(), cidrBlock, gwName, gwId)
	return route.OK
}

func (r *routeTable) deleteRoute(req *route.Request, cidrBlock string) route.Response {
	if r.Destroyed() {
		return route.OK
	}
	msg.Info("Deleting Route in RouteTable %s for Cidr %s", r.name(), cidrBlock)
	if !r.routeCreated(cidrBlock) {
		msg.Detail("Route does not exist, skipping...")
		return route.OK
	}

	params := &ec2.DeleteRouteInput{
		DestinationCidrBlock: aws.String(cidrBlock),
		RouteTableId:         aws.String(r.id()),
	}
	if _, err := r.ec2.DeleteRoute(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if err := r.Load(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	aaa.Accounting("Deleting Route in RouteTable %s for %s", r.id(), cidrBlock)
	return route.OK
}

func (r *routeTable) associated(s *subnet) bool {
	for _, association := range r.routeTable.Associations {
		if association.SubnetId != nil {
			if *association.SubnetId == s.Id() {
				return true
			}
		}
	}
	return false
}

func (r *routeTable) associationId(s *subnet) string {
	for _, association := range r.routeTable.Associations {
		if association.SubnetId != nil && association.RouteTableAssociationId != nil {
			if *association.SubnetId == s.Id() {
				return *association.RouteTableAssociationId
			}
		}
	}
	return ""
}

func (r *routeTable) associate(req *route.Request, s *subnet) route.Response {
	if r.Destroyed() || s.Destroyed() {
		return route.OK
	}
	msg.Info("Associating Subnet %s to RouteTable %s", s.Name(), r.name())
	if r.associated(s) {
		msg.Detail("Association exists, skipping...")
		return route.OK
	}

	params := &ec2.AssociateRouteTableInput{
		RouteTableId: aws.String(r.id()),
		SubnetId:     aws.String(s.Id()),
	}
	if _, err := r.ec2.AssociateRouteTable(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if err := r.Load(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	aaa.Accounting("Associated subnet %s to routetable %s", s.Id(), r.id())
	return route.OK
}

func (r *routeTable) disassociate(req *route.Request, s *subnet) route.Response {
	if r.Destroyed() {
		return route.OK
	}
	msg.Info("Disassociating Subnet %s from RouteTable %s", s.Name(), r.name())
	if !r.associated(s) {
		msg.Detail("Route does not exist, skipping...")
		return route.OK
	}
	associationId := r.associationId(s)
	if associationId == "" {
		msg.Error("Cannot find association between Subnet %s and RouteTable %s", s.Name(), r.name())
		return route.FAIL
	}

	params := &ec2.DisassociateRouteTableInput{
		AssociationId: aws.String(associationId),
	}
	if _, err := r.ec2.DisassociateRouteTable(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if err := r.Load(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	aaa.Accounting("Disassociated subnet %s from routetable %s", s.Id(), r.id())
	return route.OK
}
