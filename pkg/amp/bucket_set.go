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
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type bucketSet struct {
	*resource.Resources
	*config.BucketSet
	bucketSets []resource.BucketSet

	buckets []resource.Bucket
}

func newBucketSet(cfg *config.BucketSet, s *storage, prov provider.Storage) (*bucketSet, error) {
	log.Debug("Initializing Bucket Set, %q", cfg.Name())
	b := &bucketSet{
		BucketSet: cfg,
	}

	for _, conf := range cfg.Buckets() {
		bucket, err := newBucket(conf, s, prov)
		if err != nil {
			return nil, err
		}
		b.buckets = append(b.buckets, bucket)
	}
	return b, nil
}

// Route satisfies the embedded resource.Resource interface in resource.Bucket.
// Bucket handles load, create, destroy, config and info requests by delegating them
// to the providerBucket.
func (b *bucketSet) Route(req *route.Request) route.Response {
	log.Route(req, "Bucket %q", b.Name())

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
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
	case route.Info:
		b.Info()
		return route.OK
	}
	return b.RouteInOrder(req)
}

// Created satisfies the embedded resource.Resource interface in resource.Bucket.
// It delegates the call to the provider's bucket.
func (b *bucketSet) Created() bool {
	for _, bkt := range b.buckets {
		if bkt.Created() == false {
			return false
		}
	}
	return true
}

// Destroyed satisfies the embedded resource.Resource interaface in resource.Bucket.
// It delegates the call to the provider's bucket.
func (b *bucketSet) Destroyed() bool {
	for _, bkt := range b.buckets {
		if bkt.Destroyed() == false {
			return false
		}
	}
	return true
}

func (b *bucketSet) BucketSets() []resource.BucketSet {
	return b.bucketSets
}

func (b *bucketSet) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	for _, bkt := range b.buckets {
		if err := bkt.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (b *bucketSet) Create(flags ...string) error {
	if b.Created() {
		msg.Detail("Bucket Set exists, skipping...")
		return nil
	}
	for _, bkt := range b.buckets {
		if err := bkt.Create(flags...); err != nil {
			return err
		}
	}
	for _, bkt := range b.buckets {
		if err := bkt.EnableReplication(); err != nil {
			return err
		}
	}
	return nil
}

func (b *bucketSet) Destroy(flags ...string) error {
	if b.Destroyed() {
		msg.Detail("Bucket Set does not exist, skipping...")
		return nil
	}
	for _, bkt := range b.buckets {
		if err := bkt.Destroy(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (b *bucketSet) Provision(flags ...string) error {
	if b.Destroyed() {
		msg.Detail("Bucket Set does not exist, skipping...")
		return nil
	}
	for _, bkt := range b.buckets {
		if err := bkt.Provision(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (b *bucketSet) Info() {
	if b.Destroyed() {
		return
	}
	msg.Info("Bucket Set")
	msg.Detail("%-20s\t%s", "name", b.Name())
	msg.IndentInc()
	for _, bkt := range b.buckets {
		bkt.Info()
	}
	msg.IndentDec()
}
