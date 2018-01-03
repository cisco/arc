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
	"strings"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type dnsRecords struct {
	*resource.Resources
	*config.DnsRecords
	dns        *dns
	recordType string
	dnsRecords map[string]resource.DnsRecord
	podRecords map[string][]resource.DnsRecord
}

// newDnsRecords is a constructor for a dnsRecords object. It returns a non-nil error
// upon failure. You want to use either newDnsARecords or newDnsCNameRecords instead.
func newDnsRecords(dns *dns, cfg *config.DnsRecords, t string) (*dnsRecords, error) {
	log.Debug("Initializing DNS %s Records", t)

	if t != "A" && t != "CNAME" {
		return nil, fmt.Errorf("Unknown dns record type %s", t)
	}

	d := &dnsRecords{
		Resources:  resource.NewResources(),
		DnsRecords: cfg,
		dns:        dns,
		recordType: t,
		dnsRecords: map[string]resource.DnsRecord{},
		podRecords: map[string][]resource.DnsRecord{},
	}

	for _, conf := range *cfg {
		if d.Find(conf.Name()) != nil {
			return nil, fmt.Errorf("DNS %s Record name %q must be unique but is used multiple times", t, conf.Name())
		}
		var record *dnsRecord
		var err error
		switch t {
		case "A":
			record, err = newDnsARecord(dns, conf)
		case "CNAME":
			record, err = newDnsCNameRecord(dns, conf)
		default:
			return nil, fmt.Errorf("Unknown dns record type %s for %s", t, conf.Name())
		}
		if err != nil {
			return nil, err
		}
		d.dnsRecords[conf.Name()] = record
		d.Append(record)
		pod := record.Pod()
		if pod != "" {
			if d.podRecords[pod] == nil {
				d.podRecords[pod] = []resource.DnsRecord{record}
			} else {
				d.podRecords[pod] = append(d.podRecords[pod], record)
			}
		}
	}
	return d, nil
}

// newDnsARecords is a constructor for a dnsRecords object given a list of A records via config.DnsRecords.
func newDnsARecords(dns *dns, cfg *config.DnsRecords) (*dnsRecords, error) {
	return newDnsRecords(dns, cfg, "A")
}

// newDnsCNameRecords is a constructor for a dnsRecords object given a list of CNAME records via config.DnsRecords.
func newDnsCNameRecords(dns *dns, cfg *config.DnsRecords) (*dnsRecords, error) {
	return newDnsRecords(dns, cfg, "CNAME")
}

// Find satisfies the resource.DnsRecords interface and provides a way
// to search for a specific dns record. This assumes dns record names are unique.
func (d *dnsRecords) Find(name string) resource.DnsRecord {
	// Remove the domain name if it exists.
	if strings.HasSuffix(name, d.dns.Domain()) {
		s := strings.Split(name, "."+d.dns.Domain())
		name = s[0]
	}
	return d.dnsRecords[name]
}

// FindByPod returns the DnsRecords associated with the given pod name.
func (d *dnsRecords) FindByPod(pod string) []resource.DnsRecord {
	return d.podRecords[pod]
}

func (d *dnsRecords) Type() string {
	return d.recordType
}

// Route satisfies the embedded resource.Resource interface in resource.DnsRecords.
func (d *dnsRecords) Route(req *route.Request) route.Response {
	log.Route(req, "Dns %s Records", d.recordType)

	// Is the resource the name of a dns record?
	if record := d.Find(req.Top()); record != nil {
		return record.Route(req.Pop())
	}

	// The path should be empty at this point.
	if req.Top() != "" {
		msg.Error("Unknown dns record %q.", req.Top())
		return route.FAIL
	}

	// Skip if the test flag is set.
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Handle the command
	switch req.Command() {
	case route.Load, route.Create, route.Provision:
		return d.RouteInOrder(req)
	case route.Destroy:
		return d.RouteReverseOrder(req)
	case route.Help:
		d.help()
		return route.OK
	case route.Info:
		d.info(req)
		return route.OK
	case route.Config:
		d.config()
		return route.OK
	default:
		msg.Error("Unknown dns record command %q.", req.Command().String())
	}
	return route.FAIL
}

func (d *dnsRecords) help() {
	t := strings.ToLower(d.Type())
	commands := []help.Command{
		{route.Create.String(), fmt.Sprintf("create all dns %s records", t)},
		{route.Destroy.String(), fmt.Sprintf("destroy all dns %s records", t)},
		{"'name'", fmt.Sprintf("manage named dns %s record", t)},
		{route.Config.String(), fmt.Sprintf("show the dns %s records configuration", t)},
		{route.Info.String(), fmt.Sprintf("show information about allocated dns %s records", t)},
		{route.Help.String(), "show this help"},
	}
	help.Print(fmt.Sprintf("dns %s", t), commands)
}

func (d *dnsRecords) config() {
	d.DnsRecords.Print(d.Type())
}

func (d *dnsRecords) info(req *route.Request) {
	if d.Destroyed() {
		return
	}
	msg.Info("Dns %s Records", d.Type())
	msg.IndentInc()
	d.RouteInOrder(req)
	msg.IndentDec()
}
