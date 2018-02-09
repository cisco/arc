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

// DnsRecords is a collection of DnsRecord objects.
type DnsRecords []*DnsRecord

// Print provides a user friendly way to view the dns records configuration.
func (d *DnsRecords) Print(s string) {
	msg.Info("DNS %s Records", s)
	msg.IndentInc()
	for _, r := range *d {
		r.Print(s)
	}
	msg.IndentDec()
}

// DnsRecord configuration consists of a name, a ttl, an optional pod, an access value
// and an optional set of values. For cname records that are associated with a pod
// the name, ttl and pod values are mandatory. If this cname needs to be associated with
// the public ip address of an instance in the pod, the access field need to be set
// to "public". For a records the name, ttl and values are required.
type DnsRecord struct {
	Name_   string   `json:"name"`
	Ttl_    int      `json:"ttl"`
	Pod_    string   `json:"pod"`
	Access_ string   `json:"access"`
	Values_ []string `json:"values"`
}

// Name satisfies the resource.StaticDnsRecord interface.
func (d *DnsRecord) Name() string {
	return d.Name_
}

// Ttl satisfies the resource.StaticDnsRecord interface.
func (d *DnsRecord) Ttl() int {
	return d.Ttl_
}

// Pod satisfies the resource.StaticDnsRecord interface.
func (d *DnsRecord) Pod() string {
	return d.Pod_
}

// Access satisfies the resource.StaticDnsRecord interface.
func (d *DnsRecord) Access() string {
	access := d.Access_
	if access == "" {
		access = "private"
	}
	return access
}

// Values satisfies the resource.StaticDnsRecord interface.
func (d *DnsRecord) Values() []string {
	return d.Values_
}

// SetValues will set the values of the record. It is intended for those dns records that are created dynamically.
func (d *DnsRecord) SetValues(v []string) {
	d.Values_ = v
}

// PrintLocal provides a user friendly way to view the configuration local of the dns record.
// This is a shallow print.
func (d *DnsRecord) PrintLocal(s string) {
	msg.Info("DNS %s Record Config", s)
	msg.Detail("%-20s\t%s", "name", d.Name())
	msg.Detail("%-20s\t%d", "ttl", d.Ttl())
	if d.Pod() != "" {
		msg.Detail("%-20s\t%s", "pod", d.Pod())
		msg.Detail("%-20s\t%s", "access", d.Access())
	}
	values, sep := "", ""
	for _, value := range d.Values() {
		values += sep + value
		sep = ", "
	}
	if values != "" {
		msg.Detail("%-20s\t%s", "values", values)
	}
}

// Print provides a user friendly way to view the dns record.
// This is a deep print.
func (d *DnsRecord) Print(s string) {
	d.PrintLocal(s)
}
