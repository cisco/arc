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

// securityGroup implements the resource.ProviderSecurityGroup interface.
type securityGroup struct {
	*config.SecurityGroup
	ec2     *ec2.EC2
	network resource.Network

	secgroup *ec2.SecurityGroup
	id       string
	secrules *securityRules
}

// newSecurityGroup constructs the aws security group.
func newSecurityGroup(net resource.Network, cfg *config.SecurityGroup, c *ec2.EC2) (*securityGroup, error) {
	log.Debug("Initializing AWS Security Group %q", cfg.Name())

	n, ok := net.ProviderNetwork().(*network)
	if !ok {
		return nil, fmt.Errorf("AWS newSecurityGroup: Unable to obtain network")
	}

	s := &securityGroup{
		SecurityGroup: cfg,
		ec2:           c,
		network:       net,
	}
	s.secrules = newSecurityRules(s)
	s.set(n.secgroupCache.find(s))

	return s, nil
}

func (s *securityGroup) Route(req *route.Request) route.Response {
	log.Route(req, "AWS Security Group %q", s.Name())

	switch req.Command() {
	case route.Create:
		return s.create(req)
	case route.Destroy:
		return s.destroy(req)
	case route.Provision:
		return s.update(req)
	case route.Info:
		s.info()
		return route.OK
	}
	return route.FAIL
}

func (s *securityGroup) Created() bool {
	return s.secgroup != nil
}

func (s *securityGroup) Destroyed() bool {
	return !s.Created()
}

func (s *securityGroup) Id() string {
	return s.id
}

func (s *securityGroup) setId(id string) {
	s.id = id
}

func (s *securityGroup) vpcId() string {
	if s.secgroup == nil || s.secgroup.VpcId == nil {
		return ""
	}
	return *s.secgroup.VpcId
}

func (s *securityGroup) set(secgroup *ec2.SecurityGroup) {
	if secgroup == nil || secgroup.GroupId == nil {
		return
	}
	s.secgroup = secgroup
	s.id = *secgroup.GroupId
	s.secrules.set(secgroup)
}

func (s *securityGroup) clear() {
	s.secgroup = nil
	s.id = ""
	s.secrules.clear()
}

func (s *securityGroup) Load() error {

	// Use a cached value if it exists.
	if s.secgroup != nil {
		log.Debug("Skipping security group load, cached...")
		return nil
	}

	params := &ec2.DescribeSecurityGroupsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("vpc-id"),
				Values: []*string{
					aws.String(s.network.Id()),
				},
			},
			{
				Name: aws.String("group-name"),
				Values: []*string{
					aws.String(s.Name()),
				},
			},
		},
	}
	if s.Id() != "" {
		params.Filters = append(params.Filters, &ec2.Filter{
			Name: aws.String("group-id"),
			Values: []*string{
				aws.String(s.Id()),
			},
		})
	}
	resp, err := s.ec2.DescribeSecurityGroups(params)
	if err != nil {
		return err
	}

	for _, securityGroup := range resp.SecurityGroups {
		if s.Id() != "" && securityGroup.GroupId != nil {
			if s.Id() == *securityGroup.GroupId {
				s.set(securityGroup)
				return nil
			}
		}
		if securityGroup.GroupName != nil {
			if s.Name() == *securityGroup.GroupName {
				s.set(securityGroup)
				return nil
			}
		}
	}
	return nil
}

func (s *securityGroup) reload(test func() bool) bool {
	// Clear the cached record.
	s.secgroup = nil

	if test == nil {
		test = s.Created
	}
	return msg.Wait(
		fmt.Sprintf("Waiting for SecurityGroup %s, %s to become available", s.Name(), s.Id()), // title
		fmt.Sprintf("SecurityGroup %s, %s never became available", s.Name(), s.Id()),          // err
		60,   // duration
		test, // test()
		func() bool { // load()
			if err := s.Load(); err != nil {
				msg.Error(err.Error())
				return false
			}
			return true
		},
	)
}

