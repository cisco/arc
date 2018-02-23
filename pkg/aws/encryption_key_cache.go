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
	"github.com/aws/aws-sdk-go/service/kms"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/log"
)

type encryptionKeyCacheEntry struct {
	deployed   *kms.AliasListEntry
	configured *encryptionKey
}

type encryptionKeyCache struct {
	cache   map[string]*encryptionKeyCacheEntry
	unnamed []*kms.KeyListEntry
}

func newEncryptionKeyCache(k *keyManagement) (*encryptionKeyCache, error) {
	log.Debug("Initializing AWS Encryption Key Cache")

	c := &encryptionKeyCache{
		cache: map[string]*encryptionKeyCacheEntry{},
	}

	aliasList, err := listAliases(k)
	if err != nil {
		return nil, err
	}

	keyList, err := listKeys(k)
	if err != nil {
		return nil, err
	}

	for keyId, entry := range keyList {
		if aliasList[keyId] == nil {
			log.Verbose("Unnamed encryption key")
			c.unnamed = append(c.unnamed, entry)
		}
	}

	for _, entry := range aliasList {
		c.cache[aws.StringValue(entry.AliasName)] = &encryptionKeyCacheEntry{deployed: entry}
	}

	return c, nil
}

func (c *encryptionKeyCache) find(k *encryptionKey) *kms.AliasListEntry {
	e := c.cache["alias/"+k.Name()]
	if e == nil {
		return nil
	}
	e.configured = k
	return e.deployed
}

func (c *encryptionKeyCache) remove(k *encryptionKey) {
	log.Debug("Deleting %s from encryptionKeyCache", k.Name())
	delete(c.cache, k.Name())
}

func (c *encryptionKeyCache) audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit Object does not exist")
	}
	for k, v := range c.cache {
		if v.configured == nil && !strings.Contains(aws.StringValue(v.deployed.AliasName), "aws") {
			a.Audit(aaa.Deployed, "%s", k)
		}
	}
	if c.unnamed != nil {
		a.Audit(aaa.Deployed, "\r")
		for _, v := range c.unnamed {
			m := fmt.Sprintf("Unnamed Encryption Key %s ", aws.StringValue(v.KeyArn))
			a.Audit(aaa.Deployed, m)
		}
	}
	return nil
}

func listAliases(k *keyManagement) (map[string]*kms.AliasListEntry, error) {
	aliasList := map[string]*kms.AliasListEntry{}
	params := &kms.ListAliasesInput{}

	resp, err := k.kms.ListAliases(params)
	if err != nil {
		return nil, err
	}

	for _, key := range resp.Aliases {
		log.Debug("Caching %s", aws.StringValue(key.AliasName))
		aliasList[aws.StringValue(key.TargetKeyId)] = key
	}
	return aliasList, nil
}

func listKeys(k *keyManagement) (map[string]*kms.KeyListEntry, error) {
	keyList := map[string]*kms.KeyListEntry{}
	params := &kms.ListKeysInput{}

	resp, err := k.kms.ListKeys(params)
	if err != nil {
		return nil, err
	}

	for _, key := range resp.Keys {
		keyList[aws.StringValue(key.KeyId)] = key
	}

	return keyList, nil
}
