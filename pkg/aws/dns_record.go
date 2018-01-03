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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

// dnsRecord implements the resource.ProviderDnsRecord interface.
type dnsRecord struct {
	*config.DnsRecord
	route53 *route53.Route53
	dns     *dns
	record  resource.DnsRecord
	id      string
	rrset   *route53.ResourceRecordSet
}

// newDnsRecord constructs the aws dnsRecord.
func newDnsRecord(rec resource.DnsRecord, cfg *config.DnsRecord, p *dnsProvider) (*dnsRecord, error) {
	log.Debug("Initializing AWS DNS %s Record %q", rec.Type(), cfg.Name())

	// Type assertion to get underlying aws dns.
	dns, ok := rec.Dns().ProviderDns().(*dns)
	if !ok {
		return nil, fmt.Errorf("AWS newDnsRecord: Unable to obtain prov dns associated with record %s", cfg.Name())
	}

	fqdn := cfg.Name() + "." + dns.Domain() + "."

	r := &dnsRecord{
		DnsRecord: cfg,
		route53:   p.route53,
		dns:       dns,
		record:    rec,
		id:        fqdn,
	}
	r.set(r.dns.cache.find(r))
	return r, nil
}

func (r *dnsRecord) Route(req *route.Request) route.Response {
	log.Route(req, "AWS DNS %s Record %q", r.record.Type(), r.Id())

	switch req.Command() {
	case route.Create:
		return r.create(req)
	case route.Provision:
		return r.provision(req)
	case route.Destroy:
		return r.destroy(req)
	case route.Info:
		r.info()
		return route.OK
	}
	return route.FAIL
}

func (r *dnsRecord) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit Object does not exist")
	}
	// Configured but not Deployed
	if !r.Created() {
		a.Audit(aaa.Configured, "Dns Record %q is configured but not deployed", r.Name())
	}
	// Mismatches
	if r.rrset == nil {
		return nil
	}
	if r.rrset.Type != nil && *r.rrset.Type != r.Type() {
		a.Audit(aaa.Mismatched, "Dns Record %q | Configured: %q - Deployed: %q", r.Name(), r.Type(), *r.rrset.Type)
	}
	if r.rrset.TTL != nil && *r.rrset.TTL != int64(r.Ttl()) {
		a.Audit(aaa.Mismatched, "Dns Record %q | Configured: \"%d\" - Deployed: \"%d\"", r.Name(), r.Ttl(), *r.rrset.TTL)
	}
	return nil
}

func (r *dnsRecord) Created() bool {
	return r.rrset != nil
}

func (r *dnsRecord) Destroyed() bool {
	return r.rrset == nil
}

func (r *dnsRecord) Id() string {
	return r.id
}

func (r *dnsRecord) DynamicValues() []string {
	return r.rrValues()
}

func (r *dnsRecord) Type() string {
	return r.record.Type()
}

func (r *dnsRecord) rrType() string {
	if r.rrset == nil || r.rrset.Type == nil {
		return ""
	}
	return *r.rrset.Type
}

func (r *dnsRecord) rrTtl() int64 {
	if r.rrset == nil || r.rrset.TTL == nil {
		return 0
	}
	return *r.rrset.TTL
}

func (r *dnsRecord) rrValues() []string {
	result := []string{}
	if r.rrset == nil || len(r.rrset.ResourceRecords) < 1 {
		return result
	}
	for _, rr := range r.rrset.ResourceRecords {
		if rr.Value == nil {
			continue
		}
		result = append(result, *rr.Value)
	}
	return result
}