func (s *securityGroup) create(req *route.Request) route.Response {
	msg.Info("SecurityGroup Creation: %s", s.Name())
	if s.Name() == "default" {
		msg.Detail("Can't create default security group, skipping...")
		return route.OK
	}
	if s.Created() {
		msg.Detail("SecurityGroup exists, skipping...")
		return route.OK
	}

	params := &ec2.CreateSecurityGroupInput{
		Description: aws.String(s.Name()),
		GroupName:   aws.String(s.Name()),
		VpcId:       aws.String(s.network.Id()),
	}
	resp, err := s.ec2.CreateSecurityGroup(params)
	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	// Since CreateSecurityGroup doesn't return the same structure as
	// DescribeSecurityGroups, we will set the id of the secgroup and
	// reload rather than calling set().
	if resp == nil || resp.GroupId == nil {
		msg.Error("Failed to create security group %s\n", s.Name())
		return route.FAIL
	}
	s.id = *resp.GroupId
	if !s.reload(nil) {
		return route.FAIL
	}

	if err := createTags(s.ec2, s.Name(), s.Id(), req); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	msg.Detail("Created: %s", s.Id())
	aaa.Accounting("SecurityGroup created: %s", s.Id())

	// Delete default egress rule.
	defaultParams := &ec2.RevokeSecurityGroupEgressInput{
		GroupId: aws.String(s.Id()),
		IpPermissions: []*ec2.IpPermission{
			{
				IpProtocol: aws.String("-1"),
				IpRanges: []*ec2.IpRange{
					{
						CidrIp: aws.String("0.0.0.0/0"),
					},
				},
				FromPort: aws.Int64(-1),
				ToPort:   aws.Int64(-1),
			},
		},
	}
	if _, err := s.ec2.RevokeSecurityGroupEgress(defaultParams); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	// Reload to update the removed egress rule.
	if !s.reload(func() bool {
		return s.Created() && s.secrules.egressRules == nil
	}) {
		return route.FAIL
	}

	// Skip creating the security rules if the "create norules" flag is set.
	if req.Flag("norules") {
		return route.OK
	}

	if err := s.secrules.populate(s); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	if err := s.secrules.create(req, s); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	return route.OK
}

func (s *securityGroup) destroy(req *route.Request) route.Response {
	msg.Info("SecurityGroup Destruction: %s", s.Name())
	if s.Name() == "default" {
		msg.Detail("Can't destroy default security group, skipping...")
		return route.OK
	}
	if s.Destroyed() {
		msg.Detail("SecurityGroup does not exist, skipping...")
		return route.OK
	}

	if err := s.secrules.destroy(req, s); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	// Skip destroying the security groups if the "rules_only" flag is set.
	if req.Flag("rules_only") {
		s.secrules.clear()
		return route.OK
	}

	params := &ec2.DeleteSecurityGroupInput{
		GroupId: aws.String(s.Id()),
	}
	if _, err := s.ec2.DeleteSecurityGroup(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	msg.Detail("Destroyed: %s", s.Id())
	aaa.Accounting("SecurityGroup destroyed: %s", s.Id())
	s.clear()
	n := s.network.ProviderNetwork().(*network)
	n.secgroupCache.remove(s.Name())
	return route.OK
}

func (s *securityGroup) update(req *route.Request) route.Response {
	msg.Info("SecurityGroup Update: %s", s.Name())
	if s.Destroyed() {
		msg.Detail("SecurityGroup does not exist, skipping...")
		return route.OK
	}
	// Only update the rules if they are different.
	cfgSecrules := newSecurityRules(s)
	if err := cfgSecrules.populate(s); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	quiet := true
	if s.secrules.compareEqual(cfgSecrules, quiet, s.Name()) == true {
		msg.Detail("SecurityGroup %s has not changed, skipping...", s.Name())
		return route.OK
	}

	// Destroy existing rules
	if err := s.secrules.destroy(req, s); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	// Create new rules from configuration
	s.secrules = newSecurityRules(s)
	if err := s.secrules.populate(s); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if err := s.secrules.create(req, s); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if err := s.Load(); err != nil {
		return route.FAIL
	}
	// Update tags
	if err := createTags(s.ec2, s.Name(), s.Id(), req); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	return route.OK
}

func (s *securityGroup) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if s.Destroyed() {
		a.Audit(aaa.Configured, "%s", s.Name())
		return nil
	}
	// Compare security rules loaded from AWS against secrules created from
	// configuration.
	cfgSecrules := newSecurityRules(s)
	if err := cfgSecrules.populate(s); err != nil {
		return err
	}
	verbose := true
	s.secrules.compareEqual(cfgSecrules, verbose, s.Name(), flags...)
	return nil
}

func (s *securityGroup) info() {
	if s.Destroyed() {
		return
	}
	msg.Info("SecurityGroup")
	msg.Detail("%-20s\t%s", "name", s.Name())
	msg.Detail("%-20s\t%s", "id", s.Id())
	msg.Detail("%-20s\t%s", "vpc", s.vpcId())
	msg.IndentInc()
	s.secrules.info()
	msg.IndentDec()
	printTags(s.secgroup.Tags)
}
