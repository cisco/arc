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

type natGateway struct {
	network           *network
	ec2               *ec2.EC2
	name_             string
	availabilityZone_ string
	subnet            *subnet

	id_        string
	ngw        *ec2.NatGateway
	ngwAddress *ec2.NatGatewayAddress
}

func newNatGateway(c *ec2.EC2, net resource.Network, name, availabilityZone string) (*natGateway, error) {
	log.Debug("Initializing AWS NatGateway %q", name)

	n, ok := net.ProviderNetwork().(*network)
	if !ok {
		return nil, fmt.Errorf("AWS newNatGateway: Unable to obtain provider network")
	}

	subnetGroup := net.SubnetGroups().Find("public")
	if subnetGroup == nil {
		return nil, fmt.Errorf("Cannot find public subnet group")
	}

	sub := subnetGroup.Find("public-" + availabilityZone)
	if sub == nil {
		return nil, fmt.Errorf("Cannot find public subnet for az %s", availabilityZone)
	}

	s, ok := sub.ProviderSubnet().(*subnet)
	if !ok {
		return nil, fmt.Errorf("AWS newNatGateway: Unable to obtain provider subnet")
	}

	ngw := &natGateway{
		network:           n,
		ec2:               c,
		name_:             name,
		availabilityZone_: availabilityZone,
		subnet:            s,
	}
	return ngw, nil
}

func (n *natGateway) Route(req *route.Request) route.Response {
	log.Route(req, "AWS NatGateway %q", n.name())

	if req.Top() != "" {
		n.help()
		return route.FAIL
	}

	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load:
		if err := n.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Create:
		return n.create(req)
	case route.Destroy:
		return n.destroy(req)
	case route.Help:
		n.help()
		return route.OK
	case route.Info:
		n.info()
		return route.OK
	}
	msg.Error("Unknown natgateway command %q.", req.Command())
	return route.FAIL
}

func (n *natGateway) Created() bool {
	return n != nil && n.ngw != nil
}

func (n *natGateway) Destroyed() bool {
	return !n.Created()
}

func (n *natGateway) name() string {
	return n.name_
}

func (n *natGateway) availabilityZone() string {
	return n.availabilityZone_
}

func (n *natGateway) id() string {
	return n.id_
}

func (n *natGateway) state() string {
	if n.ngw == nil || n.ngw.State == nil {
		return ""
	}
	return *n.ngw.State
}

func (n *natGateway) vpcId() string {
	if n.ngw == nil || n.ngw.VpcId == nil {
		return ""
	}
	return *n.ngw.VpcId
}

func (n *natGateway) subnetId() string {
	if n.ngw == nil || n.ngw.SubnetId == nil {
		return ""
	}
	return *n.ngw.SubnetId
}

func (n *natGateway) allocationId() string {
	if n.ngwAddress == nil || n.ngwAddress.AllocationId == nil {
		return ""
	}
	return *n.ngwAddress.AllocationId
}

func (n *natGateway) privateIp() string {
	if n.ngwAddress == nil || n.ngwAddress.PrivateIp == nil {
		return ""
	}
	return *n.ngwAddress.PrivateIp
}

func (n *natGateway) publicIp() string {
	if n.ngwAddress == nil || n.ngwAddress.PublicIp == nil {
		return ""
	}
	return *n.ngwAddress.PublicIp
}

func (n *natGateway) set(ngw *ec2.NatGateway, ngwAddress *ec2.NatGatewayAddress) {
	if ngw == nil || ngwAddress == nil || ngw.NatGatewayId == nil {
		return
	}
	n.ngw = ngw
	n.ngwAddress = ngwAddress
	n.id_ = *ngw.NatGatewayId
}

func (n *natGateway) clear() {
	n.ngw = nil
	n.ngwAddress = nil
	n.id_ = ""
}

func (n *natGateway) Load() error {
	var max int64 = 1000
	var token *string

	n.ngw = nil
	n.ngwAddress = nil

	done := false
	for !done {
		params := &ec2.DescribeNatGatewaysInput{
			Filter: []*ec2.Filter{
				{
					Name: aws.String("subnet-id"),
					Values: []*string{
						aws.String(n.subnet.Id()),
					},
				},
				{
					Name: aws.String("vpc-id"),
					Values: []*string{
						aws.String(n.network.vpc.id()),
					},
				},
			},
			MaxResults: aws.Int64(max),
			NextToken:  token,
		}
		if n.id() != "" {
			// If we know the id (we are doing a reload) we will add it to the params.
			// We want to know about all states.
			params.Filter = append(params.Filter, &ec2.Filter{
				Name: aws.String("nat-gateway-id"),
				Values: []*string{
					aws.String(n.id()),
				},
			})
			params.NatGatewayIds = []*string{aws.String(n.id())}
		} else {
			// Otherwise we will only search for pending or available ngws.
			params.Filter = append(params.Filter, &ec2.Filter{
				Name: aws.String("state"),
				Values: []*string{
					aws.String("available"),
					aws.String("pending"),
				},
			})
		}
		resp, err := n.ec2.DescribeNatGateways(params)
		if err != nil {
			return err
		}

		for _, natGateway := range resp.NatGateways {
			if natGateway.SubnetId == nil {
				continue
			}
			if *natGateway.SubnetId == n.subnet.Id() {
				for _, natGatewayAddress := range natGateway.NatGatewayAddresses {
					if natGatewayAddress.AllocationId != nil {
						n.set(natGateway, natGatewayAddress)
						return nil
					}
				}
			}
		}
		if resp.NextToken != nil {
			token = resp.NextToken
		} else {
			done = true
		}
	}
	return nil
}

