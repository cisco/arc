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
	"github.com/cisco/arc/pkg/route"
)

func (i *Instance) start(req *route.Request) route.Response {
	msg.Info("Instance Start: %s", i.Name())
	if i.Destroyed() {
		msg.Detail("Instance does not exist. Skipping...")
		return route.OK
	}
	if !i.Stopped() {
		msg.Detail("Instance has been started. Skipping...")
		return route.OK
	}

	if resp := i.Derived().PreStart(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().Start(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().PostStart(req); resp != route.OK {
		return resp
	}
	msg.Detail("Started: %s", i.Id())
	aaa.Accounting("Instance started: %s, %s", i.Name(), i.Id())
	return route.OK
}

func (i *Instance) PreStart(req *route.Request) route.Response {
	return route.OK
}

func (i *Instance) Start(req *route.Request) route.Response {
	return i.providerInstance.Route(req)
}

func (i *Instance) PostStart(req *route.Request) route.Response {
	if resp := i.setupArc(req, "fix arc permissions", false); resp != route.OK {
		return resp
	}
	if resp := i.StartPaging(req); resp != route.OK {
		return resp
	}
	// Pickup new dns a record.
	if resp := i.load(req.Clone(route.Load)); resp != route.OK {
		return resp
	}
	return route.OK
}

// Stop

func (i *Instance) stop(req *route.Request) route.Response {
	msg.Info("Instance Stop: %s", i.Name())
	if i.Destroyed() {
		msg.Detail("Instance does not exist, skipping...")
		return route.OK
	}
	if !i.Started() {
		msg.Detail("Instance has been stopped. Skipping...")
		return route.OK
	}

	if resp := i.Derived().PreStop(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().Stop(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().PostStop(req); resp != route.OK {
		return resp
	}
	msg.Detail("Stopped: %s", i.Id())
	aaa.Accounting("Instance stopped: %s, %s", i.Name(), i.Id())
	return route.OK
}

func (i *Instance) PreStop(req *route.Request) route.Response {
	if resp := i.setupArc(req, "fix arc permissions", false); resp != route.OK {
		return resp
	}
	if resp := i.StopPaging(req); resp != route.OK {
		return resp
	}
	return route.OK
}

func (i *Instance) Stop(req *route.Request) route.Response {
	// Hard stop invokes the provider api to stop the instance.
	if req.Flag("hard") {
		return i.providerInstance.Route(req)
	}

	// Otherwise shutdown the instance gracefully.
	commands := []command.Command{
		{
			Type: command.Sudo,
			Desc: "stop instance",
			Src:  "/sbin/shutdown -h now",
		},
	}
	command.RunQuiet(commands, i)

	if !i.reloadStopped(req.Clone(route.Load)) {
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) PostStop(req *route.Request) route.Response {
	return route.OK
}

// Restart

/* FIXME: The commented out code below attempts to reboot the system using the
   provider api. However the AWS api for reboot is, let's say,
   difficult to use correctly. First off it's an asynchronous call
   that returns immediately. Second there isn't a "rebooting" state,
   so we can't query the state of the machine to know if/when it has
   rebooted. Finally, it does a best effort at cleanly shutting down
   the instance, but will revert to a hard restart after 4 minutes.
   I haven't yet managed to get it to shutdown cleanly, I suspect due
   to the hardening puppet module removing /etc/init/control-alt-delete.conf.

   I'll leave this here for now. Since using the provider API provides quite
   a few benefits...

   When you reboot an instance, it remains on the same physical host, so
   your instance keeps its public DNS name (IPv4), private IPv4 address,
   IPv6 address (if applicable), and any data on its instance store volumes.

   Rebooting an instance doesn't start a new instance billing hour, unlike
   stopping and restarting your instance.
*/

func (i *Instance) restart(req *route.Request) route.Response {
	msg.Info("Instance Restart: %s", i.Name())
	if i.Destroyed() {
		msg.Detail("Instance does not exist, skipping...")
		return route.OK
	}
	if resp := i.Derived().PreRestart(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().Restart(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().PostRestart(req); resp != route.OK {
		return resp
	}
	msg.Detail("Restarted: %s", i.Id())
	aaa.Accounting("Instance restarted: %s, %s", i.Name(), i.Id())
	return route.OK
}

func (i *Instance) PreRestart(req *route.Request) route.Response {
	//return i.PreStop(req)
	return route.OK
}

func (i *Instance) Restart(req *route.Request) route.Response {
	//return i.providerInstance.Route(req)
	if resp := i.stop(req.Clone(route.Stop)); resp != route.OK {
		return resp
	}
	if resp := i.start(req.Clone(route.Start)); resp != route.OK {
		return resp
	}
	return route.OK
}

func (i *Instance) PostRestart(req *route.Request) route.Response {
	//return i.PostStart(req)
	return route.OK
}
