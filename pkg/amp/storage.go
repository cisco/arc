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

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"

	_ "github.com/cisco/arc/pkg/aws"
)

type storage struct {
	*resource.Resources
	*config.Storage
	amp             *amp
	buckets         []*bucket
	bucketSets      []*bucketSet
	providerStorage resource.ProviderStorage
}

// newStorage is the constructor for a storage object. It returns a non-nil error upon failure.
func newStorage(amp *amp, cfg *config.Storage) (*storage, error) {
	log.Debug("Initializing Storage")

	// Validate the config.Storage object.
	if cfg.Buckets == nil {
		return nil, fmt.Errorf("The buckets element is missing from the storage configuration")
	}

	s := &storage{
		Resources: resource.NewResources(),
		Storage:   cfg,
		amp:       amp,
	}

	prov, err := provider.NewStorage(amp.Amp)
	if err != nil {
		return nil, err
	}
	s.providerStorage, err = prov.NewStorage(cfg)
	if err != nil {
		return nil, err
	}
	for _, conf := range cfg.Buckets {
		bucket, err := newBucket(conf, s, prov)
		if err != nil {
			return nil, err
		}
		s.buckets = append(s.buckets, bucket)
	}

	for _, conf := range cfg.BucketSets {
		bucketSet, err := newBucketSet(conf, s, prov)
		if err != nil {
			return nil, err
		}
		s.bucketSets = append(s.bucketSets, bucketSet)
	}

	return s, nil
}

// Amp satisfies the resource.Storage interface and provides access
// to storage's parent.
func (s *storage) Amp() resource.Amp {
	return s.amp
}

// FindBucket returns the bucket with the given name.
func (s *storage) FindBucket(name string) resource.Bucket {
	for _, bkt := range s.buckets {
		if name == bkt.Name() {
			return bkt
		}
	}
	return nil
}

// Find returns the bucket with the given name.
func (s *storage) FindBucketSet(name string) resource.BucketSet {
	for _, bs := range s.bucketSets {
		if name == bs.Name() {
			return bs
		}
	}
	return nil
}

func (s *storage) ProviderStorage() resource.ProviderStorage {
	return s.providerStorage
}

// Route satisfies the embedded resource.Resource interface in resource.Storage.
// Storage does not directly terminate a request so only handles load and info
// requests from it's parent.  All other commands are routed to amp's children.
func (s *storage) Route(req *route.Request) route.Response {
	log.Route(req, "Storage")

	// Route to the appropriate resource
	switch req.Top() {
	case "":
		break
	case "bucket":
		req.Pop()
		if req.Top() == "" {
			s.Help()
			return route.FAIL
		}
		bucket := s.FindBucket(req.Top())
		if bucket == nil {
			msg.Error("Unknown bucket %q.", req.Top())
			return route.FAIL
		}
		if req.Command() == route.Audit {
			aaa.NewAudit("Bucket")
		}
		req.Flags().Append("Bucket")
		return bucket.Route(req)
	case "bucket_set":
		req.Pop()
		bucketSet := s.FindBucketSet(req.Top())
		if bucketSet == nil {
			msg.Error("Unknown bucket set %q.", req.Top())
			return route.FAIL
		}
		if req.Command() == route.Audit {
			aaa.NewAudit("Bucket Set")
		}
		req.Flags().Append("Bucket Set")
		return bucketSet.Route(req)
	default:
		Help()
		return route.FAIL
	}

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Info:
		s.Info()
		return route.OK
	case route.Config:
		s.Print()
		return route.OK
	case route.Load:
		for _, b := range s.buckets {
			if resp := b.Route(req); resp != route.OK {
				return route.FAIL
			}
		}
		for _, bs := range s.bucketSets {
			if resp := bs.Route(req); resp != route.OK {
				return route.FAIL
			}
		}
		return route.OK
	case route.Provision:
		if err := s.Provision(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Audit:
		if err := s.Audit("Bucket", "Bucket Set"); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Help:
		s.Help()
		return route.OK
	default:
		msg.Error("Internal Error: amp/storage.go Unknown command " + req.Command().String())
		s.Help()
		return route.FAIL
	}
}

func (s *storage) Provision(flags ...string) error {
	for _, b := range s.buckets {
		if err := b.Provision(flags...); err != nil {
			return err
		}
	}
	for _, bs := range s.bucketSets {
		if err := bs.Provision(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (s *storage) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	for _, v := range flags {
		err := aaa.NewAudit(v)
		if err != nil {
			return err
		}
	}
	if err := s.providerStorage.Audit("Bucket"); err != nil {
		return err
	}
	for _, b := range s.buckets {
		if err := b.Audit("Bucket"); err != nil {
			return err
		}
	}
	for _, bs := range s.bucketSets {
		if err := bs.Audit("Bucket Set"); err != nil {
		}
	}
	return nil
}

func (s *storage) Info() {
	msg.Info("Storage")
	msg.IndentInc()
	msg.Info("Buckets")
	msg.IndentInc()
	for _, b := range s.buckets {
		b.Info()
	}
	msg.IndentDec()
	msg.Info("Bucket Sets")
	msg.IndentInc()
	for _, bs := range s.bucketSets {
		bs.Info()
	}
	msg.IndentDec()
	msg.IndentDec()
}

func (s *storage) Help() {
	commands := []help.Command{
		{Name: route.Provision.String(), Desc: "update the storage"},
		{Name: route.Audit.String(), Desc: "audit the storage"},
		{Name: route.Info.String(), Desc: "show information about allocated storage"},
		{Name: route.Config.String(), Desc: "show the configuration for the given storage"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("storage", commands)
}
