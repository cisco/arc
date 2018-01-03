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

package config

import "github.com/cisco/arc/pkg/msg"

// Dns configuration conatins a domain name, a subdomain, a provider record
// a list of a records, and a list of cname records.
type Dns struct {
	DomainName_  string      `json:"domain_name"`
	Subdomain_   string      `json:"subdomain"`
	Provider     *Provider   `json:"provider"`
	ARecords     *DnsRecords `json:"a_records"`
	CNameRecords *DnsRecords `json:"cname_records"`
	CacheIgnore  []string    `json:"cache_ignore"`
}

// DomainName satisfies the resource.StaticNetwork interface.
// You most likely want to use the Domain() method instead of this.
func (d *Dns) DomainName() string {
	return d.DomainName_
}

// DnsNameServers satisfies the resource.StaticNetwork interface.
func (d *Dns) Subdomain() string {
	return d.Subdomain_
}

// Domain satisfies the resource.StaticNetwork interface.
// This is a convenience method that will provide the
// domain name being used.
func (d *Dns) Domain() string {
	if d.Subdomain() != "" {
		return d.Subdomain() + "." + d.DomainName()
	}
	return d.DomainName()
}

// PrintLocal provides a user friendly way to view the configuration local to the dns object.
func (d *Dns) PrintLocal() {
	msg.Info("DNS Config")
	msg.Detail("%-20s\t%s", "domain_name", d.DomainName())
	msg.Detail("%-20s\t%s", "subdomain", d.Subdomain())
	msg.Detail("%-20s\t%s", "domain", d.Domain())
}

// Print provides a user friendly way to view the datacenter configuration.
func (d *Dns) Print() {
	d.PrintLocal()
	msg.IndentInc()
	if d.Provider != nil {
		d.Provider.Print()
	}
	if d.ARecords != nil {
		d.ARecords.Print("A")
	}
	if d.CNameRecords != nil {
		d.CNameRecords.Print("CNAME")
	}
	msg.IndentDec()
}
