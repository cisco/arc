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
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/command"
	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/hiera"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/route"
	"github.com/cisco/arc/pkg/secrets"
)

func (i *Instance) provision(req *route.Request) route.Response {
	msg.Info("Instance Provision: %s", i.Name())
	if i.Destroyed() {
		msg.Detail("Instance does not exist, skipping...")
		return route.OK
	}

	if resp := i.provisionSingleResource(req); resp != route.CONTINUE {
		return resp
	}

	if resp := i.Derived().PreProvision(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().Provision(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().PostProvision(req); resp != route.OK {
		return resp
	}
	msg.Detail("Provisioned: %s", i.Id())
	aaa.Accounting("Instance provisioned: %s, %s", i.Name(), i.Id())
	return route.OK
}

func (i *Instance) provisionSingleResource(req *route.Request) route.Response {
	r := route.CONTINUE
	switch {
	case req.Flag("aide"):
		if resp := i.setupArc(req, "setting up arc", false); resp != route.OK {
			return resp
		}
		return i.provisionAide(req)
	case req.Flag("role"):
		if err := i.role.Update(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case req.Flag("users"):
		if resp := i.UserUpdate(req); resp != route.OK {
			return resp
		}
		if resp := i.provisionAide(req); resp != route.OK {
			return resp
		}
		return route.OK
	case req.Flag("tags"):
		msg.Detail("Updating tags")
		if err := i.updateTags(req); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		if err := i.createSecurityTags(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	}
	return r
}

func (i *Instance) UserUpdate(req *route.Request) route.Response {
	if resp := i.setupArc(req, "setup arc on pre-refactor instances", false); resp != route.OK {
		return resp
	}
	if resp := i.configureUsers(req, false); resp != route.OK {
		return resp
	}
	msg.Detail("Updating tags")
	if err := i.updateTags(req); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if err := i.createSecurityTags(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) PreProvision(req *route.Request) route.Response {
	if !req.Flag("initial") {
		if resp := i.setupArc(req, "setup arc on pre-refactor instances", false); resp != route.OK {
			return resp
		}
		if resp := i.configureUsers(req, false); resp != route.OK {
			return resp
		}
		msg.Detail("Updating tags")
		if err := i.updateTags(req); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		if err := i.createSecurityTags(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		if resp := i.StopPaging(req); resp != route.OK {
			return resp
		}
	}
	return route.OK
}

func (i *Instance) Provision(req *route.Request) route.Response {
	if err := i.role.Update(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if err := hiera.Install(i, req.Flag("bootstrap")); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if resp := i.provisionInstallServertype(req); resp != route.OK {
		return resp
	}
	if resp := i.provisionInstallPackages(req); resp != route.OK {
		return resp
	}
	if resp := i.provisionInstallPuppet(req); resp != route.OK {
		return resp
	}
	if !req.Flag("bootstrap") {
		if err := secrets.InstallCerts(i); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		if err := secrets.InstallMachineUser(i); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	}
	if resp := i.provisionApplyServertype(req); resp != route.OK {
		return resp
	}
	return route.OK
}

func (i *Instance) PostProvision(req *route.Request) route.Response {
	// If the bootstrap flag is set, update aide... that's it.
	if req.Flag("bootstrap") {
		return i.provisionAide(req)
	}

	// If the initial flag is set, meaning the initial provisioning run,
	// we want to update the software and restart...
	if req.Flag("initial") {
		if !command.RunRemote(command.Command{
			Instance: i,
			Desc:     "update software",
			Src:      "/usr/lib/arc/provision/update_software",
		}) {
			return route.FAIL
		}
		if resp := i.Route(req.Clone(route.Restart)); resp != route.OK {
			return resp
		}
	} else {
		// ... otherwise we need to start paging.
		if resp := i.StartPaging(req); resp != route.OK {
			return resp
		}
	}
	// ... and update aide.
	if resp := i.provisionAide(req); resp != route.OK {
		return resp
	}
	return route.OK
}

func (i *Instance) provisionInstallPuppet(req *route.Request) route.Response {
	if req.Flag("nopuppet") {
		return route.OK
	}
	if !command.RunRemote(command.Command{
		Instance: i,
		Desc:     "setup puppet",
		Src:      "/usr/lib/arc/provision/setup_puppet",
		Args:     []string{"fresh_install"}, // FIXME
	}) {
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) provisionInstallServertype(req *route.Request) route.Response {
	commands := []command.Command{
		{
			Type: command.Local,
			Desc: "pull servertype from mirror",
			Src:  "/usr/lib/arc/provision/pull_pkg",
			Args: []string{i.Pod().PkgName(), env.Lookup("ARC"), i.ServerType()},
		},
		{
			Type: command.Copy,
			Desc: "push servertype",
			Src:  env.Lookup("ARC") + "/" + i.Pod().PkgName(),
			Dest: "/usr/lib/arc/" + i.Pod().PkgName(),
		},
		{
			Type: command.Remote,
			Desc: "install servertype",
			Src:  "/usr/lib/arc/tools/install_pkg",
			Args: []string{"/usr/lib/arc/" + i.Pod().PkgName()},
		},
	}
	if !command.Run(commands, i) {
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) provisionInstallPackages(req *route.Request) route.Response {
	msg.Quiet(true)
	defer msg.Quiet(false)

	// The pull_packages script copies the packages.txt file and the
	// packages from the master mirror to the ARC directory.
	// If the packages.txt file doesn't exist the script will still succeed.
	if !command.RunLocal(command.Command{
		Desc: "pull packages from mirror",
		Src:  "/usr/lib/arc/provision/pull_packages",
		Args: []string{i.ServerType(), i.Version(), env.Lookup("ARC")},
	}) {
		return route.FAIL
	}

	// Read the contents of the packages.txt file. Each line of the file contains a name of a single package.
	packagesFile := env.Lookup("ARC") + fmt.Sprintf("/packages-%s.txt", i.Version())
	log.Debug("Packages file: %s", packagesFile)

	f, err := os.Open(packagesFile)
	if err != nil {
		log.Debug(err.Error())
		msg.Detail("No packages found. Skipping...")
		return route.OK
	}
	defer f.Close()

	commands := []command.Command{
		{
			Type: command.Message,
			Dest: "Info",
			Desc: "Installing packages",
		},
		{
			Type: command.Copy,
			Desc: "copy manifest file to the instance",
			Src:  packagesFile,
			Dest: "/usr/lib/arc/" + fmt.Sprintf("packages-%s.txt", i.Version()),
		},
	}
	s := bufio.NewScanner(f)
	s.Split(bufio.ScanLines)
	for s.Scan() {
		pkg := filepath.Base(s.Text())
		commands = append(commands,
			command.Command{
				Type: command.Copy,
				Desc: fmt.Sprintf("upload %q", pkg),
				Src:  env.Lookup("ARC") + "/" + pkg,
				Dest: "/usr/lib/arc/" + pkg,
			},
			command.Command{
				Type: command.Local,
				Desc: fmt.Sprintf("cleanup %q", pkg),
				Src:  "rm",
				Args: []string{"-f", env.Lookup("ARC") + "/" + pkg},
			},
			command.Command{
				Type: command.Message,
				Desc: "Installing: " + pkg,
			},
		)
	}
	commands = append(commands, command.Command{
		Type: command.Remote,
		Desc: "install all the packages",
		Src:  "/usr/lib/arc/tools/install_packages",
		Args: []string{fmt.Sprintf("/usr/lib/arc/packages-%s.txt", i.Version())},
	},
		command.Command{
			Type: command.Message,
			Dest: "Detail",
			Desc: "Packages Installed",
		},
		command.Command{
			Type: command.Local,
			Desc: "cleanup packages from mirror",
			Src:  "rm",
			Args: []string{"-f", packagesFile},
		},
	)
	if !command.RunQuiet(commands, i) {
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) provisionApplyServertype(req *route.Request) route.Response {
	if !command.RunRemote(command.Command{
		Instance: i,
		Desc:     "apply servertype",
		Src:      "/usr/lib/arc/provision/apply_module",
		Args:     []string{"st_" + i.ServerType()},
	}) {
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) provisionAide(req *route.Request) route.Response {
	pkgName := ""
	switch {
	case strings.HasPrefix(i.Pod().Image(), "centos"):
		pkgName = "puppet-aide-1.0.0.0-" + strconv.Itoa(i.Pod().Cluster().Compute().AideVersion()) + ".noarch.rpm"
	case strings.HasPrefix(i.Pod().Image(), "ubuntu"):
		pkgName = "puppet-aide_1.0.0.0-" + strconv.Itoa(i.Pod().Cluster().Compute().AideVersion()) + "_all.deb"
	}
	log.Debug("Aide Package Name: %q", pkgName)

	commands := []command.Command{
		{
			Type: command.Local,
			Desc: "pull aide puppet module from mirror",
			Src:  "/usr/lib/arc/provision/pull_pkg",
			Args: []string{pkgName, env.Lookup("ARC")},
		},
		{
			Type: command.Copy,
			Desc: "push aide puppet module",
			Src:  env.Lookup("ARC") + "/" + pkgName,
			Dest: "/usr/lib/arc/" + pkgName,
		},
		{
			Type: command.Remote,
			Desc: "install aide puppet module",
			Src:  "/usr/lib/arc/tools/install_pkg",
			Args: []string{"/usr/lib/arc/" + pkgName},
		},
		{
			Type: command.Remote,
			Desc: "apply aide",
			Src:  "/usr/lib/arc/provision/apply_module",
			Args: []string{"aide"},
		},
	}
	if !command.Run(commands, i) {
		return route.FAIL
	}
	return route.OK
}
