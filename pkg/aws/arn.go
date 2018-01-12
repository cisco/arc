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

import "strings"

type arn struct {
	service    string
	region     string
	namespace  string
	relativeId string
}

func newIamRole(ns string, role string) *arn {
	return &arn{
		service:    "iam",
		namespace:  ns,
		relativeId: "role/" + role,
	}
}

func newIamInstanceProfile(ns string, role string) *arn {
	return &arn{
		service:    "iam",
		namespace:  ns,
		relativeId: "instance-profile/" + role,
	}
}

func parseIamInstanceProfile(s *string) *arn {
	if s == nil {
		return nil
	}
	field := strings.Split(*s, ":")
	if len(field) != 6 {
		return nil
	}
	if field[2] != "iam" {
		return nil
	}
	path := strings.Split(field[5], "/")
	if len(path) != 2 {
		return nil
	}
	if path[0] != "instance-profile" {
		return nil
	}
	return &arn{
		service:    field[2],
		region:     field[3],
		namespace:  field[4],
		relativeId: field[5],
	}
}

func (a *arn) String() string {
	return "arn:aws:" + a.service + ":" + a.region + ":" + a.namespace + ":" + a.relativeId
}
