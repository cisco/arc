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
	buckets         []resource.Bucket
	bucketSets      []resource.BucketSet
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

	return s, nil
}

// Amp satisfies the resource.Storage interface and provides access
// to storage's parent.
func (s *storage) Amp() resource.Amp {
	return s.amp
}

// Find returns the bucket with the given name.
func (s *storage) Find(name string) resource.Bucket {
	for _, bkt := range s.buckets {
		if name == bkt.Name() {
			return bkt
		}
	}
	return nil
}

// Buckets satisfies the resource.Storage interface and provides access
// to storage's children.
func (s *storage) Buckets() []resource.Bucket {
	return s.buckets
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
		bucket := s.Find(req.Top())
		if bucket == nil {
			msg.Error("Unknown bucket %q.", req.Top())
			return route.FAIL
		}
		return bucket.Route(req)
	case "bucket_set":
		/*
			req.Pop()
			if req.Top() == "" {
				return s.bucketSets.Route(req)
			}
			bucketSet := s.bucketSets.Find(req.Top())
			if bucketSet == nil {
				msg.Error("Unknown bucket set %q.", req.Top())
				return route.FAIL
			}
			return bucketSet.Route(req)
		*/
	}

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Info:
		s.info(req)
		return route.OK
	case route.Config:
		s.config(req)
		return route.OK
	case route.Load:
		return s.RouteInOrder(req)
	case route.Provision:
		return s.RouteInOrder(req)
	case route.Audit:
		if err := s.Audit("Bucket"); err != nil {
			return route.FAIL
		}
		return route.OK
	}
	msg.Error("Internal Error: amp/storage.go. Unknown command %s", req.Command())
	return route.FAIL
}

func (s *storage) info(req *route.Request) {
	msg.Info("Storage")
	msg.IndentInc()
	for _, b := range s.buckets {
		b.Route(req)
	}
	for _, bs := range s.bucketSets {
		bs.Route(req)
	}
	msg.IndentDec()
}

func (s *storage) config(req *route.Request) {
	msg.Info("Storage")
	msg.IndentInc()
	s.RouteInOrder(req)
	msg.IndentDec()
}

func (s *storage) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	err := aaa.NewAudit(flags[0])
	if err != nil {
		return err
	}
	if err := s.providerStorage.Audit(flags...); err != nil {
		return err
	}
	for _, b := range s.buckets {
		if err := b.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (s *storage) Help() {
}
