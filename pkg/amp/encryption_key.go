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

package amp

import (
	"fmt"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type encryptionKey struct {
	*config.EncryptionKey
	keyManagement         resource.KeyManagement
	providerEncryptionKey resource.ProviderEncryptionKey
}

func newEncryptionKey(km *keyManagement, prov provider.Account, cfg *config.EncryptionKey) (*encryptionKey, error) {
	log.Debug("Initializing Encryption Key, %q", cfg.Name())
	k := &encryptionKey{
		EncryptionKey: cfg,
		keyManagement: km,
	}

	var err error
	k.providerEncryptionKey, err = prov.NewEncryptionKey(k, cfg)
	if err != nil {
		return nil, err
	}

	return k, nil
}

// Route satisfies the embedded resource.Resource interface in resource.Bucket.
// Bucket handles load, create, destroy, config and info requests by delegating them
// to the providerBucket.
func (k *encryptionKey) Route(req *route.Request) route.Response {
	log.Route(req, "Encryption Key %q", k.Name())
	switch req.Command() {
	case route.Create:
		if err := k.Create(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Destroy:
		if err := k.Destroy(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Info:
		k.Info()
		return route.OK
	case route.Config:
		k.Print()
		return route.OK
	default:
		panic("Internal Error: Unknown command " + req.Command().String())
	}
}

// Created satisfies the embedded resource.Resource interface in resource.Bucket.
// It delegates the call to the provider's bucket.
func (k *encryptionKey) Created() bool {
	return k.providerEncryptionKey.Created()
}

// Destroyed satisfies the embedded resource.Resource interaface in resource.Bucket.
// It delegates the call to the provider's bucket.
func (k *encryptionKey) Destroyed() bool {
	return k.providerEncryptionKey.Destroyed()
}

func (k *encryptionKey) KeyManagement() resource.KeyManagement {
	return k.keyManagement
}

func (k *encryptionKey) ProviderEncryptionKey() resource.ProviderEncryptionKey {
	return k.providerEncryptionKey
}

func (k *encryptionKey) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	return k.providerEncryptionKey.Audit(flags...)
}

func (k *encryptionKey) Create(flags ...string) error {
	if k.Created() {
		msg.Detail("Encryption Key exists, skipping...")
		return nil
	}
	if err := k.providerEncryptionKey.Create(flags...); err != nil {
		return err
	}
	return nil
}

func (k *encryptionKey) Destroy(flags ...string) error {
	if k.Destroyed() {
		msg.Detail("Encryption Key does not exist, skipping...")
		return nil
	}
	return k.ProviderEncryptionKey().Destroy(flags...)
}

func (k *encryptionKey) Info() {
	k.ProviderEncryptionKey().Info()
}
