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

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type volumes struct {
	*resource.Resources
	*config.Volumes
	volumes map[string]resource.Volume
}

// newVolumes is the constructor for a volumes object. It returns a non-nil error upon failure.
func newVolumes(comp resource.Compute, prov provider.DataCenter, cfg *config.Volumes) (*volumes, error) {
	log.Debug("Initializing Volumes")

	v := &volumes{
		Resources: resource.NewResources(),
		Volumes:   cfg,
		volumes:   map[string]resource.Volume{},
	}

	for _, conf := range *cfg {
		if v.Find(conf.Device()) != nil {
			return nil, fmt.Errorf("Volume device %q must be unique but is used multiple times", conf.Device())
		}
		volume, err := newVolume(comp, prov, conf)
		if err != nil {
			return nil, err
		}
		v.volumes[volume.Device()] = volume
		v.Append(volume)
	}
	return v, nil
}

// Find satisfies the resource.Volumes interface and provides a way
// to search for a specific volume. This assumes volume names are unique.
func (v *volumes) Find(device string) resource.Volume {
	return v.volumes[device]
}

// Route satisfies the embedded resource.Resource interface in resource.Volumes.
func (v *volumes) Route(req *route.Request) route.Response {
	log.Route(req, "Volumes")

	return route.FAIL
}

func (v *volumes) Audit(flags ...string) error {
	for _, a := range v.volumes {
		if err := a.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (v *volumes) info() {
	msg.IndentInc()
	msg.Info("Volumes")
	msg.IndentInc()
	for _, v := range v.volumes {
		v.Info()
	}
	msg.IndentDec()
	msg.IndentDec()
}
