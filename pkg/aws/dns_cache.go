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
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/log"
)

type dnsCacheEntry struct {
	deployed   *route53.ResourceRecordSet
	configured *dnsRecord
}

type dnsCache struct {
	cache   map[string]*dnsCacheEntry
	unnamed []*route53.ResourceRecordSet
}

func newDnsCache(d *dns) (*dnsCache, error) {
	log.Debug("Initializing AWS DNS Cache")

	c := &dnsCache{
		cache: map[string]*dnsCacheEntry{},
	}

	next := ""
	for {
		params := &route53.ListResourceRecordSetsInput{
			HostedZoneId: aws.String(d.Id()),
		}
		if next != "" {
			params.StartRecordName = aws.String(next)
		}

		resp, err := d.route53.ListResourceRecordSets(params)
		if err != nil {
			return nil, err
		}

		truncated := false
		if resp.IsTruncated != nil {
			truncated = *resp.IsTruncated
		}

		next = ""
		if resp.NextRecordName != nil {
			next = *resp.NextRecordName
		}
		log.Debug("Load DnsRecords: truncated: %t, next record name: %s ", truncated, next)

		for _, r := range resp.ResourceRecordSets {
			if r.Type == nil {
				if r.Name == nil || len(*r.Name) == 0 {
					c.unnamed = append(c.unnamed, r)
					continue
				}
				log.Verbose("Skipping %+v", r)
				continue
			}
			name := strings.Replace(*r.Name, "\\052", "*", -1)
			if name[len(name)-1:] != "." {
				name += "."
			}
			if !strings.HasSuffix(name, d.Domain()+".") {
				continue
			}
			switch *r.Type {
			case "A", "CNAME":
				found := false
				for _, v := range d.CacheIgnore {
					if strings.HasPrefix(*r.Name, v) {
						found = true
						break
					}
				}
				if found {
					continue
				}
				log.Debug("Caching %s", name)
				c.cache[name] = &dnsCacheEntry{deployed: r}
			}
		}
		if truncated == false {
			break
		}
	}

	return c, nil
}

func (c *dnsCache) find(d *dnsRecord) *route53.ResourceRecordSet {
	e := c.cache[d.Id()]
	if e == nil {
		return nil
	}
	e.configured = d
	return e.deployed
}

func (c *dnsCache) remove(d *dnsRecord) {
	log.Debug("Deleting %s from dnsCache", d.Id())
	delete(c.cache, d.Id())
}

func (c *dnsCache) audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit Object does not exist")
	}
	for k, v := range c.cache {
		if v.configured == nil {
			a.Audit(aaa.Deployed, "%s", k)
		}
	}
	if c.unnamed != nil {
		a.Audit(aaa.Deployed, "\r")
		for i, v := range c.unnamed {
			u := "\t" + strings.Replace(fmt.Sprintf("%+v", v), "\n", "\n\t", -1)
			m := fmt.Sprintf("Unnamed Dns Record %d - DnsName: %q %s", i+1, *v.Name, u)
			a.Audit(aaa.Deployed, m)
		}
	}
	return nil
}
