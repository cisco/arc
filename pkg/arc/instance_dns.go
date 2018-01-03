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
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/route"
)

func (i *Instance) newDnsARecords() error {
	var err error
	// Allocate the dns a record for the private ip address.
	if i.privateARecord == nil {
		i.privateARecord, err = newDynamicDnsARecord(i, i.dns, i.PrivateHostname(), "private", i.Pod().Cluster().AuditIgnore())
		if err != nil {
			return err
		}
	}
	// Allocation the dns a record for the public ip address.
	if (i.subnet.Access() == "public" || i.subnet.Access() == "public_elastic") && i.publicARecord == nil {
		i.publicARecord, err = newDynamicDnsARecord(i, i.dns, i.PublicHostname(), i.subnet.Access(), i.Pod().Cluster().AuditIgnore())
		if err != nil {
			return err
		}
	}
	return nil
}

func (i *Instance) loadDnsARecords(req *route.Request) route.Response {
	log.Debug("Loading Instance %q, DNS A Records", i.Name())

	// Load the dns records
	if i.privateARecord.Route(req) != route.OK {
		return route.FAIL
	}
	if i.publicARecord != nil && i.publicARecord.Route(req) != route.OK {
		return route.FAIL
	}

	// Fix dns a records if necessary
	if i.updateDnsARecord(req, i.privateARecord, i.PrivateIPAddress(), "private") != route.OK {
		return route.FAIL
	}
	if i.updateDnsARecord(req, i.publicARecord, i.PublicIPAddress(), "public") != route.OK {
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) updateDnsARecord(req *route.Request, r *dnsRecord, ip, t string) route.Response {
	if r != nil && len(r.DynamicValues()) == 1 && ip != "" {
		cIp := r.DynamicValues()[0]
		if cIp != ip {
			// Update the dns record to match the deployed ip address.
			log.Debug("Updating Instance %q, DNS A Record %q", i.Name(), r.Id())
			msg.Warn("Instance %q, configured %s ip %q does not match deployed ip %q", i.Name(), t, cIp, ip)
			c := req.Clone(route.Create)
			c.Flags().Append("skip_created_check")
			if r.Route(c) != route.OK {
				return route.FAIL
			}
		}
	}
	return route.OK
}

func (i *Instance) createDnsARecords(req *route.Request) route.Response {
	r := req.Clone(route.Create)
	if i.privateARecord.Route(r) != route.OK {
		return route.FAIL
	}
	if i.publicARecord != nil && i.publicARecord.Route(r) != route.OK {
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) destroyDnsARecords(req *route.Request) route.Response {
	r := req.Clone(route.Destroy)
	if i.publicARecord != nil && i.publicARecord.Route(r) != route.OK {
		return route.FAIL
	}
	if i.privateARecord.Route(r) != route.OK {
		return route.FAIL
	}
	return route.OK
}
