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
	"github.com/cisco/arc/pkg/log"
)

type encryptionKeyCacheEntry struct {
	deployed   *kms.AliasListEntry
	configured *encryptionKey
}

type encryptionKeyCache struct {
	cache   map[string]*encryptionKeyCacheEntry
	unnamed []*kms.AliasListEntry
}

func newEncryptionKeyCache(k *keyManagement) (*encryptionKeyCache, error) {
	log.Debug("Initializing AWS Encryption Key Cache")

	c := &encryptionKeyCache{
		cache: map[string]*encryptionKeyCacheEntry{},
	}

	params := &kms.ListAliasesInput{}

	resp, err := k.kms.ListAliases(params)
	if err != nil {
		return nil, err
	}

	for _, k := range resp.Aliases {
		if k.AliasArn == nil {
			log.Verbose("Unnamed encryption key")
			c.unnamed = append(c.unnamed, k)
			continue
		}
		log.Debug("Caching %s", aws.StringValue(k.AliasName))
		c.cache[aws.StringValue(k.AliasName)] = &encryptionKeyCacheEntry{deployed: k}
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
		if v.configured == nil {
			a.Audit(aaa.Deployed, "%s", k)
		}
	}
	if c.unnamed != nil {
		a.Audit(aaa.Deployed, "\r")
		for i := range c.unnamed {
			m := fmt.Sprintf("Unnamed Encryption Key %d ", i+1)
			a.Audit(aaa.Deployed, m)
		}
	}
	return nil
}
