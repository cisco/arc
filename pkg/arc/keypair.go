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

package arc

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh/agent"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type keypair struct {
	*config.KeyPair
	providerKeyPair resource.ProviderKeyPair
}

// newKeyPair is the constructor for a keypair object. It returns a non-nil error upon failure.
func newKeyPair(p provider.DataCenter) (*keypair, error) {
	log.Debug("Initializing KeyPair")

	// Lookup the ssh user name
	username := env.Lookup("SSH_USER")
	filename := "id_rsa"
	if username != env.Lookup("USER") {
		filename = username
	}

	// Connect to the ssh-agent to get the key material.
	conn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	sshAgent := agent.NewClient(conn)
	keys, err := sshAgent.List()
	if err != nil {
		return nil, err
	}

	// We will look at all the keys available from the ssh-agent looking for the id_rsa key.
	// To avoid key collisions in the provider, the id_rsa key will be created as the user's
	// username instead of id_rsa.
	var cfg *config.KeyPair
	for _, key := range keys {
		_, file := filepath.Split(key.Comment)
		if file == filename {
			cfg = &config.KeyPair{
				Name_:        username,
				LocalName_:   file,
				Format_:      key.Format,
				Comment_:     key.Comment,
				KeyMaterial_: key.String(),
			}
			break
		}
	}
	if cfg == nil {
		return nil, fmt.Errorf("Cannot find id_rsa. Is it available from ssh-agent?")
	}

	providerKeyPair, err := p.NewKeyPair(cfg)
	if err != nil {
		return nil, err
	}

	return &keypair{
		KeyPair:         cfg,
		providerKeyPair: providerKeyPair,
	}, nil
}

func (k *keypair) Created() bool {
	return k.providerKeyPair.Created()
}

func (k *keypair) Destroyed() bool {
	return k.providerKeyPair.Destroyed()
}

func (k *keypair) FingerPrint() string {
	return k.providerKeyPair.FingerPrint()
}

func (k *keypair) Route(req *route.Request) route.Response {
	log.Route(req, "Keypair %q", k.Name())

	if req.Top() != "" {
		k.help()
		return route.FAIL
	}

	if err := aaa.Authorized(req, "keypair", k.Name()); err != nil {
		msg.Error(err.Error())
		return route.UNAUTHORIZED
	}

	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load:
		if err := k.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Create, route.Destroy:
		return k.providerKeyPair.Route(req)
	case route.Help:
		k.help()
	case route.Config:
		k.config()
	case route.Info:
		k.info(req)
	default:
		msg.Error("Unknown keypair command %q.", req.Command().String())
		return route.FAIL
	}
	return route.OK
}

func (k *keypair) Load() error {
	return k.providerKeyPair.Load()
}

func (k *keypair) help() {
	commands := []help.Command{
		{route.Create.String(), "create the keypair"},
		{route.Destroy.String(), "destroy the keypair"},
		{route.Config.String(), "provide the configuration for the keypair"},
		{route.Info.String(), "provide information about the keypair"},
		{route.Help.String(), "provide this help"},
	}
	help.Print("keypair", commands)
}

func (k *keypair) config() {
	k.KeyPair.Print()
}

func (k *keypair) info(req *route.Request) {
	if k.Destroyed() {
		return
	}
	msg.Info("KeyPair")
	msg.IndentInc()
	k.PrintLocal()
	k.providerKeyPair.Route(req)
	msg.IndentDec()
}
