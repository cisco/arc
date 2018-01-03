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

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/resource"
)

// dns implements the resource.ProviderDns interface.
type dns struct {
	*config.Dns
	route53 *route53.Route53
	id      string
	cache   *dnsCache
}

// newDns constructs the aws dns.
func newDns(cfg *config.Dns, p *dnsProvider) (resource.ProviderDns, error) {
	log.Info("Initializing AWS Dns")

	d := &dns{
		Dns:     cfg,
		route53: p.route53,
	}

	// Lookup our hosted zone id
	params := &route53.ListHostedZonesByNameInput{
		DNSName: aws.String(d.DomainName()),
	}
	resp, err := d.route53.ListHostedZonesByName(params)
	if err != nil {
		return nil, err
	}

	hostedZoneId := ""
	for _, hostedZone := range resp.HostedZones {
		if hostedZone.Name != nil {
			if d.DomainName() == *hostedZone.Name || (d.DomainName()+".") == *hostedZone.Name {
				if hostedZone.Id != nil {
					hostedZoneId = *hostedZone.Id
					break
				}
			}
		}
	}
	if hostedZoneId == "" {
		return nil, fmt.Errorf("Cannot find Hosted Zone ID for domain name %s", d.DomainName())
	}
	log.Info("AWS Dns Hosted Zone ID: %s", hostedZoneId)
	d.id = hostedZoneId

	d.cache, err = newDnsCache(d)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *dns) Id() string {
	return d.id
}

func (d *dns) AuditDnsRecords(flags ...string) error {
	return d.cache.audit(flags...)
}