func (r *dnsRecord) valuesEqual() bool {
	values := r.Values()
	rrvalues := r.rrValues()

	if values == nil && rrvalues == nil {
		return true
	}

	if values == nil || rrvalues == nil {
		return false
	}

	if len(values) != len(rrvalues) {
		return false
	}

	// If everything in values is found in rrvalues...
	for _, v := range values {
		found := false
		for _, rrv := range rrvalues {
			if v == rrv {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	// If everything in rrvalues is found in values...
	for _, rrv := range rrvalues {
		found := false
		for _, v := range values {
			if rrv == v {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func (r *dnsRecord) set(rrset *route53.ResourceRecordSet) {
	r.rrset = rrset
}

func (r *dnsRecord) clear() {
	r.rrset = nil
}

func (r *dnsRecord) Load() error {

	// Use a cached value if it exists.
	if r.rrset != nil {
		log.Debug("Skipping dns record load, cached...")
		return nil
	}

	// Ask AWS for the record
	params := &route53.ListResourceRecordSetsInput{
		HostedZoneId:    aws.String(r.dns.Id()),
		MaxItems:        aws.String("1"),
		StartRecordName: aws.String(r.Id()),
		StartRecordType: aws.String(r.Type()),
	}
	resp, err := r.route53.ListResourceRecordSets(params)
	if err != nil {
		log.Error("%#v\n", params)
		return err
	}

	// The record may not exist yet
	l := len(resp.ResourceRecordSets)
	if l < 1 {
		return nil
	}

	// We should only have one record since dns names are required to be unique.
	if l > 1 {
		log.Warn("AWS dnsRecord:load(), asked for 1 record set, returned %d", len(resp.ResourceRecordSets))
	}
	rrset := resp.ResourceRecordSets[0]

	// The record may not exist yet
	if rrset.Name != nil && *rrset.Name != r.Id() {
		return nil
	}

	r.set(rrset)
	return nil
}

func (r *dnsRecord) reload() bool {
	// Clear the cached value
	r.rrset = nil

	log.Debug("Reloading %s", r.Name())
	return msg.Wait(
		fmt.Sprintf("Waiting for DNS %s Record %s to become available", r.Type(), r.Name()), // title
		fmt.Sprintf("Dns %s Record %s never became available", r.Type(), r.Name()),          // err
		60,        // duration
		r.Created, // test()
		func() bool { // load()
			if err := r.Load(); err != nil {
				msg.Error(err.Error())
				return false
			}
			return true
		},
	)
}

func (r *dnsRecord) sync(changeInfo *route53.ChangeInfo) bool {
	log.Debug("Syncing %s", r.Name())

	if changeInfo == nil {
		return false
	}

	id := *changeInfo.Id
	status := *changeInfo.Status

	return msg.Wait(
		fmt.Sprintf("Waiting for DNS %s Record %s to sync", r.Type(), r.Name()), // title
		fmt.Sprintf("Dns %s Record %s didn't sync", r.Type(), r.Name()),         // err
		600,
		func() bool { // test()
			return status == "INSYNC"
		},
		func() bool { // load()
			resp, err := r.route53.GetChange(&route53.GetChangeInput{Id: aws.String(id)})
			if err != nil || resp.ChangeInfo == nil || resp.ChangeInfo.Id == nil || resp.ChangeInfo.Status == nil {
				return false
			}
			id = *resp.ChangeInfo.Id
			status = *resp.ChangeInfo.Status
			return true
		},
	)
}

func (r *dnsRecord) create(req *route.Request) route.Response {
	values, sep := "", ""
	for _, value := range r.Values() {
		values += sep + value
		sep = ", "
	}

	if req.Flag("skip_created_check") {
		msg.Info("DnsRecord %s Update: %s %s", r.Type(), r.Id(), values)
	} else {
		msg.Info("DnsRecord %s Creation: %s %s", r.Type(), r.Id(), values)
		if r.Created() {
			msg.Detail("DnsRecord exists, skipping...")
			return route.OK
		}
	}

	resourceRecords := []*route53.ResourceRecord{}
	for _, value := range r.Values() {
		resourceRecords = append(resourceRecords, &route53.ResourceRecord{Value: aws.String(value)})
	}
	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(r.Id()),
						Type:            aws.String(r.Type()),
						ResourceRecords: resourceRecords,
						TTL:             aws.Int64(int64(r.Ttl())),
					},
				},
			},
		},
		HostedZoneId: aws.String(r.dns.Id()),
	}

	resp, err := r.route53.ChangeResourceRecordSets(params)
	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	msg.Detail("Created: %s", r.Id())

	if !r.reload() {
		return route.FAIL
	}

	if req.Flag("dnssync") && !r.sync(resp.ChangeInfo) {
		return route.FAIL
	}

	return route.OK
}

func (r *dnsRecord) provision(req *route.Request) route.Response {
	req.Flags().Append("skip_created_check")
	resp := r.create(req)
	req.Flags().Remove("skip_created_check")
	return resp
}

func (r *dnsRecord) destroy(req *route.Request) route.Response {
	msg.Info("DnsRecord %s Destruction: %s", r.Type(), r.Id())
	if r.Destroyed() {
		msg.Detail("DnsRecord does not exist, skipping...")
		return route.OK
	}

	params := &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				{
					Action:            aws.String("DELETE"),
					ResourceRecordSet: r.rrset,
				},
			},
		},
		HostedZoneId: aws.String(r.dns.Id()),
	}

	_, err := r.route53.ChangeResourceRecordSets(params)
	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	msg.Detail("Destroyed: %s", r.Id())
	r.clear()
	r.dns.cache.remove(r)
	return route.OK
}

func (r *dnsRecord) info() {
	if r.Destroyed() {
		return
	}
	msg.Info("Dns %s Record", r.Type())
	msg.Detail("%-20s\t%s", "id", r.Id())
	msg.Detail("%-20s\t%d", "ttl", r.rrTtl())
	msg.Detail("%-20s\t%q", "values", r.rrValues())
}