func (n *natGateway) reload() bool {
	return msg.Wait(
		fmt.Sprintf("Waiting for NatGateway %s, %s to become available", n.name(), n.id()), //title
		fmt.Sprintf("NatGateway %s, %s never became available", n.name(), n.id()),          // err
		600, // duration
		func() bool { return n.state() == "available" }, // test()
		func() bool { // load()
			switch n.state() {
			case "failed", "deleting", "deleted":
				msg.Error("NatGateway %s %s failed to become available, state: %s ", n.name(), n.id(), n.state())
				return false
			}
			if err := n.Load(); err != nil {
				msg.Error(err.Error())
				return false
			}
			return true
		},
	)
}

func (n *natGateway) create(req *route.Request) route.Response {
	if n.Created() {
		msg.Info("NatGateway Creation: %s", n.name())
		msg.Detail("NatGateway exists, skipping...")
		return route.OK
	}

	log.Debug("Creating natGateway using subnet %s, id: %s", n.subnet.Name(), n.subnet.Id())

	msg.Info("Elastic IP Allocation")
	eipParams := &ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
	}
	eip, err := n.ec2.AllocateAddress(eipParams)
	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	msg.Detail("Created: %s", *eip.AllocationId)
	aaa.Accounting("Elastic IP allocated: %s", *eip.AllocationId)

	msg.Info("NatGateway Creation: %s", n.name())
	params := &ec2.CreateNatGatewayInput{
		AllocationId: eip.AllocationId,
		SubnetId:     aws.String(n.subnet.Id()),
	}
	resp, err := n.ec2.CreateNatGateway(params)
	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	natGatewayAddress := &ec2.NatGatewayAddress{AllocationId: eip.AllocationId}
	n.set(resp.NatGateway, natGatewayAddress)
	if !n.reload() {
		return route.FAIL
	}

	msg.Detail("Created: %s", n.id())
	aaa.Accounting("natGateway created: %s", n.id())

	for _, routeTable := range n.network.routeTables.routeTables {
		if routeTable.access() == "private" && n.availabilityZone() == routeTable.availabilityZone() {
			if resp := routeTable.createRoute(req, "0.0.0.0/0", ngw, n); resp != route.OK {
				return resp
			}
		}
	}

	return route.OK
}

func (n *natGateway) destroy(req *route.Request) route.Response {
	if n.Destroyed() {
		msg.Info("NatGateway Destruction: %s", n.name())
		msg.Detail("NatGateway does not exist, skipping...")
		return route.OK
	}

	allocationId := n.allocationId()

	for _, routeTable := range n.network.routeTables.routeTables {
		if routeTable.access() == "private" && n.availabilityZone() == routeTable.availabilityZone() {
			if resp := routeTable.deleteRoute(req, "0.0.0.0/0"); resp != route.OK {
				return resp
			}
		}
	}

	msg.Info("NatGateway Destruction: %s", n.name())
	params := &ec2.DeleteNatGatewayInput{
		NatGatewayId: aws.String(n.id()),
	}
	if _, err := n.ec2.DeleteNatGateway(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	msg.Detail("Destroyed: %s", n.id())
	aaa.Accounting("NatGateway destroyed: %s", n.id())

	ok := msg.Wait(
		fmt.Sprintf("Waiting for NatGateway %s to delete", n.id()), // title
		fmt.Sprintf("NatGateway %s never deleted", n.id()),         // err
		600, // duration
		func() bool { return n.Destroyed() || n.state() == "deleted" }, // test()
		func() bool { return n.Load() == nil },                         // load()
	)
	if !ok {
		return route.FAIL
	}
	msg.Detail("NatGateway has been deleted")
	n.clear()

	msg.Info("Releasing elastic ip: %s", allocationId)
	eipParams := &ec2.ReleaseAddressInput{
		AllocationId: aws.String(allocationId),
	}
	_, err := n.ec2.ReleaseAddress(eipParams)
	if err != nil {
		msg.Warn(err.Error())
		return route.OK
	}

	msg.Detail("Released: %s", allocationId)
	aaa.Accounting("Elastic IP release: %s", allocationId)
	return route.OK
}

func (n *natGateway) help() {
	commands := []help.Command{
		{route.Create.String(), fmt.Sprintf("create %s nat gateway", n.name())},
		{route.Destroy.String(), fmt.Sprintf("destroy %s nat gateway", n.name())},
		{route.Info.String(), fmt.Sprintf("show information about allocated %s nat gateway", n.name())},
		{route.Help.String(), "show this help"},
	}
	help.Print("network natgateway", commands)
}

func (n *natGateway) info() {
	if n.Destroyed() {
		return
	}
	msg.Info("NatGateway")
	msg.Detail("%-20s\t%s", "name", n.name())
	msg.Detail("%-20s\t%s", "id", n.id())
	msg.Detail("%-20s\t%s", "state", n.state())
	msg.Detail("%-20s\t%s", "vpc id", n.vpcId())
	msg.Detail("%-20s\t%s", "subnet id", n.subnetId())
	msg.Detail("%-20s\t%s", "allocation id", n.allocationId())
	msg.Detail("%-20s\t%s", "private ip", n.privateIp())
	msg.Detail("%-20s\t%s", "public ip", n.publicIp())
}
