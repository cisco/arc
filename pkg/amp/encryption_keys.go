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

package amp

import (
	"fmt"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type encryptionKeys struct {
	*resource.Resources
	encryptionKeys map[string]resource.EncryptionKey
	keyManagement  *keyManagement
}

func newEncryptionKeys(km *keyManagement, prov provider.Account, cfg *config.EncryptionKeys) (*encryptionKeys, error) {
	log.Debug("Initializing Encryption Keys")

	k := &encryptionKeys{
		Resources:      resource.NewResources(),
		encryptionKeys: map[string]resource.EncryptionKey{},
		keyManagement:  km,
	}

	for _, conf := range *cfg {
		if k.Find(conf.Name()) != nil {
			return nil, fmt.Errorf("Encryption Key name %q must be unique, but it is used multiple times", conf.Name())
		}
		encryptionKey, err := newEncryptionKey(km, prov, conf)
		if err != nil {
			return nil, err
		}
		k.encryptionKeys[conf.Name()] = encryptionKey
		k.Append(encryptionKey)
	}
	return k, nil
}

func (k *encryptionKeys) Find(name string) resource.EncryptionKey {
	return k.encryptionKeys[name]
}

func (k *encryptionKeys) Route(req *route.Request) route.Response {
	log.Route(req, "Encryption Keys")
	return k.RouteInOrder(req)
}

func (k *encryptionKeys) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	for _, ek := range k.encryptionKeys {
		if err := ek.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}
