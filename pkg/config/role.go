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

package config

import "github.com/cisco/arc/pkg/msg"

type Role struct {
	Name_              string   `json:"role"`
	TrustRelationship_ string   `json:"trust_relationship"`
	InstanceProfile_   string   `json:"instance_profile"`
	Description_       string   `json:"description"`
	Policies_          []string `json:"policies"`
}

func (r *Role) Name() string {
	return r.Name_
}

func (r *Role) TrustRelationship() string {
	return r.TrustRelationship_
}

func (r *Role) InstanceProfile() string {
	return r.InstanceProfile_
}

func (r *Role) Description() string {
	return r.Description_
}

func (r *Role) Policies() []string {
	return r.Policies_
}

func (r *Role) Print() {
	msg.Info("Role Config")
	msg.Detail("%-20s\t%s", "name", r.Name())
	msg.Detail("%-20s\t%s", "trust relationship", r.TrustRelationship())
	msg.Detail("%-20s\t%s", "description", r.Description())
	for _, p := range r.Policies() {
		msg.Detail("%-20s\t%s", "policy", p)
	}
}
