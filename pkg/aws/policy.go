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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

type policy struct {
	*config.Policy
	provider *identityManagementProvider
	iam      *iam.IAM

	identityManagement *identityManagement

	policy *iam.Policy
}

func newPolicy(pol resource.Policy, cfg *config.Policy, prov *identityManagementProvider) (resource.ProviderPolicy, error) {
	log.Debug("Initializing AWS Policy %q", cfg.Name())

	p := &policy{
		Policy:             cfg,
		provider:           prov,
		iam:                prov.iam,
		identityManagement: pol.IdentityManagement().ProviderIdentityManagement().(*identityManagement),
	}
	p.set(p.identityManagement.policyCache.find(p))

	return p, nil
}

func (p *policy) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit Object does not exist")
	}
	if p.policy == nil {
		a.Audit(aaa.Configured, "%s", p.Name())
	}
	return nil
}

func (p *policy) set(policy *iam.Policy) {
	p.policy = policy
}

func (p *policy) clear() {
	p.policy = nil
}

func (p *policy) Created() bool {
	return p.policy != nil
}

func (p *policy) Destroyed() bool {
	return p.policy == nil
}

func (p *policy) Create(flags ...string) error {
	msg.Info("Policy Creation: %s", p.Name())
	file := fmt.Sprintf(env.Lookup("ROOT")+"/etc/arc/IAM_policies/IAM_policies/%s.json", p.PolicyDocument())
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	params := &iam.CreatePolicyInput{
		Description:    aws.String(p.Description()),
		PolicyDocument: aws.String(string(data)),
		PolicyName:     aws.String(p.Name()),
	}

	resp, err := p.iam.CreatePolicy(params)
	if err != nil {
		return err
	}
	p.policy = resp.Policy
	if err := p.Load(); err != nil {
		return err
	}
	msg.Detail("Policy created: %s", p.Name())
	return nil
}

func (p *policy) Load() error {
	if p.Created() {
		log.Debug("Skipping Policy load, cached...")
		return nil
	}
	if p.policy == nil {
		return fmt.Errorf("Could not find policy on AWS")
	}
	return nil
}

func (p *policy) Destroy(flags ...string) error {
	msg.Info("Policy Deletion: %s", p.Name())
	arn := p.policy.Arn
	params := &iam.DeletePolicyInput{
		PolicyArn: arn,
	}
	_, err := p.iam.DeletePolicy(params)
	if err != nil {
		return err
	}
	msg.Detail("Policy deleted: %s", p.Name())
	p.clear()
	return nil
}

func (p *policy) Info() {
	if p.Destroyed() {
		return
	}
	msg.Info("Policy")
	msg.Detail("%-20s\t%s", "name", aws.StringValue(p.policy.PolicyName))
	msg.Detail("%-20s\t%+v", "policy", p.policy)
}
