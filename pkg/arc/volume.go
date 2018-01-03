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
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type volume struct {
	*config.Volume
	providerVolume resource.ProviderVolume
}

// newVolume is the constructor for a volume object. It returns a non-nil error upon failure.
func newVolume(comp resource.Compute, prov provider.DataCenter, cfg *config.Volume) (*volume, error) {
	log.Debug("Initializing Volume '%s'", cfg.Device())

	v := &volume{
		Volume: cfg,
	}

	var err error
	v.providerVolume, err = prov.NewVolume(comp, cfg)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// Route satisfies the embedded resource.Resource interface in resource.Volume.
// Volume handles load, create, destroy, and info requests by delegating them
// to the providerVolume.
func (v *volume) Route(req *route.Request) route.Response {
	log.Route(req, "Volume %q", v.Device())
	return route.FAIL
}

func (v *volume) Audit(flags ...string) error {
	return v.providerVolume.Audit(flags...)
}

func (v *volume) Load() error {
	return v.providerVolume.Load()
}

// Created satisfies the embedded resource.Resource interface in resource.Volume.
// It delegates the call to the provider's volume.
func (v *volume) Created() bool {
	return v.providerVolume.Created()
}

// Destroyed satisfies the embedded resource.Resource interface in resource.Volume.
// It delegates the call to the provider's volume.
func (v *volume) Destroyed() bool {
	return v.providerVolume.Destroyed()
}

// Id provides the id of the provider specific volume resource.
// This satisfies the resource.DynamicVolume interface.
func (v *volume) Id() string {
	return v.providerVolume.Id()
}

func (v *volume) State() string {
	return v.providerVolume.State()
}

func (v *volume) Attached() bool {
	return v.providerVolume.Attached()
}

func (v *volume) Detached() bool {
	return v.providerVolume.Detached()
}

func (v *volume) Detach() error {
	msg.Detail("Volume Detach: %s, %s", v.Device(), v.MountPoint())
	return v.providerVolume.Detach()
}

func (v *volume) Attach() error {
	msg.Detail("Volume Attach: %s, %s", v.Device(), v.MountPoint())
	return v.providerVolume.Attach()
}

func (v *volume) Destroy() error {
	msg.Detail("Volume Destroy: %s, %s", v.Device(), v.MountPoint())
	return v.providerVolume.Destroy()
}

func (v *volume) Reset() {
	v.providerVolume.Reset()
}

func (v *volume) SetTags(t map[string]string) error {
	return v.providerVolume.SetTags(t)
}

func (v *volume) ProviderVolume() resource.ProviderVolume {
	return v.providerVolume
}

func (v *volume) Info() {
	if v.Volume.Boot() {
		msg.Info("Volume: /")
	} else {
		msg.Info("Volume: %s", v.Volume.MountPoint())
	}
	msg.IndentInc()
	v.Volume.Print()
	v.providerVolume.Info()
	msg.IndentDec()
}
