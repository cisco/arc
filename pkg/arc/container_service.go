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

package arc

import (
	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type containerService struct {
	*config.ContainerService
	arc                      resource.Arc
	providerContainerService resource.ProviderContainerService
}

// newContainerService is the constructor for a container service object. It returns a non-nil error upon failure.
func newContainerService(cfg *config.ContainerService, arc *arc) (*containerService, error) {
	if cfg == nil {
		return nil, nil
	}
	log.Debug("Initializing Container Service")

	// Use the arc provider, if it exists, when the datacenter provider isn't available.
	if cfg.Provider == nil && arc.Arc.Provider != nil {
		cfg.Provider = arc.Arc.Provider
	}

	cs := &containerService{
		ContainerService: cfg,
		arc:              arc,
	}

	p, err := provider.NewContainerService(cfg)
	if err != nil {
		return nil, err
	}

	cs.providerContainerService, err = p.NewContainerService(cfg)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

// Route satisfies the resource.ContainerService interface.
func (cs *containerService) Route(req *route.Request) route.Response {
	log.Route(req, "ContainerService")

	if req.Top() != "" {
		cs.Help()
		return route.FAIL
	}

	// Skip if the test flag is set
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Commands that can be handled locally
	switch req.Command() {
	case route.Load:
		if err := cs.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Create:
		if err := cs.Create(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Destroy:
		if err := cs.Provision(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Provision:
		if err := cs.Provision(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Audit:
		if err := cs.Audit(req.Flags().Get()...); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case route.Info:
		cs.Info()
	case route.Config:
		cs.Print()
	case route.Help:
		cs.Help()
	default:
		cs.Help()
		return route.FAIL
	}
	return route.OK
}

// Load satisfies the resource.ContainerService interface.
func (cs *containerService) Load() error {
	log.Info("Loading Container Service: %q", cs.Name())
	return cs.providerContainerService.Load()
}

// Create satisfies the resource.ContainerService interface.
func (cs *containerService) Create(flags ...string) error {
	msg.Info("Container Service Creation: %s", cs.Name())
	if cs.Created() {
		msg.Detail("Container service exists, skipping...")
		return nil
	}
	if err := cs.providerContainerService.Create(flags...); err != nil {
		return err
	}
	msg.Detail("Created %s", cs.Name())
	return nil
}

// Created is required since the parent of this object, Arc, wants to treat it like a resource.Resource.
func (cs *containerService) Created() bool {
	return cs.providerContainerService.Created()
}

// Destroy satisfies the resource.ContainerService interface.
func (cs *containerService) Destroy(flags ...string) error {
	msg.Info("Container Service Destruction: %s", cs.Name())
	if cs.Destroyed() {
		msg.Detail("Container service does not exist, skipping...")
		return nil
	}
	if err := cs.providerContainerService.Destroy(flags...); err != nil {
		return err
	}
	msg.Detail("Destroyed %s", cs.Name())
	return nil
}

// Destroyed is required since the parent of this object, Arc, wants to treat it like a resource.Resource.
func (cs *containerService) Destroyed() bool {
	return cs.providerContainerService.Destroyed()
}

// Provision satisfies the resource.ContainerService interface.
func (cs *containerService) Provision(flags ...string) error {
	msg.Info("Container Service Provision: %s", cs.Name())
	if cs.Destroyed() {
		msg.Detail("Container service does not exist, skipping...")
		return nil
	}
	if err := cs.providerContainerService.Provision(flags...); err != nil {
		return err
	}
	msg.Detail("Provisioned %s", cs.Name())
	return nil
}

// Audit satisfies the resource.ContainerService interface.
func (cs *containerService) Audit(flags ...string) error {
	auditSession := "Container"
	flags = append(flags, auditSession)

	err := aaa.NewAudit(auditSession)
	if err != nil {
		return err
	}
	if err := cs.providerContainerService.Audit(flags...); err != nil {
		return err
	}
	return nil
}

// Info satisfies the resource.ContainerService interface.
func (cs *containerService) Info() {
	if cs.Destroyed() {
		return
	}
	msg.Info("Container Service")
	cs.providerContainerService.Info()
}

// Help satisfies resource.ContainerService.
func (cs *containerService) Help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: "create the container service"},
		{Name: route.Destroy.String(), Desc: "destroy the container service"},
		{Name: route.Provision.String(), Desc: "update the container service"},
		{Name: route.Audit.String(), Desc: "audit the container service"},
		{Name: route.Info.String(), Desc: "show information about allocated container service"},
		{Name: route.Config.String(), Desc: "show the configuration for the given container service"},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print("container", commands)
}

// Arc satisfies the resource.ContainerService interface and provides access to container's parent.
func (cs *containerService) Arc() resource.Arc {
	return cs.arc
}

// ProviderContainerSerivces allows access to the provider's container service object.
func (cs *containerService) ProviderContainerService() resource.ProviderContainerService {
	return cs.providerContainerService
}
