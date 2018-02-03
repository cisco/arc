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
	"github.com/cisco/arc/pkg/config"
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
	encryptionKeys        *encryptionKeys
	providerKeyManagement resource.ProviderKeyManagement
}

// newKeyManagement is the constructor for a keyManagement object. It returns a non-nil error upon failure.
func newKeyManagement(account *account, prov provider.KeyManagement, cfg *config.KeyManagement) (*keyManagement, error) {
	log.Debug("Initializing Key Management")

	k := &keyManagement{
		Resources:     resource.NewResources(),
		KeyManagement: cfg,
		account:       account,
	}

	var err error
	k.providerKeyManagement, err = prov.NewKeyManagement(cfg)
	if err != nil {
		return nil, err
	}

	k.encryptionKeys, err = newEncryptionKeys(k, prov, cfg.EncryptionKeys())
	if err != nil {
		return nil, err
	}
	k.Append(k.encryptionKeys)

	return k, nil
}

// Account satisfies the resource.KeyManagement interface and provides access
// to storage's parent.
func (k *keyManagement) Account() resource.Account {
	return k.account
}

func (k *keyManagement) ProviderKeyManagement() resource.ProviderKeyManagement {
	return k.providerKeyManagement
}

// Route satisfies the embedded resource.Resource interface in resource.Storage.
// Storage does not directly terminate a request so only handles load and info
// requests from it's parent.  All other commands are routed to amp's children.
func (k *keyManagement) Route(req *route.Request) route.Response {
	log.Route(req, "Key Management")

	switch req.Top() {
	case "":
		break
	case "key", "encryption_key":
		req.Pop()
		key := k.encryptionKeys.Find(req.Top())
		if key == nil {
			msg.Error("Unknown Encryption Key %q.", key)
			return route.FAIL
		}
		key.Route(req)
	}
	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Info:
		k.info(req)
		return route.OK
	case route.Config:
		k.config(req)
		return route.OK
	case route.Load:
		return k.RouteInOrder(req)
	case route.Provision:
		return k.RouteInOrder(req)
	default:
		panic("Internal Error: Unknown command " + req.Command().String())
	}
	return route.FAIL
}

func (k *keyManagement) info(req *route.Request) {
	if k.Destroyed() {
		return
	}
	msg.Info("Key Management")
	msg.IndentInc()
	k.RouteInOrder(req)
	msg.IndentDec()
}

func (k *keyManagement) config(req *route.Request) {
	if k.Destroyed() {
		return
	}
	msg.Info("Key Management")
	msg.IndentInc()
	k.RouteInOrder(req)
	msg.IndentDec()
}

func (k *keyManagement) Audit(flags ...string) error {
	return nil
}
