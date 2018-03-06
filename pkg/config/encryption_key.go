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

type EncryptionKey struct {
	Name_                  string       `json:"encryption_key"`
	Region_                string       `json:"region"`
	DeletionPendingWindow_ int          `json:"deletion_pending_window"`
	SecurityTags_          SecurityTags `json:"security_tags"`
}

func (k *EncryptionKey) Name() string {
	return k.Name_
}

func (k *EncryptionKey) Region() string {
	return k.Region_
}

func (k *EncryptionKey) DeletionPendingWindow() int {
	return k.DeletionPendingWindow_
}

func (k *EncryptionKey) SecurityTags() SecurityTags {
	return k.SecurityTags_
}

func (k *EncryptionKey) Print() {
	msg.Info("Bucket Config")
	msg.Detail("%-20s\t%s", "name", k.Name())
	msg.Detail("%-20s\t%s", "region", k.Region())
	msg.Detail("%-20s\t%d", "deletiong pending window", k.DeletionPendingWindow())
}
