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
	"os/user"
	"time"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
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

func newEncryptionKey(cfg *config.EncryptionKey, km *keyManagement, prov provider.KeyManagement) (*encryptionKey, error) {
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

// Route satisfies the embedded route.Router interface in resource.EncryptionKey.
// EncryptionKey handles load, create, destroy, audit, config and info requests by delegating them
// to the providerEncryptionKey.
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
	case route.Provision:
		if err := k.Provision(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Audit:
		if err := k.Audit("Encryption Key"); err != nil {
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

func (k *encryptionKey) Load() error {
	return k.providerEncryptionKey.Load()
}

func (k *encryptionKey) Create(flags ...string) error {
	if k.Created() {
		msg.Detail("Encryption Key exists, skipping...")
		return nil
	}
	if err := k.providerEncryptionKey.Create(flags...); err != nil {
		return err
	}
	if err := k.createSecurityTags(); err != nil {
		return err
	}
	return nil
}

// Created satisfies the embedded resource.Creator interface in resource.Bucket.
// It delegates the call to the provider's encryptionKey.
func (k *encryptionKey) Created() bool {
	return k.providerEncryptionKey.Created()
}

func (k *encryptionKey) Destroy(flags ...string) error {
	if k.Destroyed() {
		msg.Detail("Encryption Key does not exist, skipping...")
		return nil
	}
	return k.ProviderEncryptionKey().Destroy(flags...)
}

// Destroyed satisfies the embedded resource.Destroyer interaface in resource.Bucket.
// It delegates the call to the provider's encryptionKey.
func (k *encryptionKey) Destroyed() bool {
	return k.providerEncryptionKey.Destroyed()
}

func (k *encryptionKey) Provision(flags ...string) error {
	if k.Destroyed() {
		msg.Detail("Encryption Key does not exist, skipping...")
		return nil
	}
	if err := k.createSecurityTags(); err != nil {
		return err
	}
	return k.ProviderEncryptionKey().Provision(flags...)
}
func (k *encryptionKey) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	return k.providerEncryptionKey.Audit(flags...)
}

func (k *encryptionKey) Info() {
	k.ProviderEncryptionKey().Info()
}

func (k *encryptionKey) SetTags(t map[string]string) error {
	if k.providerEncryptionKey == nil {
		return fmt.Errorf("providerEncryptionKey not created")
	}
	return k.providerEncryptionKey.SetTags(t)
}

func (k *encryptionKey) createSecurityTags() error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	tags := map[string]string{
		"Name":          k.Name(),
		"Created By":    u.Username,
		"Last Modified": time.Now().UTC().String(),
	}
	for k, v := range k.KeyManagement().Amp().SecurityTags() {
		tags[k] = v
	}
	for k, v := range k.SecurityTags() {
		tags[k] = v
	}
	return k.SetTags(tags)
}

func (k *encryptionKey) Help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: fmt.Sprintf("create encryption key %s", k.Name())},
		{Name: route.Destroy.String(), Desc: fmt.Sprintf("destroy encryption key %s", k.Name())},
		{Name: route.Audit.String(), Desc: fmt.Sprintf("audit encrypption key %s", k.Name())},
		{Name: route.Info.String(), Desc: "show information about allocated encryption key"},
		{Name: route.Config.String(), Desc: "show the configuration for the given encryption key"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("encryption_key", commands)
}

func (k *encryptionKey) KeyManagement() resource.KeyManagement {
	return k.keyManagement
}

func (k *encryptionKey) ProviderEncryptionKey() resource.ProviderEncryptionKey {
	return k.providerEncryptionKey
}
