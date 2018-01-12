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
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

// role implements the resource.ProviderRole interface.
type role struct {
	name     string
	instance resource.Instance
	ec2      *ec2.EC2

	id    string
	state string
	arn   string
	role  *ec2.IamInstanceProfileSpecification
}

// newRole constructs the role.
func newRole(name string, p *dataCenterProvider, in resource.Instance) (resource.ProviderRole, error) {
	log.Info("Initializing role")
	var s *ec2.IamInstanceProfileSpecification = nil
	if name != "" {
		s = &ec2.IamInstanceProfileSpecification{
			Arn:  aws.String(newIamInstanceProfile(p.number, name).String()),
			Name: aws.String(name),
		}
	}

	r := &role{
		name:     name,
		instance: in,
		ec2:      p.ec2,
		role:     s,
	}
	return r, nil
}

func (r *role) Route(req *route.Request) route.Response {
	return route.OK
}

func (r *role) Load() error {
	if r.InstanceId() == "" {
		return nil
	}

	params := &ec2.DescribeIamInstanceProfileAssociationsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("instance-id"),
				Values: []*string{aws.String(r.InstanceId())},
			},
		},
	}
	role, err := r.ec2.DescribeIamInstanceProfileAssociations(params)
	if err != nil {
		msg.Error(err.Error())
		return err
	}
	if role.IamInstanceProfileAssociations != nil {
		r.set(role.IamInstanceProfileAssociations[0])
		return nil
	}
	r.state = "disassociated"
	return nil
}

func (r *role) Created() bool {
	return r.role != nil
}

func (r *role) Destroyed() bool {
	return r.role == nil
}

func (r *role) Id() string {
	return r.id
}

func (r *role) State() string {
	return r.state
}

func (r *role) InstanceId() string {
	return r.instance.Id()
}

func (r *role) Attached() bool {
	return r.State() == "associated"
}

func (r *role) Detached() bool {
	// The state could be associating, associated, disassociating, or disassociated
	return r.State() != "associated"
}

func (r *role) Attach() error {
	params := &ec2.AssociateIamInstanceProfileInput{
		IamInstanceProfile: r.role,
		InstanceId:         aws.String(r.InstanceId()),
	}

	role, err := r.ec2.AssociateIamInstanceProfile(params)
	if err != nil {
		return err
	}
	r.set(role.IamInstanceProfileAssociation)
	msg.Detail("Role Attached: %s", r.name)
	aaa.Accounting("Role created: %s, %s", r.name, r.Id())
	return nil
}

func (r *role) Detach() error {
	params := &ec2.DisassociateIamInstanceProfileInput{
		AssociationId: aws.String(r.Id()),
	}

	role, err := r.ec2.DisassociateIamInstanceProfile(params)
	if err != nil {
		return err
	}
	r.state = *role.IamInstanceProfileAssociation.State
	msg.Detail("Role Detached: %s", filepath.Base(r.arn))
	aaa.Accounting("Role destroyed: %s, %s", filepath.Base(r.arn), r.Id())
	return nil
}

func (r *role) Update() error {
	if err := r.Load(); err != nil {
		return err
	}

	if filepath.Base(r.arn) == "." && r.name == "" {
		return nil
	}
	if filepath.Base(r.arn) == r.name {
		return nil
	}

	msg.Info("Role Update: %s", r.name)
	if r.Detached() {
		msg.Detail("Role does not exist... creating")
		return r.Attach()
	}
	if r.Attached() && r.name == "" {
		msg.Detail("Role no longer exists... removing")
		return r.Detach()
	}

	params := &ec2.ReplaceIamInstanceProfileAssociationInput{
		AssociationId: aws.String(r.Id()),
		IamInstanceProfile: &ec2.IamInstanceProfileSpecification{
			Arn: r.role.Arn,
		},
	}
	msg.Detail("Changing role from %s to %s", filepath.Base(r.arn), r.name)
	role, err := r.ec2.ReplaceIamInstanceProfileAssociation(params)
	if err != nil {
		return err
	}
	r.set(role.IamInstanceProfileAssociation)
	aaa.Accounting("Role replaced: %s, %s", r.name, r.Id())
	return nil
}

func (r *role) set(role *ec2.IamInstanceProfileAssociation) {
	r.state = *role.State
	r.id = *role.AssociationId
	r.arn = *role.IamInstanceProfile.Arn
}
