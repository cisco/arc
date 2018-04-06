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
	"github.com/aws/aws-sdk-go/service/iam"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/resource"
)

type identityManagement struct {
	*config.IdentityManagement
	iam *iam.IAM

	roleCache   *roleCache
	policyCache *policyCache
}

func newIdentityManagement(cfg *config.IdentityManagement, iam *iam.IAM) (resource.ProviderIdentityManagement, error) {
	log.Debug("Initializing AWS Identity Management")

	i := &identityManagement{
		IdentityManagement: cfg,
		iam:                iam,
	}

	var err error
	i.roleCache, err = newRoleCache(i)
	if err != nil {
		return nil, err
	}
	i.policyCache, err = newPolicyCache(i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func (i *identityManagement) Audit(flags ...string) error {
	if err := i.policyCache.audit("Policy"); err != nil {
		return err
	}
	if err := i.roleCache.audit("Role"); err != nil {
		return err
	}
	return nil
}
