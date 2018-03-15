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

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type keyManagement struct {
	*resource.Resources
	*config.KeyManagement
	amp                   *amp
	encryptionKeys        []*encryptionKey
	providerKeyManagement resource.ProviderKeyManagement
}

// newKeyManagement is the constructor for a keyManagement object. It returns a non-nil error upon failure.
func newKeyManagement(amp *amp, cfg *config.KeyManagement) (*keyManagement, error) {
	log.Debug("Initializing Key Management")

	k := &keyManagement{
		Resources:     resource.NewResources(),
		KeyManagement: cfg,
		amp:           amp,
	}

	prov, err := provider.NewKeyManagement(amp.Amp)
	if err != nil {
		return nil, err
	}
	k.providerKeyManagement, err = prov.NewKeyManagement(cfg)
	if err != nil {
		return nil, err
	}

	for _, conf := range cfg.EncryptionKeys {
		key, err := newEncryptionKey(conf, k, prov)
		if err != nil {
			return nil, err
		}
		k.encryptionKeys = append(k.encryptionKeys, key)
	}

	return k, nil
}

// Route satisfies the embedded route.Router interface in resource.KeyManagement.
// KeyManagement terminates Load, Audit, Info, and Config commands. All other
// commands are routed to keyManagement's children.
func (k *keyManagement) Route(req *route.Request) route.Response {
	log.Route(req, "Key Management")

	switch req.Top() {
	case "":
		break
	case "key", "encryption_key":
		req.Pop()
		key := k.FindEncryptionKey(req.Top())
		if key == nil {
			msg.Error("Unknown Encryption Key %q.", key)
			return route.FAIL
		}
		return key.Route(req)
	}
	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Load:
		if err := k.Load(); err != nil {
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
	case route.Help:
		k.Help()
		return route.OK
	default:
		msg.Error("Error: amp/key_management.go Unknown command " + req.Command().String())
		k.Help()
		return route.FAIL
	}
}

func (k *keyManagement) Load() error {
	for _, key := range k.encryptionKeys {
		if err := key.Load(); err != nil {
			return err
		}
	}
	return nil
}

func (k *keyManagement) Provision(flags ...string) error {
	for _, key := range k.encryptionKeys {
		if err := key.Provision(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (k *keyManagement) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	err := aaa.NewAudit(flags[0])
	if err != nil {
		return err
	}

	if err = k.ProviderKeyManagement().Audit(flags...); err != nil {
		return err
	}

	for _, key := range k.encryptionKeys {
		if err := key.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (k *keyManagement) Info() {
	msg.Info("Key Management")
	msg.IndentInc()
	for _, key := range k.encryptionKeys {
		key.Info()
	}
	msg.IndentDec()
}

func (k *keyManagement) Help() {
	commands := []help.Command{
		{Name: route.Audit.String(), Desc: "audit the storage"},
		{Name: route.Info.String(), Desc: "show information about allocated storage"},
		{Name: route.Config.String(), Desc: "show the configuration for the given storage"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("key_management", commands)
}

// Amp satisfies the resource.KeyManagement interface and provides access
// to keyManagement's parent.
func (k *keyManagement) Amp() resource.Amp {
	return k.amp
}

// FindEncryptionKey returns the  EncryptionKey with the given name
func (k *keyManagement) FindEncryptionKey(name string) resource.EncryptionKey {
	for _, key := range k.encryptionKeys {
		if key.Name() == name {
			return key
		}
	}
	return nil
}

// ProviderStorage provides access to the provider KeyManagement object.
func (k *keyManagement) ProviderKeyManagement() resource.ProviderKeyManagement {
	return k.providerKeyManagement
}
