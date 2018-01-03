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
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/route"
)

func (i *Instance) replace(req *route.Request) route.Response {
	msg.Info("Instance Replace: %s", i.Name())
	if i.Destroyed() {
		msg.Detail("Instance does not exist, skipping...")
		return route.OK
	}
	if resp := i.Derived().PreReplace(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().Replace(req); resp != route.OK {
		return resp
	}
	if resp := i.Derived().PostReplace(req); resp != route.OK {
		return resp
	}
	msg.Detail("Replaced: %s", i.Id())
	aaa.Accounting("Instance replaced: %s", i.Id())
	return route.OK
}

func (i *Instance) PreReplace(req *route.Request) route.Response {
	return route.OK
}

func (i *Instance) Replace(req *route.Request) route.Response {

	// Destroy
	destroyReq := req.Clone(route.Destroy)
	destroyReq.Flags().Append("preserve_volume")
	if i.eip != nil && i.eip.Id() != "" {
		destroyReq.Flags().Append("preserve_eip")
	}
	if resp := i.destroy(destroyReq); resp != route.OK {
		return resp
	}

	// Create
	createReq := req.Clone(route.Create)
	createReq.Flags().Append("preserve_volume")
	if i.eip != nil && i.eip.Id() != "" {
		createReq.Flags().Append("preserve_eip")
	}
	if resp := i.create(createReq); resp != route.OK {
		return resp
	}
	if req.Flag("noprovision") {
		return route.OK
	}

	// Initial Provision
	provisionReq := req.Clone(route.Provision)
	provisionReq.Flags().Append("initial")
	if resp := i.provision(provisionReq); resp != route.OK {
		return resp
	}

	return route.OK
}

func (i *Instance) PostReplace(req *route.Request) route.Response {
	return route.OK
}
