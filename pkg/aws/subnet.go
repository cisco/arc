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
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

// subnet implements the resource.ProviderSubnet interface.
type subnet struct {
	*config.Subnet
	ec2     *ec2.EC2
	network *network

	id_    string
	subnet *ec2.Subnet
}

// newSubnet constructs the aws subnet.
func newSubnet(net resource.Network, cfg *config.Subnet, c *ec2.EC2) (*subnet, error) {
	log.Debug("Initializing AWS Subnet %q", cfg.Name())

	n, ok := net.ProviderNetwork().(*network)
	if !ok {
		return nil, fmt.Errorf("AWS newSubnet: Unable to obtain network")
	}

	s := &subnet{
		Subnet:  cfg,
		ec2:     c,
		network: n,
	}
	s.set(n.subnetCache.find(s))

	return s, nil
}

func (s *subnet) Route(req *route.Request) route.Response {
	log.Route(req, "AWS Subnet %q", s.Name())

	switch req.Command() {
	case route.Create:
		return s.create(req)
	case route.Destroy:
		return s.destroy(req)
	case route.Info:
		s.info()
		return route.OK
	}
	return route.FAIL
}

func (s *subnet) Created() bool {
	return s.subnet != nil
}

func (s *subnet) Destroyed() bool {
	return !s.Created()
}

func (s *subnet) Id() string {
	return s.id_
}

func (s *subnet) vpcId() string {
	if s.subnet == nil || s.subnet.VpcId == nil {
		return ""
	}
	return *s.subnet.VpcId
}

func (s *subnet) State() string {
	if s.subnet == nil || s.subnet.State == nil {
		return ""
	}
	return *s.subnet.State
}

func (s *subnet) availableIpAddressCount() int64 {
	if s.subnet == nil || s.subnet.AvailableIpAddressCount == nil {
		return -1
	}
	return *s.subnet.AvailableIpAddressCount
}

func (s *subnet) defaultForAz() bool {
	if s.subnet == nil || s.subnet.DefaultForAz == nil {
		return false
	}
	return *s.subnet.DefaultForAz
}

func (s *subnet) mapPublicIpOnLaunch() bool {
	if s.subnet == nil || s.subnet.MapPublicIpOnLaunch == nil {
		return false
	}
	return *s.subnet.MapPublicIpOnLaunch
}

func (s *subnet) set(subnet *ec2.Subnet) {
	if subnet == nil || subnet.SubnetId == nil {
		return
	}
	s.subnet = subnet
	s.id_ = *subnet.SubnetId
}

func (s *subnet) clear() {
	s.subnet = nil
	s.id_ = ""
}

func (s *subnet) Load() error {

	// Use a cached value if it exists.
	if s.subnet != nil {
		log.Debug("Skipping subnet load, cached...")
		return nil
	}

	params := &ec2.DescribeSubnetsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(s.network.vpc.id()),
				},
			},
		},
	}
	if s.Id() != "" {
		params.Filters = append(params.Filters, &ec2.Filter{
			Name: aws.String("subnet-id"),
			Values: []*string{
				aws.String(s.Id()),
			},
		})
	} else {
		params.Filters = append(params.Filters, &ec2.Filter{
			Name: aws.String("tag:Name"),
			Values: []*string{
				aws.String(s.Name()),
			},
		})
	}
	if len(params.Filters) == 0 {
		return fmt.Errorf("Internal error, cannot load subnet, name or id not present")
	}
	resp, err := s.ec2.DescribeSubnets(params)
	if err != nil {
		msg.Error(err.Error())
		return err
	}
	for _, subnet := range resp.Subnets {
		if s.Id() != "" && subnet.SubnetId != nil {
			if s.Id() == *subnet.SubnetId {
				s.set(subnet)
				return nil
			}
		}
		if subnet.Tags == nil {
			break
		}
		for _, tag := range subnet.Tags {
			if tag != nil && tag.Key != nil && *tag.Key == "Name" && tag.Value != nil && *tag.Value == s.Name() {
				s.set(subnet)
				return nil
			}
		}
	}
	return nil
}

func (s *subnet) reload() bool {
	// Clear the cached value
	s.subnet = nil

	return msg.Wait(
		fmt.Sprintf("Waiting for Subnet %s, %s to become available", s.Name(), s.Id()), //title
		fmt.Sprintf("Subnet %s, %s never became available", s.Name(), s.Id()),          // err
		600, // duration
		func() bool { return s.State() == "available" }, // test()
		func() bool { //load()
			if err := s.Load(); err != nil {
				msg.Error(err.Error())
				return false
			}
			return true
		},
	)
}

