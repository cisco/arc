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
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

// elasticIP implements the resource.ProviderElasticIP interface.
type elasticIP struct {
	ec2      *ec2.EC2
	instance resource.Instance

	id            string
	associationId string
	eip           *ec2.Address
}

// newElasticIP is the constructor for a elasticIP object. it returns a non-nil error upon failure.
func newElasticIP(i resource.Instance, p *dataCenterProvider) (resource.ProviderElasticIP, error) {
	log.Info("Initializing AWS elasticIP for instance %q", i.Name())
	return &elasticIP{
		ec2:      p.ec2,
		instance: i,
	}, nil
}

func (e *elasticIP) Route(req *route.Request) route.Response {
	return route.OK
}

func (e *elasticIP) Id() string {
	return e.id
}

func (e *elasticIP) Instance() resource.Instance {
	return e.instance
}

func (e *elasticIP) IpAddress() string {
	if e.eip == nil || e.eip.PublicIp == nil {
		return ""
	}
	return *e.eip.PublicIp
}

func (e *elasticIP) Created() bool {
	return e.eip != nil
}

func (e *elasticIP) Destroyed() bool {
	return e.eip == nil
}

func (e *elasticIP) Attached() bool {
	return e.associationId != ""
}

func (e *elasticIP) Detached() bool {
	return e.associationId == ""
}

func (e *elasticIP) set(eip *ec2.Address) {
	e.eip = eip
	if e.eip != nil && e.eip.AllocationId != nil {
		e.id = *e.eip.AllocationId
	}
	if e.eip.AssociationId != nil {
		e.associationId = *eip.AssociationId
	}
}

func (e *elasticIP) clear() {
	e.eip = nil
	e.id = ""
}

func (e *elasticIP) Load() error {
	filters := []*ec2.Filter{}

	log.Debug("Load AWS elastic IP. Instance Id: %q", e.Instance().Id())
	if e.Instance().Id() != "" {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("instance-id"),
			Values: []*string{aws.String(e.Instance().Id())},
		})
	}

	log.Debug("Load AWS elastic IP. Allocation Id: %q", e.Id())
	if e.Id() != "" {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("allocation-id"),
			Values: []*string{aws.String(e.Id())},
		})
	}

	if len(filters) == 0 {
		log.Debug("Cannot load AWS elastic IP. No instance id or allocation id available")
		return nil
	}
	params := &ec2.DescribeAddressesInput{
		Filters: filters,
	}
	addr, err := e.ec2.DescribeAddresses(params)
	if err != nil {
		return err
	}
	if addr.Addresses != nil && len(addr.Addresses) == 1 {
		e.set(addr.Addresses[0])
	}
	return nil
}

// Create allocates the elastic IP.
func (e *elasticIP) Create() error {
	if e.Created() {
		return nil
	}
	log.Debug("Allocating elastic IP %q", e.Instance().Name())
	params := &ec2.AllocateAddressInput{
		Domain: aws.String("vpc"),
	}
	addr, err := e.ec2.AllocateAddress(params)
	if err != nil {
		return err
	}
	e.set(&ec2.Address{
		AllocationId: addr.AllocationId,
		Domain:       addr.Domain,
		PublicIp:     addr.PublicIp,
	})
	aaa.Accounting("Elastic IP created: %s", e.Id())
	return nil
}

// Attach associates the allocated elastic IP to the instance.
func (e *elasticIP) Attach() error {
	if e.Destroyed() || e.Attached() {
		return fmt.Errorf("IpAddress %q already attached to %q", e.Id(), e.Instance().Name())
	}
	log.Debug("Associating elastic IP %q with instance %q", e.IpAddress(), e.Instance().Name())

	params := &ec2.AssociateAddressInput{
		AllocationId:       aws.String(e.Id()),
		AllowReassociation: aws.Bool(false),
		InstanceId:         aws.String(e.Instance().Id()),
	}
	if _, err := e.ec2.AssociateAddress(params); err != nil {
		return err
	}
	aaa.Accounting("Elastic IP associated: %s", e.Id())
	return nil
}

// Detach disassocates the allocated elastic IP from the instance.
func (e *elasticIP) Detach() error {
	if e.Destroyed() {
		return nil
	}
	log.Debug("Disassociating elastic IP %q from %q", e.Id(), e.Instance().Name())

	params := &ec2.DisassociateAddressInput{
		AssociationId: e.eip.AssociationId,
	}
	if _, err := e.ec2.DisassociateAddress(params); err != nil {
		return err
	}
	aaa.Accounting("Elastic IP disassociated: %s", e.Id())
	e.associationId = ""
	return nil
}

// Destroy releases the elastic IP.
func (e *elasticIP) Destroy() error {
	if e.Destroyed() {
		return nil
	}
	log.Debug("Releasing elasticIP %q from %q", e.Id(), e.Instance().Name())

	params := &ec2.ReleaseAddressInput{
		AllocationId: aws.String(e.Id()),
	}
	if _, err := e.ec2.ReleaseAddress(params); err != nil {
		return err
	}
	aaa.Accounting("Elastic IP destroyed: %s", e.Id())
	e.clear()
	return nil
}
