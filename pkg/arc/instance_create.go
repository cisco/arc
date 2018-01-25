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
	"strconv"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/command"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

func (i *Instance) create(req *route.Request) route.Response {
	msg.Info("Instance Create: %s", i.Name())
	if resp := i.Derived().PreCreate(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().Create(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().PostCreate(req); resp != route.OK {
		return resp
	}
	msg.Detail("Created: %s", i.Id())
	aaa.Accounting("Instance created: %s, %s", i.Name(), i.Id())
	return route.OK
}

func (i *Instance) PreCreate(req *route.Request) route.Response {
	return route.OK
}

func (i *Instance) Create(req *route.Request) route.Response {
	if !i.subnet.Created() {
		msg.Error("Instance %q cannot be associated with subnet %q. The subnet group needs to be created.", i.Name(), i.subnet.Name())
		return route.FAIL
	}
	if resp := i.providerInstance.Route(req); resp != route.OK {
		return resp
	}
	if err := i.role.Attach(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if resp := i.createElasticIP(req); resp != route.OK {
		return resp
	}
	if resp := i.attachVolumes(req); resp != route.OK {
		return resp
	}
	return route.OK
}

func (i *Instance) createElasticIP(req *route.Request) route.Response {
	if i.eip == nil {
		return route.OK
	}
	if req.Flag("preserve_eip") {
		if err := i.eip.Attach(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	}
	if err := i.eip.Create(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if err := i.eip.Attach(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) attachVolumes(req *route.Request) route.Response {
	if !req.Flag("preserve_volume") {
		return route.OK
	}
	for _, r := range i.volumes.Get() {
		v := r.(resource.Volume)
		if !v.Preserve() {
			continue
		}
		if v.Attached() {
			continue
		}
		err := v.Attach()
		if err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	}
	return route.OK
}

func (i *Instance) PostCreate(req *route.Request) route.Response {
	enabled := "enabled"
	if req.Flag("bootstrap") {
		enabled = "disabled"
	}

	if err := i.createTags(req); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if err := i.createSecurityTags(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if i.createDnsARecords(req) != route.OK {
		return route.FAIL
	}
	if resp := i.setupArc(req, "setup arc", true); resp != route.OK {
		return resp
	}
	if resp := i.configureUsers(req, true); resp != route.OK {
		return resp
	}
	if resp := i.setupArc(req, "fix arc permissions", false); resp != route.OK {
		return resp
	}

	dcName := i.Pod().Cluster().Compute().Name()
	domain := dcName + ".local"
	consulDomain := dcName + ".consul"
	if i.Dns() != nil {
		domain = i.Dns().Domain()
		consulDomain = i.Dns().Subdomain() + ".consul"
	}

	commands := []command.Command{
		{
			Type: command.Remote,
			Desc: "setup hostname",
			Src:  "/usr/lib/arc/create/setup_hostname",
			Args: []string{i.PrivateFQDN()},
		},
		{
			Type: command.Remote,
			Desc: "setup dhcp",
			Src:  "/usr/lib/arc/create/setup_dhcp",
			Args: []string{consulDomain, domain},
		},
		{
			Type: command.Remote,
			Desc: "setup repos",
			Src:  "/usr/lib/arc/create/setup_repos",
			Args: []string{enabled},
		},
	}
	for _, r := range i.volumes.Get() {
		v := r.(resource.Volume)
		if v.Boot() {
			continue
		}
		var args []string
		if req.Flag("preserve_volume") && v.Preserve() {
			args = []string{v.Device(), v.MountPoint(), v.FsType(), strconv.Itoa(v.Inodes()), "skip_format"}
		} else {
			args = []string{v.Device(), v.MountPoint(), v.FsType(), strconv.Itoa(v.Inodes())}
		}
		commands = append(commands, command.Command{
			Type: command.Remote,
			Desc: "setup volume for " + v.Device(),
			Src:  "/usr/lib/arc/create/setup_volume",
			Args: args,
		})
	}
	if !command.Run(commands, i) {
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) setupArc(req *route.Request, desc string, asRoot bool) route.Response {
	commands := []command.Command{
		{
			Type: command.Remote,
			Desc: desc,
			Src:  "/usr/lib/arc/create/setup_arc",
			Dest: "/tmp/setup_arc",
		},
		{
			Type: command.Copy,
			Desc: "install arc.sh library",
			Src:  "/usr/lib/arc/arc.sh",
		},
	}
	f := command.Run
	if asRoot {
		f = command.RunAsRoot
	}
	if !f(commands, i) {
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) createTags(req *route.Request) error {
	tags := map[string]string{
		"Name":             i.Name(),
		"Created By":       req.UserId(),
		"Last Modified By": req.UserId(),
		"Last Modified":    req.Time(),
		"DataCenter":       req.DataCenter(),
	}

	msg.Info("Set Tags: %s", i.Name())
	err := i.SetTags(tags)
	if err != nil {
		return err
	}

	msg.Info("Set Volume Tags: %s", i.Name())
	for _, r := range i.volumes.Get() {
		v := r.(resource.Volume)
		mnt := ""
		if v.MountPoint() != "" {
			mnt = ", " + v.MountPoint()
		}
		msg.Detail("Set volume: %s%s", v.Device(), mnt)
		if v.MountPoint() == "" {
			tags["Name"] = "/"
		} else {
			tags["Name"] = v.MountPoint()
		}
		err = v.SetTags(tags)
		if err != nil {
			msg.Warn("Failed to set tags for %s\n\t%s", v.Device(), err.Error())
		}
	}
	return nil
}

func (i *Instance) updateTags(req *route.Request) error {
	tags := map[string]string{
		"Last Modified By": req.UserId(),
		"Last Modified":    req.Time(),
	}

	err := i.SetTags(tags)
	if err != nil {
		return err
	}

	msg.Info("Update Volume Tags: %s", i.Name())
	for _, r := range i.volumes.Get() {
		v := r.(resource.Volume)
		mnt := ""
		if v.MountPoint() != "" {
			mnt = ", " + v.MountPoint()
		}
		msg.Detail("Update volume: %s%s", v.Device(), mnt)
		err = v.SetTags(tags)
		if err != nil {
			msg.Warn("Failed to update tags for %s\n\t%s", v.Device(), err.Error())
		}
	}
	return nil
}

func (i *Instance) createSecurityTags() error {
	tags := map[string]string{}
	for k, v := range i.Pod().Cluster().Compute().DataCenter().SecurityTags() {
		tags[k] = v
	}
	for k, v := range i.Pod().Cluster().SecurityTags() {
		tags[k] = v
	}
	return i.SetTags(tags)
}
