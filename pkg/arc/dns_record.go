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

type dnsRecord struct {
	*config.DnsRecord
	dns        *dns
	recordType string

	instance    resource.Instance
	ipType      string
	auditIgnore bool

	pod               resource.Pod
	providerDnsRecord resource.ProviderDnsRecord
}

// newDnsRecord is a constructor for a dnsRecord.
func newDnsRecord(dns *dns, cfg *config.DnsRecord, t string) (*dnsRecord, error) {
	r := &dnsRecord{
		DnsRecord:  cfg,
		dns:        dns,
		recordType: t,
	}
	var err error
	r.providerDnsRecord, err = dns.provider.NewDnsRecord(r, cfg)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// newDnsARecord is a constructor for an "A" type resource.DnsRecord.
func newDnsARecord(dns *dns, cfg *config.DnsRecord) (*dnsRecord, error) {
	log.Debug("Initializing DNS A Record %q", cfg.Name())
	return newDnsRecord(dns, cfg, "A")
}

// newDynamicDnsARecord is a constructor for an "A" type resource.DnsRecord for dynamic records created by instances.
func newDynamicDnsARecord(i resource.Instance, dns *dns, hostname, ipType string, ignoreAudit bool) (*dnsRecord, error) {
	r, err := newDnsARecord(dns, &config.DnsRecord{
		Name_: hostname,
		Ttl_:  300,
	})
	if err != nil {
		return nil, err
	}
	r.instance = i
	r.ipType = ipType
	r.auditIgnore = ignoreAudit

	dns.aRecords.dnsRecords[hostname] = r
	dns.aRecords.Append(r)
	return r, nil
}

// newDnsCNameRecord is a constructor for an "CNAME" type resource.DnsRecord.
func newDnsCNameRecord(dns *dns, cfg *config.DnsRecord) (*dnsRecord, error) {
	log.Debug("Initializing DNS CNAME Record %q", cfg.Name())
	return newDnsRecord(dns, cfg, "CNAME")
}

// Dns provides access to the dns record's parent.
func (r *dnsRecord) Dns() resource.Dns {
	return r.dns
}

func (r *dnsRecord) Audit(flags ...string) error {
	if r.auditIgnore {
		return nil
	}
	return r.providerDnsRecord.Audit(flags...)
}

// The provider id of the dns record.
func (r *dnsRecord) Id() string {
	return r.providerDnsRecord.Id()
}

// The record type, "A" or "CNAME"
func (r *dnsRecord) Type() string {
	return r.recordType
}

func (r *dnsRecord) DynamicValues() []string {
	return r.providerDnsRecord.DynamicValues()
}

// Route satisfies the embedded resource.Resource interface in resource.DnsRecord.
func (r *dnsRecord) Route(req *route.Request) route.Response {
	log.Route(req, "DnsRecord %s %q", r.Type(), r.Name())

	// The path should be empty at this point.
	if req.Top() != "" {
		r.help()
		return route.FAIL
	}

	// Skip if the test flag is set.
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load:
		if r.Type() == "CNAME" {
			if err := r.preload(); err != nil {
				msg.Error(err.Error())
				return route.FAIL
			}
		}
		if err := r.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Create:
		if resp := r.preCreate(); resp != route.CONTINUE {
			return resp
		}
		return r.providerDnsRecord.Route(req)
	case route.Provision:
		if resp := r.preProvision(req); resp != route.CONTINUE {
			return resp
		}
		return r.providerDnsRecord.Route(req)
	case route.Destroy:
		return r.providerDnsRecord.Route(req)
	case route.Help:
		r.help()
		return route.OK
	case route.Config:
		r.config()
		return route.OK
	case route.Info:
		r.info(req)
		return route.OK
	default:
		msg.Error("Unknown dns %s command %q.", strings.ToLower(r.Type()), req.Command().String())
	}
	return route.FAIL
}

func (r *dnsRecord) preload() error {
	// The dns record config must either have a pod set or the values set.
	if len(r.Values()) < 1 && r.Pod() == "" {
		return fmt.Errorf("Cannot create dns cname record for %s, no values nor pod present.", r.Name())
	}
	// If values aren't set and pod is set, remember the associated pod. We will populate the values with the
	// fqdn of a created instance in the pod during create.
	if len(r.Values()) < 1 && r.Pod() != "" {
		p := r.Dns().DataCenter().Compute().Clusters().FindPod(r.Pod())
		if p == nil {
			return fmt.Errorf("Cannot find pod %s configured for dns cname record %s", r.Pod(), r.Name())
		}
		log.Debug("Dns %s Record Load: Using pod %s, servertype: %s", r.Type(), p.Name(), p.ServerType())
		r.pod = p
		r.auditIgnore = p.Cluster().AuditIgnore()
	}
	return nil
}

func (r *dnsRecord) preCreate() route.Response {
	switch r.Type() {
	case "A":
		return r.preCreateA()
	case "CNAME":
		return r.preCreateCName()
	}
	msg.Error("Unknown dns record type %p", r.Type())
	return route.FAIL
}

