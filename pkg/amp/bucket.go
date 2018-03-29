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
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type bucket struct {
	*config.Bucket
	storage        *storage
	providerBucket resource.ProviderBucket
}

func newBucket(cfg *config.Bucket, s *storage, prov provider.Storage) (*bucket, error) {
	log.Debug("Initializing Bucket, %q", cfg.Name())
	b := &bucket{
		Bucket:  cfg,
		storage: s,
	}

	var err error
	b.providerBucket, err = prov.NewBucket(b, cfg)
	if err != nil {
		return nil, err
	}

	return b, nil
}

// Route satisfies the embedded resource.Resource interface in resource.Bucket.
// Bucket handles load, create, destroy, config and info requests by delegating them
// to the providerBucket.
func (b *bucket) Route(req *route.Request) route.Response {
	log.Route(req, "Bucket %q", b.Name())
	switch req.Command() {
	case route.Load:
		if err := b.Load(); err != nil {
			return route.FAIL
		}
		return route.OK
	case route.Create:
		if err := b.Create(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Destroy:
		if err := b.Destroy(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Provision:
		if err := b.Provision(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Audit:
		if err := b.Audit(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Info:
		b.Info()
		return route.OK
	case route.Config:
		b.Print()
		return route.OK
	case route.Help:
		b.Help()
		return route.OK
	default:
		msg.Error("Internal Error: Unknown command " + req.Command().String())
		b.Help()
		return route.FAIL
	}
}

// Created satisfies the embedded resource.Resource interface in resource.Bucket.
// It delegates the call to the provider's bucket.
func (b *bucket) Created() bool {
	return b.providerBucket.Created()
}

// Destroyed satisfies the embedded resource.Resource interaface in resource.Bucket.
// It delegates the call to the provider's bucket.
func (b *bucket) Destroyed() bool {
	return b.providerBucket.Destroyed()
}

func (b *bucket) Storage() resource.Storage {
	return b.storage
}

func (b *bucket) ProviderBucket() resource.ProviderBucket {
	return b.providerBucket
}

func (b *bucket) Load() error {
	return b.providerBucket.Load()
}

func (b *bucket) Create(flags ...string) error {
	if b.Created() {
		msg.Detail("Bucket exists, skipping...")
		return nil
	}
	if err := b.providerBucket.Create(flags...); err != nil {
		return err
	}
	if err := b.createSecurityTags(); err != nil {
		return err
	}
	if err := b.enableEncryption(); err != nil {
		return err
	}
	return nil
}

func (b *bucket) Destroy(flags ...string) error {
	if b.Destroyed() {
		msg.Detail("Bucket does not exist, skipping...")
		return nil
	}
	return b.ProviderBucket().Destroy(flags...)
}

func (b *bucket) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	return b.providerBucket.Audit(flags...)
}

func (b *bucket) Provision(flags ...string) error {
	if b.Destroyed() {
		msg.Detail("Bucket does not exist, skipping...")
		return nil
	}
	tagsFlagSet := false
	encryptFlagSet := false
	if len(flags) != 0 {
		for _, v := range flags {
			if v == "tags" {
				tagsFlagSet = true
			}
			if v == "encrypt" {
				encryptFlagSet = true
			}
		}
	}
	if encryptFlagSet {
		if err := b.enableEncryption(); err != nil {
			return err
		}
		return nil
	}
	if tagsFlagSet {
		if err := b.createSecurityTags(); err != nil {
			return err
		}
		return nil
	}
	if err := b.enableEncryption(); err != nil {
		return err
	}
	if err := b.createSecurityTags(); err != nil {
		return err
	}
	return nil
}

func (b *bucket) SetTags(t map[string]string) error {
	if b.providerBucket == nil {
		return fmt.Errorf("providerBucket not created")
	}
	return b.providerBucket.SetTags(t)
}

func (b *bucket) createSecurityTags() error {
	tags := map[string]string{}
	for k, v := range b.Storage().Amp().SecurityTags() {
		tags[k] = v
	}
	for k, v := range b.SecurityTags() {
		tags[k] = v
	}
	return b.SetTags(tags)
}

func (b *bucket) Info() {
	if b.Destroyed() {
		return
	}
	b.ProviderBucket().Info()
}

func (b *bucket) Help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: fmt.Sprintf("create bucket %s", b.Name())},
		{Name: route.Destroy.String(), Desc: fmt.Sprintf("destroy bucket %s", b.Name())},
		{Name: route.Provision.String(), Desc: fmt.Sprintf("update the tags for %s", b.Name())},
		{Name: route.Audit.String(), Desc: fmt.Sprintf("audit bucket %s", b.Name())},
		{Name: route.Info.String(), Desc: "show information about allocated bucket"},
		{Name: route.Config.String(), Desc: "show the configuration for the given bucket"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("bucket", commands)
}

func (b *bucket) enableReplication() error {
	keyName := b.DestinationEncryptionKey()
	if keyName == "" {
		return nil
	}
	key := b.Storage().Amp().KeyManagement().FindEncryptionKey(keyName)
	return b.providerBucket.EnableReplication(key)
}

func (b *bucket) enableEncryption() error {
	keyName := b.EncryptionKey()
	if keyName == "" {
		keyName = b.storage.EncryptionKey()
	}
	if keyName == "" {
		return nil
	}
	key := b.Storage().Amp().KeyManagement().FindEncryptionKey(keyName)
	if key != nil {
		return b.ProviderBucket().EnableEncryption(key)
	}
	return nil
}
