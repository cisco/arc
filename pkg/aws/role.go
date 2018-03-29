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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	// "github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

type role struct {
	*config.Role
	provider *identityManagementProvider
	iam      *iam.IAM

	identityManagement *identityManagement

	role *iam.Role
}

func newRole(rl resource.Role, cfg *config.Role, prov *identityManagementProvider) (resource.ProviderRole, error) {
	log.Debug("Initializing AWS Role %q", cfg.Name())

	r := &role{
		Role:               cfg,
		provider:           prov,
		iam:                prov.iam,
		identityManagement: rl.IdentityManagement().ProviderIdentityManagement().(*identityManagement),
	}

	return r, nil
}

func (r *role) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit Object does not exist")
	}
	if r.role == nil {
		a.Audit(aaa.Configured, "%s", r.Name())
		return nil
	}
	log.Debug("Deployed description %q", aws.StringValue(r.role.Description))
	log.Debug("Configed description %q", r.Description())
	if r.role.Description != nil && r.Description() != "" && strings.Compare(aws.StringValue(r.role.Description), r.Description()) != 0 {
		a.Audit(aaa.Mismatched, "Role %q's description does not match configured description", aws.StringValue(r.role.RoleName))
		return nil
	}
	if r.role.Description != nil && r.Description() == "" {
		a.Audit(aaa.Mismatched, "Role %q has a deployed description but not a configured one", aws.StringValue(r.role.RoleName))
		return nil
	}
	if r.role.Description == nil && r.Description() != "" {
		a.Audit(aaa.Mismatched, "Role %q has a configured description but not a deployed one", aws.StringValue(r.role.RoleName))
		return nil
	}
	return nil
}

func (r *role) set(role *iam.Role) {
	r.role = role
}

func (r *role) clear() {
	r.role = nil
}

func (r *role) Created() bool {
	return r.role != nil
}

func (r *role) Destroyed() bool {
	return r.role == nil
}

func (r *role) Create(flags ...string) error {
	msg.Info("Role Creation: %s", r.Name())
	params := &iam.CreateRoleInput{}

	resp, err := r.iam.CreateRole(params)
	if err != nil {
		return err
	}
	r.role = resp.Role
	if err := r.Load(); err != nil {
		return err
	}
	msg.Detail("Role created: %s", r.Name())
	return nil
}

func (r *role) Load() error {
	return nil
}

func (r *role) Destroy(flags ...string) error {
	msg.Info("Role Deletion: %s", r.Name())
	params := &iam.DeleteRoleInput{}
	_, err := r.iam.DeleteRole(params)
	if err != nil {
		return err
	}
	msg.Detail("Role deleted: %s", r.Name())
	r.clear()
	return nil
}

func (r *role) Info() {
	if r.Destroyed() {
		return
	}
	msg.Info("Role")
	msg.Detail("%-20s\t%s", "name", aws.StringValue(r.role.RoleName))
	msg.Detail("%-20s\t%+v", "role", r.role)
}
