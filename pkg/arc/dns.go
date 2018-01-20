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

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"

	"github.com/cisco/arc/pkg/aws"
	"github.com/cisco/arc/pkg/mock"
	//"github.com/cisco/arc/pkg/gcp"
	//"github.com/cisco/arc/pkg/azure"
)

type dns struct {
	*resource.Resources
	*config.Dns
	arc          *arc
	provider     provider.Dns
	providerDns  resource.ProviderDns
	aRecords     *dnsRecords
	cnameRecords *dnsRecords
	datacenter   *dataCenter
}

// newDns is the constructor for a dns object. It returns a non-nil error upon failure.
func newDns(arc *arc, cfg *config.Dns) (*dns, error) {
	if cfg == nil {
		return nil, nil
	}
	log.Debug("Initializing Dns")

	if cfg.Provider == nil {
		return nil, fmt.Errorf("The provider element is missing from the dns configuration")
	}
	if cfg.CNameRecords == nil {
		return nil, fmt.Errorf("The records element is missing from the dns configuration")
	}

	d := &dns{
		Resources: resource.NewResources(),
		Dns:       cfg,
		arc:       arc,
	}

	vendor := cfg.Provider.Vendor
	var err error

	switch vendor {
	case "mock":
		d.provider, err = mock.NewDnsProvider(cfg)
	case "aws":
		d.provider, err = aws.NewDnsProvider(cfg)
	//case "gcp":
	//	d.provider, err = gcp.NewDnsProvider(cfg)
	//case "azure":
	//	d.provider, err = azure.NewDnsProvider(cfg)
	default:
		err = fmt.Errorf("Unknown vendor %s", vendor)
	}
	if err != nil {
		return nil, err
	}

	d.providerDns, err = d.provider.NewDns(cfg)
	if err != nil {
		return nil, err
	}

	d.aRecords, err = newDnsARecords(d, cfg.ARecords)
	if err != nil {
		return nil, err
	}
	d.Append(d.aRecords)

	d.cnameRecords, err = newDnsCNameRecords(d, cfg.CNameRecords)
	if err != nil {
		return nil, err
	}
	d.Append(d.cnameRecords)

	return d, nil
}

// Id provides the id of the provider specific dns resource.
// This satisfies the resource.DynamicDns interface.
func (d *dns) Id() string {
	if d.providerDns == nil {
		return ""
	}
	return d.providerDns.Id()
}

// Arc satisfies the resource.Dns interface and provides access
// to dns' parent.
func (d *dns) Arc() resource.Arc {
	return d.arc
}

// ProviderDns satisfies the resource.Dns interface and provides access
// to the provider specific dns resource.
func (d *dns) ProviderDns() resource.ProviderDns {
	return d.providerDns
}

// ARecords satisfies the resource.DataCenter interface and provides access
// to dns' a records.
func (d *dns) ARecords() resource.DnsRecords {
	return d.aRecords
}

// CNameRecords satisfies the resource.DataCenter interface and provides access
// to dns' cname records.
func (d *dns) CNameRecords() resource.DnsRecords {
	return d.cnameRecords
}

func (d *dns) AuditDnsRecords(flags ...string) error {
	return d.ProviderDns().AuditDnsRecords(flags...)
}

// associate the DataCenter resource with this Dns resource.
func (d *dns) associate(dc *dataCenter) {
	d.datacenter = dc
}

// DataCenter providess acess to the DataCenter resource.
func (d *dns) DataCenter() resource.DataCenter {
	if d.datacenter == nil {
		return nil
	}
	return d.datacenter
}

// Route satisfies the embedded resource.Resource interface in resource.Dns.
// DataCenter does not directly terminate a request so only handles load and info
// requests from it's parent.  All other commands are routed to arc's children.
func (d *dns) Route(req *route.Request) route.Response {
	log.Route(req, "Dns")

	// Route to the appropriate resource
	switch req.Top() {
	case "":
		break
	case "a":
		return d.aRecords.Route(req.Pop())
	case "cname":
		return d.cnameRecords.Route(req.Pop())
	default:
		d.help()
		return route.FAIL
	}

	// Skip if the test flag is set.
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load, route.Create, route.Provision:
		return d.RouteInOrder(req)
	case route.Destroy:
		return d.RouteReverseOrder(req)
	case route.Help:
		d.help()
		return route.OK
	case route.Config:
		d.config()
		return route.OK
	case route.Info:
		d.info(req)
		return route.OK
	case route.Audit:
		if err := d.Audit("Dns Record"); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	default:
		msg.Error("Unknown dns command '%s'.", req.Command())
	}
	return route.FAIL
}

func (d *dns) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	err := aaa.NewAudit(flags[0])
	if err != nil {
		return err
	}
	if err := d.AuditDnsRecords(flags...); err != nil {
		return err
	}
	for _, v := range d.aRecords.dnsRecords {
		if err := v.Audit(flags...); err != nil {
			return err
		}
	}
	for _, v := range d.cnameRecords.dnsRecords {
		if err := v.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (d *dns) help() {
	commands := []help.Command{
		{route.Create.String(), "create all dns records"},
		{route.Destroy.String(), "destroy all dns records"},
		{"a", "manage dns a records"},
		{"a 'name'", "manage named dns a record"},
		{"cname", "manage dns cname records"},
		{"cname 'name'", "manage named dns cname record"},
		{route.Config.String(), "show the dns configuration"},
		{route.Info.String(), "show information about allocated dns resource"},
		{route.Help.String(), "show this help"},
	}
	help.Print("dns", commands)
}

func (d *dns) config() {
	d.Dns.Print()
}

func (d *dns) info(req *route.Request) {
	if d.Destroyed() {
		return
	}
	msg.Info("Dns")
	msg.IndentInc()
	d.Dns.PrintLocal()
	d.RouteInOrder(req)
	msg.IndentDec()
}
