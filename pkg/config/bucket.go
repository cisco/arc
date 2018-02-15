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

package config

import "github.com/cisco/arc/pkg/msg"

type Bucket struct {
	Name_         string       `json:"bucket"`
	Region_       string       `json:"region"`
	SecurityTags_ SecurityTags `json:"security_tags"`
	Role_         string       `json:"role"`
	Destination_  string       `json:"Destination"`
}

func (b *Bucket) Name() string {
	return b.Name_
}

func (b *Bucket) Region() string {
	return b.Region_
}

func (b *Bucket) SecurityTags() SecurityTags {
	return b.SecurityTags_
}

func (b *Bucket) Role() string {
	return b.Role_
}

func (b *Bucket) Destination() string {
	return b.Destination_
}

func (b *Bucket) Print() {
	msg.Info("Bucket Config")
	msg.Detail("%-20s\t%s", "name", b.Name())
	msg.Detail("%-20s\t%s", "region", b.Region())
	msg.IndentInc()
	if b.SecurityTags() != nil {
		b.SecurityTags().Print()
	}
	msg.IndentDec()
}
