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

type bucketSets struct {
	*resource.Resources
	bucketSets map[string]resource.BucketSet
	storage    *storage
}

func newBucketSets(s *storage, prov provider.Account, cfg *config.BucketSets) (*bucketSets, error) {
	log.Debug("Initializing Bucket Sets")

	b := &bucketSets{
		Resources:  resource.NewResources(),
		bucketSets: map[string]resource.BucketSet{},
		storage:    s,
	}

	for _, conf := range *cfg {
		if b.Find(conf.Name()) != nil {
			return nil, fmt.Errorf("Bucket Replication Set name %q must be unique, but it is used multiple times", conf.Name())
		}
		bucketSet, err := newBucketSet(b, prov, conf)
		if err != nil {
			return nil, err
		}
		b.bucketSets[conf.Name()] = bucketSet
		b.Append(bucketSet)
	}
	return b, nil
}

func (b *bucketSets) Storage() resource.Storage {
	return b.storage
}

func (b *bucketSets) Find(name string) resource.BucketSet {
	return b.bucketSets[name]
}

func (b *bucketSets) Route(req *route.Request) route.Response {
	log.Route(req, "Bucket Replication Sets")
	switch req.Command() {
	case route.Info:
		b.Info()
		return route.OK
	}
	return b.RouteInOrder(req)
}

func (b *bucketSets) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	for _, v := range b.bucketSets {
		if err := v.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (b *bucketSets) Info() {
	if b.Destroyed() {
		return
	}
	msg.Info("Bucket Sets")
	msg.IndentInc()
	for _, bs := range b.bucketSets {
		bs.Info()
	}
	msg.IndentDec()
}
