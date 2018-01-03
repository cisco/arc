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

package mock

import (
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/resource"
)

// dnsRecord implements the resource.ProviderDnsRecord interface.
type dnsRecord struct {
	*mock
	*config.DnsRecord
	id         string
	recordType string
}

// newDnsRecord constructs the mock dnsRecord.
func newDnsRecord(r resource.DnsRecord, cfg *config.DnsRecord, p *dnsProvider) (resource.ProviderDnsRecord, error) {
	log.Info("Initializing mock dnsRecord")
	return &dnsRecord{
		mock:       newMock("dnsRecord", p.Provider),
		DnsRecord:  cfg,
		id:         "0xdeadbeef",
		recordType: "A",
	}, nil
}

func (r *dnsRecord) Audit(flags ...string) error {
	return nil
}

func (r *dnsRecord) Id() string {
	return r.id
}

func (r *dnsRecord) Type() string {
	return r.recordType
}

func (r *dnsRecord) DynamicValues() []string {
	return []string{}
}
