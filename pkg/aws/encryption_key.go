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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kms"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

type encryptionKey struct {
	*config.EncryptionKey
	kms *kms.KMS

	keyManagement *keyManagement

	encryptionKey *kms.KeyListEntry
}

func newEncryptionKey(key resource.EncryptionKey, cfg *config.EncryptionKey, prov *keyManagementProvider) (resource.ProviderEncryptionKey, error) {
	log.Debug("Initializing AWS Encryption Key %q", cfg.Name())

	k := &encryptionKey{
		EncryptionKey: cfg,
		keyManagement: key.KeyManagement().ProviderKeyManagement().(*keyManagement),
		kms:           prov.kms,
	}
	k.set(k.keyManagement.encryptionKeyCache.find(k))

	return k, nil
}

func (k *encryptionKey) Info() {
	if k.encryptionKey == nil {
		return
	}
	msg.Info("Enryption Key")
	msg.Detail("%-20s\t%s", "Arn: ", aws.StringValue(k.encryptionKey.KeyArn))
	msg.Detail("%-20s\t%s", "Id:", aws.StringValue(k.encryptionKey.KeyId))
}

func (k *encryptionKey) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit Object does not exist")
	}
	if k.encryptionKey == nil {
		a.Audit(aaa.Configured, "%s", k.Name())
	}
	return nil
}

func (k *encryptionKey) set(encryptionKey *kms.KeyListEntry) {
	k.encryptionKey = encryptionKey
}

func (k *encryptionKey) clear() {
	k.encryptionKey = nil
}

func (k *encryptionKey) Created() bool {
	return k.encryptionKey != nil
}

func (k *encryptionKey) Destroyed() bool {
	return k.encryptionKey == nil
}

func (k *encryptionKey) Create(flags ...string) error {
	msg.Info("Encryption Key Create: %s", k.Name())
	params := &kms.CreateKeyInput{}
	resp, err := k.kms.CreateKey(params)
	if err != nil {
		return err
	}
	key := &kms.KeyListEntry{
		KeyArn: resp.KeyMetadata.Arn,
		KeyId:  resp.KeyMetadata.KeyId,
	}
	k.set(key)
	return nil
}

func (k *encryptionKey) Destroy(flags ...string) error {
	msg.Info("Encryption Key Deletion: %s", k.Name())
	k.clear()
	k.keyManagement.encryptionKeyCache.remove(k)
	return nil
}
