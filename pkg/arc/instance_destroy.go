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
	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/command"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

func (i *Instance) destroy(req *route.Request) route.Response {
	msg.Info("Instance Destruction: %s", i.Name())
	if i.Destroyed() {
		msg.Detail("Instance does not exist, skipping...")
		return route.OK
	}
	id := i.Id()
	if resp := i.Derived().PreDestroy(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().Destroy(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().PostDestroy(req); resp != route.OK {
		return resp
	}
	msg.Detail("Destroyed: %s", id)
	aaa.Accounting("Instance destroyed: %s, %s", i.Name(), id)
	return route.OK
}

func (i *Instance) PreDestroy(req *route.Request) route.Response {
	if resp := i.setupArc(req, "fix arc permissions", false); resp != route.OK {
		return resp
	}
	if resp := i.StopPaging(req); resp != route.OK {
		return resp
	}
	if i.destroyDnsARecords(req) != route.OK {
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) Destroy(req *route.Request) route.Response {
	destroy := true
	if req.Flag("preserve_volume") {
		destroy = false
	}
	if resp := i.detachVolumes(req, destroy); resp != route.OK {
		return resp
	}
	if resp := i.destroyElasticIP(req); resp != route.OK {
		return resp
	}
	return i.providerInstance.Route(req)
}

func (i *Instance) destroyElasticIP(req *route.Request) route.Response {
	if i.eip == nil || i.eip.Id() == "" {
		return route.OK
	}
	if err := i.eip.Detach(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if req.Flag("preserve_eip") {
		return route.OK
	}
	if err := i.eip.Destroy(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) detachVolumes(req *route.Request, destroy bool) route.Response {
	for _, r := range i.volumes.Get() {
		v := r.(resource.Volume)
		if !v.Preserve() {
			v.Reset()
			continue
		}
		if v.Detached() {
			continue
		}
		if !command.RunRemote(command.Command{
			Instance: i,
			Desc:     "unmount " + v.MountPoint(),
			Src:      "/usr/lib/arc/destroy/umount",
			Args:     []string{v.MountPoint()},
		}) {
			return route.FAIL
		}
		msg.Info("Volume Destruction")
		if err := v.Detach(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		if destroy {
			if err := v.Destroy(); err != nil {
				msg.Error(err.Error())
				return route.FAIL
			}
		}
	}
	return route.OK
}

func (i *Instance) PostDestroy(req *route.Request) route.Response {
	return route.OK
}
