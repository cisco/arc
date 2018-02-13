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

package mock

import (
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

type containerService struct {
	*config.ContainerService
	opt options
}

func newContainerService(cfg *config.ContainerService, p *containerServiceProvider) (resource.ProviderContainerService, error) {
	log.Info("Initializing Mock Container Service")
	cs := &containerService{
		ContainerService: cfg,
		opt:              options{p.Provider.Data},
	}
	if cs.opt.err("cs.New") {
		return nil, dberr{"cs.New"}
	}
	return cs, nil
}

func (cs *containerService) Load() error {
	log.Info("Loading Mock Container Service")
	if cs.opt.err("cs.Load") {
		return dberr{"cs.Load"}
	}
	return nil
}

func (cs *containerService) Create(flags ...string) error {
	msg.Info("Creating Mock ContainerService")
	if cs.opt.err("cs.Create") {
		return dberr{"cs.Create"}
	}
	return nil
}

func (cs *containerService) Created() bool {
	if cs.opt.err("cs.Created") {
		return false
	}
	return true
}

func (cs *containerService) Destroy(flags ...string) error {
	msg.Info("Destroying Mock ContainerService")
	if cs.opt.err("cs.Destroy") {
		return dberr{"cs.Destroy"}
	}
	return nil
}

func (cs *containerService) Destroyed() bool {
	if cs.opt.err("cs.Destroyed") {
		return false
	}
	return true
}

func (cs *containerService) Provision(flags ...string) error {
	msg.Info("Provisioning Mock ContainerService")
	if cs.opt.err("cs.Provision") {
		return dberr{"cs.Provision"}
	}
	return nil
}

func (cs *containerService) Audit(flags ...string) error {
	msg.Info("Auditing Mock ContainerService")
	if cs.opt.err("cs.Audit") {
		return dberr{"cs.Audit"}
	}
	return nil
}

func (cs *containerService) Info() {
	msg.Info("Mock ContainerService")
	msg.Detail("...")
}
