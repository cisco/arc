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

	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
)

// compute implements the resource.ProviderCompute  interface.
type compute struct {
	*config.Compute
	ec2 *ec2.EC2

	volumeCache   *volumeCache
	instanceCache *instanceCache
}

// newCompute constructs the aws compute.
func newCompute(cfg *config.Compute, e *ec2.EC2) (*compute, error) {
	log.Debug("Initializing AWS Compute")
	c := &compute{
		Compute: cfg,
		ec2:     e,
	}

	var err error
	c.volumeCache, err = newVolumeCache(c)
	if err != nil {
		return nil, err
	}
	c.instanceCache, err = newInstanceCache(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *compute) AuditVolumes(flags ...string) error {
	return c.volumeCache.audit(flags...)
}

func (c *compute) AuditEIP(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	err := aaa.NewAuditWithOptions(flags[0], true, false, false)
	if err != nil {
		return err
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit Object does not exist")
	}
	params := &ec2.DescribeAddressesInput{}
	addr, err := c.ec2.DescribeAddresses(params)
	if err != nil {
		return err
	}
	for _, v := range addr.Addresses {
		if v.AssociationId == nil {
			a.Audit(aaa.Deployed, "Elastic IP %q is not associated with anything", *v.PublicIp)
		}
	}
	return nil
}

func (c *compute) AuditInstances(flags ...string) error {
	return c.instanceCache.audit(flags...)
}
