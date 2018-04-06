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
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

type role struct {
	*config.Role
	provider *identityManagementProvider
	iam      *iam.IAM

	identityManagement *identityManagement

	role             *iam.Role
	instanceProfile  *iam.InstanceProfile
	deployedPolicies []string
	policies         []string
}

func newRole(rl resource.Role, cfg *config.Role, prov *identityManagementProvider) (resource.ProviderRole, error) {
	log.Debug("Initializing AWS Role %q", cfg.Name())

	r := &role{
		Role:               cfg,
		provider:           prov,
		iam:                prov.iam,
		identityManagement: rl.IdentityManagement().ProviderIdentityManagement().(*identityManagement),
	}
	for _, p := range r.Policies() {
		policy := newIamPolicy(prov.number, p)
		r.policies = append(r.policies, policy.String())
	}

	return r, nil
}

func (r *role) Load() error {
	if err := r.loadPolicies(); err != nil {
		return err
	}
	if role := r.identityManagement.roleCache.find(r); role != nil {
		log.Debug("Skipping role load, cached...")
		r.set(role)
		return nil
	}
	params := &iam.GetRoleInput{
		RoleName: aws.String(r.Name()),
	}
	resp, err := r.iam.GetRole(params)
	if err != nil {
		if strings.Contains(err.Error(), "NoSuchEntity") {
			log.Debug("No Such Entity: Role %q", r.Name())
			return nil
		}
		return err
	}
	r.role = resp.Role
	return nil
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
	// Mismatched Policies
	for _, p := range r.roguePolicies() {
		a.Audit(aaa.Mismatched, "Role %q's policy %q is deployed but not configured", r.Name(), p)
	}

	for _, p := range r.orphanedPolicies() {
		a.Audit(aaa.Mismatched, "Role %q's policy %q is configured but not deployed", r.Name(), p)
	}
	// Mismatched Description
	if r.role.Description != nil && r.Description() != "" && strings.Compare(aws.StringValue(r.role.Description), r.Description()) != 0 {
		a.Audit(aaa.Mismatched, "Role %q's description does not match configured description", aws.StringValue(r.role.RoleName))
	}
	if r.role.Description != nil && r.Description() == "" {
		a.Audit(aaa.Mismatched, "Role %q has a deployed description but not a configured one", aws.StringValue(r.role.RoleName))
	}
	if r.role.Description == nil && r.Description() != "" {
		a.Audit(aaa.Mismatched, "Role %q has a configured description but not a deployed one", aws.StringValue(r.role.RoleName))
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
	if err := r.createRole(flags...); err != nil {
		return err
	}
	if err := r.createInstanceProfile(); err != nil {
		return err
	}
	if err := r.attachInstanceProfile(); err != nil {
		return err
	}
	return nil
}

func (r *role) Destroy(flags ...string) error {
	msg.Info("Role Deletion: %s", r.Name())
	for _, p := range r.policies {
		if err := r.detachPolicy(p); err != nil {
			if strings.Contains(err.Error(), "NoSuchEntity") {
				continue
			}
			return err
		}
	}
	params := &iam.DeleteRoleInput{
		RoleName: aws.String(r.Name()),
	}
	_, err := r.iam.DeleteRole(params)
	if err != nil {
		return err
	}
	msg.Detail("Role deleted: %s", r.Name())
	r.clear()
	return nil
}

func (r *role) Provision(flags ...string) error {
	msg.Info("Role Provision: %s", r.Name())
	params := &iam.UpdateRoleDescriptionInput{
		Description: aws.String(r.Description()),
		RoleName:    aws.String(r.Name()),
	}

	_, err := r.iam.UpdateRoleDescription(params)
	if err != nil {
		return err
	}

	if err := r.updatePolicies(); err != nil {
		return err
	}
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

func (r *role) createRole(flags ...string) error {
	msg.Info("Role Creation: %s", r.Name())
	file := fmt.Sprintf(env.Lookup("ROOT")+"/etc/arc/policies/trust_relationships/%s.json", r.TrustRelationship())
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	params := &iam.CreateRoleInput{
		AssumeRolePolicyDocument: aws.String(string(data)),
		Description:              aws.String(r.Description()),
		RoleName:                 aws.String(r.Name()),
	}

	resp, err := r.iam.CreateRole(params)
	if err != nil {
		return err
	}
	r.role = resp.Role
	if err := r.Load(); err != nil {
		return err
	}
	msg.Detail("Role Created: %s", r.Name())
	for _, p := range r.policies {
		if err := r.attachPolicy(p); err != nil {
			return err
		}
	}
	return nil
}

func (r *role) createInstanceProfile() error {
	if r.InstanceProfile() == "" {
		return nil
	}
	msg.Info("Instance Profile Creation: %s", r.InstanceProfile())
	params := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(r.InstanceProfile()),
	}

	resp, err := r.iam.CreateInstanceProfile(params)
	if err != nil {
		return err
	}
	r.instanceProfile = resp.InstanceProfile
	msg.Detail("Instance Profile Created: %s", r.InstanceProfile())
	return nil
}

func (r *role) attachInstanceProfile() error {
	if r.InstanceProfile() == "" {
		return nil
	}
	msg.Info("Attaching Instance Profile %q", r.InstanceProfile())
	params := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(r.InstanceProfile()),
		RoleName:            aws.String(r.Name()),
	}

	_, err := r.iam.AddRoleToInstanceProfile(params)
	if err != nil {
		return err
	}
	msg.Detail("Attached Instance Profile %q", r.InstanceProfile())
	return nil
}