func (r *dnsRecord) preCreateA() route.Response {
	// If the instance doesn't exist, this should be a static dns A record with a value. Proceed with record creation.
	if r.instance == nil {
		return route.CONTINUE
	}

	// This is a dns A record associated with an instance. Only proceed with creation if
	// the instance has been created, and therefore has ip address information associated with it.
	if !r.instance.Created() {
		return route.OK
	}

	var ip string
	switch r.ipType {
	case "private":
		ip = r.instance.PrivateIPAddress()
	case "public", "public_elastic":
		ip = r.instance.PublicIPAddress()
	}
	if ip == "" {
		msg.Error("Missing ip address for dns A record %s", r.Name())
		return route.FAIL
	}
	r.SetValues([]string{ip})
	return route.CONTINUE
}

func (r *dnsRecord) preCreateCName() route.Response {
	// Proceed if the cname isn't based on a pod.
	if r.pod == nil {
		return route.CONTINUE
	}
	// We need to populate the values based on the name of an instance
	// in the pod. Look for the first created instance in the pod.
	for _, j := range r.pod.Instances().(*instances).Get() {
		i := j.(resource.Instance)
		if !i.Created() {
			continue
		}
		value := i.PrivateFQDN()
		if r.Access() == "public" || r.Access() == "public_elastic" {
			value = i.PublicFQDN()
		}
		log.Debug("Dns %s Record Create: Using value %s for pod %s, servertype: %s", r.Type(), value, r.pod.Name(), r.pod.ServerType())
		r.SetValues([]string{value})
		return route.CONTINUE
	}
	return route.OK
}

func (r *dnsRecord) preProvision(req *route.Request) route.Response {
	switch r.Type() {
	case "A":
		// create and provision are the same for A records
		return r.preCreateA()
	case "CNAME":
		return r.preProvisionCName(req)
	}
	msg.Error("Unknown dns record type %p", r.Type())
	return route.FAIL
}

func (r *dnsRecord) preProvisionCName(req *route.Request) route.Response {
	// Proceed if the cname isn't based on a pod.
	if r.pod == nil {
		return route.CONTINUE
	}

	// Attempting to set the cname to a given instance.
	for _, f := range req.Flags().Get() {
		if strings.HasPrefix(f, r.pod.Name()) {
			i := r.pod.FindInstance(f)
			if i != nil {
				value := i.PrivateFQDN()
				if r.Access() == "public" || r.Access() == "public_elastic" {
					value = i.PublicFQDN()
				}
				r.SetValues([]string{value})
				return route.CONTINUE
			}
		}
	}

	// Sets the cname to the secondary instance since there was no instance given.
	s := r.pod.SecondaryInstances()
	if s == nil || len(s) < 1 {
		return route.OK
	}
	i := s[0]

	value := i.PrivateFQDN()
	if r.Access() == "public" || r.Access() == "public_elastic" {
		value = i.PublicFQDN()
	}
	log.Debug("Dns %s Record Update: Using value %s for pod %s, servertype: %s", r.Type(), value, r.pod.Name(), r.pod.ServerType())
	r.SetValues([]string{value})
	return route.CONTINUE
}

func (r *dnsRecord) Load() error {
	return r.providerDnsRecord.Load()
}

// Created satisfies the embedded resource.Resource interface in resource.DnsRecord.
// It delegates the call to the provider's dns record.
func (r *dnsRecord) Created() bool {
	return r.providerDnsRecord.Created()
}

// Destroyed satisfies the embedded resource.Resource interface in resource.DnsRecorrd.
// It delegates the call to the provider's dns record.
func (r *dnsRecord) Destroyed() bool {
	return r.providerDnsRecord.Destroyed()
}

func (r *dnsRecord) help() {
	n := r.Name()
	t := strings.ToLower(r.Type())
	commands := []help.Command{
		{Name: route.Create.String(), Desc: fmt.Sprintf("create %s dns %s record", n, t)},
		{Name: route.Destroy.String(), Desc: fmt.Sprintf("destroy %s dns %s record", n, t)},
		{Name: route.Config.String(), Desc: fmt.Sprintf("show the %s dns %s record configuration", n, t)},
		{Name: route.Info.String(), Desc: fmt.Sprintf("show information about allocated %s dns %s record", n, t)},
		{Name: route.Help.String(), Desc: "show this help"},
	}
	help.Print(fmt.Sprintf("dns %s %s", n, t), commands)
}

func (r *dnsRecord) config() {
	r.DnsRecord.Print(r.Type())
}

func (r *dnsRecord) info(req *route.Request) {
	if r.Destroyed() {
		return
	}
	msg.Info("Dns %s Record", r.Type())
	msg.IndentInc()
	r.PrintLocal(r.Type())
	r.providerDnsRecord.Route(req)
	msg.IndentDec()
}