func (s *subnet) findRouteTable() (*routeTable, string) {
	name := s.Access()
	if s.Access() == "public_elastic" {
		name = "public"
	}
	if s.ManageRoutes() {
		name = s.GroupName()
	}
	if s.Access() == "private" {
		name = name + "-" + s.AvailabilityZone()
	}
	return s.network.routeTables.find(name), name
}

func (s *subnet) create(req *route.Request) route.Response {
	msg.Info("Subnet Creation: %s %s", s.Name(), s.CidrBlock())
	if s.Created() {
		msg.Detail("Subnet exists, skipping...")
		return route.OK
	}

	params := &ec2.CreateSubnetInput{
		CidrBlock:        aws.String(s.CidrBlock()),
		VpcId:            aws.String(s.network.vpc.id()),
		AvailabilityZone: aws.String(s.AvailabilityZone()),
	}

	resp, err := s.ec2.CreateSubnet(params)
	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	s.set(resp.Subnet)
	if !s.reload() {
		return route.FAIL
	}

	if err := createTags(s.ec2, s.Name(), s.Id(), req); err != nil {
		msg.Error(err.Error())
		s.destroy(req)
		return route.FAIL
	}
	msg.Detail("Created %s", s.Id())
	aaa.Accounting("Subnet created: %s", s.Id())

	if s.Access() != "local" {
		routeTable, name := s.findRouteTable()
		if routeTable == nil {
			msg.Error("Cannot find %s routetable", name)
			s.destroy(req)
			return route.FAIL
		}
		if resp := routeTable.associate(req, s); resp != route.OK {
			msg.Error("Failed to associate %s with %s", s.Name(), routeTable.name())
			s.destroy(req)
			return route.FAIL
		}
	}
	if s.Access() == "public" || s.Access() == "public_elastic" {
		params := &ec2.ModifySubnetAttributeInput{
			SubnetId: aws.String(s.Id()),
			MapPublicIpOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
		}
		if _, err := s.ec2.ModifySubnetAttribute(params); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	}
	return route.OK
}

func (s *subnet) destroy(req *route.Request) route.Response {
	if s.Access() != "local" {
		routeTable, name := s.findRouteTable()
		if routeTable == nil {
			msg.Error("Cannot find %s routetable", name)
			return route.FAIL
		}
		if resp := routeTable.disassociate(req, s); resp != route.OK {
			msg.Error("Failed to disassociate %s with %s", s.Name(), routeTable.name())
			return route.FAIL
		}
	}

	msg.Info("Subnet Destruction: %s", s.Name())
	if s.Destroyed() {
		msg.Detail("Subnet does not exist, skipping...")
		return route.OK
	}

	params := &ec2.DeleteSubnetInput{
		SubnetId: aws.String(s.Id()),
	}
	if _, err := s.ec2.DeleteSubnet(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	msg.Detail("Destroyed: %s", s.Id())
	aaa.Accounting("Subnet Destroyed: %s", s.Id())

	s.clear()
	s.network.subnetCache.remove(s)

	return route.OK
}

func (s *subnet) info() {
	if s.Destroyed() {
		return
	}
	msg.Info("Subnet")
	msg.Detail("%-20s\t%s", "name", s.Name())
	msg.Detail("%-20s\t%s", "cidr", s.CidrBlock())
	msg.Detail("%-20s\t%s", "access", s.Access())
	msg.Detail("%-20s\t%s", "availability zone", s.AvailabilityZone())
	msg.Detail("%-20s\t%s", "id", s.Id())
	msg.Detail("%-20s\t%s", "vpc id", s.vpcId())
	msg.Detail("%-20s\t%s", "state", s.State())
	msg.Detail("%-20s\t%d", "available ips", s.availableIpAddressCount())
	msg.Detail("%-20s\t%t", "default for az", s.defaultForAz())
	msg.Detail("%-20s\t%t", "map public ip", s.mapPublicIpOnLaunch())
	printTags(s.subnet.Tags)
}

func (s *subnet) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if !s.Created() {
		a.Audit(aaa.Configured, "%s", s.Name())
		return nil
	}
	if s.CidrBlock() != *s.subnet.CidrBlock {
		a.Audit(aaa.Mismatched, "%s: cidr block mismatch - configured: %s, deployed: %s", s.Name(), s.CidrBlock(), *s.subnet.CidrBlock)
	}
	if s.AvailabilityZone() != *s.subnet.AvailabilityZone {
		a.Audit(aaa.Mismatched, "%s: az mismatch - configured: %s, deployed: %s", s.Name(), s.AvailabilityZone(), *s.subnet.AvailabilityZone)
	}
	return nil
}