func (r *role) attachPolicy(policyArn string) error {
	log.Debug("Attaching Policy %q", policyArn)
	params := &iam.AttachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(r.Name()),
	}
	_, err := r.iam.AttachRolePolicy(params)
	if err != nil {
		return err
	}
	log.Debug("Attached Policy %q", policyArn)
	return nil
}

func (r *role) detachPolicy(policyArn string) error {
	log.Debug("Detaching Policy %q", policyArn)
	params := &iam.DetachRolePolicyInput{
		PolicyArn: aws.String(policyArn),
		RoleName:  aws.String(r.Name()),
	}
	_, err := r.iam.DetachRolePolicy(params)
	if err != nil {
		return err
	}
	log.Debug("Detached Policy %q", policyArn)
	return nil
}

func (r *role) updatePolicies() error {
	msg.Info("Updating Policies")

	// detach any rogue policies
	roguePolicies := r.roguePolicies()
	for _, p := range roguePolicies {
		if err := r.detachPolicy(p); err != nil {
			return err
		}
	}
	// attach any configured but not deployed policies
	orphanedPolicies := r.orphanedPolicies()
	for _, p := range orphanedPolicies {
		if err := r.attachPolicy(p); err != nil {
			return err
		}
	}
	return nil
}

func (r *role) loadPolicies() error {
	next := ""
	for {
		params := &iam.ListAttachedRolePoliciesInput{
			RoleName: aws.String(r.Name()),
		}
		if next != "" {
			params.Marker = aws.String(next)
		}
		resp, err := r.iam.ListAttachedRolePolicies(params)
		if err != nil {
			return err
		}
		truncated := false
		if resp.IsTruncated != nil {
			truncated = *resp.IsTruncated
		}
		next = ""
		if resp.Marker != nil {
			next = *resp.Marker
		}
		for _, v := range resp.AttachedPolicies {
			r.deployedPolicies = append(r.deployedPolicies, *v.PolicyArn)
		}
		if truncated == false {
			break
		}
	}
	return nil
}

func (r *role) roguePolicies() []string {
	policies := []string{}
	found := false
	for _, ap := range r.deployedPolicies {
		for _, cp := range r.policies {
			if ap == cp {
				found = true
			}
		}
		if found == false {
			policies = append(policies, ap)
		}
		found = false
	}

	return policies
}

func (r *role) orphanedPolicies() []string {
	policies := []string{}
	found := false
	for _, cp := range r.policies {
		for _, ap := range r.deployedPolicies {
			if ap == cp {
				found = true
			}
		}
		if found == false {
			policies = append(policies, cp)
		}
		found = false
	}
	return policies
}
