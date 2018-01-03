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
	"github.com/cisco/arc/pkg/command"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/route"
)

func (i *Instance) StartPaging(req *route.Request) route.Response {
	if req.Flag("bootstrap") || req.Flag("force") {
		return route.OK
	}
	if resp := i.startPagingConsul(req); resp != route.OK {
		return resp
	}
	if resp := i.startPagingSensu(req); resp != route.OK {
		return resp
	}
	return route.OK
}

func (i *Instance) StopPaging(req *route.Request) route.Response {
	if req.Flag("bootstrap") || req.Flag("force") {
		return route.OK
	}
	if resp := i.stopPagingConsul(req); resp != route.OK {
		return resp
	}
	if resp := i.stopPagingSensu(req); resp != route.OK {
		return resp
	}
	return route.OK
}

func (i *Instance) paging(req *route.Request, cmd, service string) route.Response {
	output, err := command.RunRemoteWithOutput(command.Command{
		Instance: i,
		Desc:     cmd + " " + service + " paging",
		Src:      "/usr/lib/arc/paging/" + service + "_paging",
		Args:     []string{cmd, i.Name()},
	})
	if err != nil {
		msg.Detail("Unable to %s %s paging", cmd, service)
		msg.Warn("%s. See log for more details.", err.Error())
		log.Warn("%s", output)
	}
	return route.OK
}

func (i *Instance) startPagingConsul(req *route.Request) route.Response {
	return i.paging(req, "enable", "consul")
}

func (i *Instance) stopPagingConsul(req *route.Request) route.Response {
	return i.paging(req, "disable", "consul")
}

func (i *Instance) startPagingSensu(req *route.Request) route.Response {
	_, err := command.CopyToWithOutput(command.Command{
		Instance: i,
		Desc:     "Copy check_monit_services",
		Src:      "/usr/lib/arc/tools/check_monit_services",
	})
	if err != nil {
		msg.Detail("Unable to enable sensu paging")
		msg.Warn("%s. See log for more details.", err.Error())
	}
	return i.paging(req, "enable", "sensu")
}

func (i *Instance) stopPagingSensu(req *route.Request) route.Response {
	return i.paging(req, "disable", "sensu")
}
