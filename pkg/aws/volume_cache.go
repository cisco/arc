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
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/log"
)

type volumeCacheEntry struct {
	deployed   *ec2.Volume
	configured *volume
}

type volumeCache struct {
	cache map[string]*volumeCacheEntry
}

func newVolumeCache(c *compute) (*volumeCache, error) {
	log.Debug("Initializing AWS Volume Cache")

	v := &volumeCache{
		cache: map[string]*volumeCacheEntry{},
	}

	var token *string
	done := false

	for !done {
		params := &ec2.DescribeVolumesInput{
			MaxResults: aws.Int64(500),
			NextToken:  token,
		}
		resp, err := c.ec2.DescribeVolumes(params)
		if err != nil {
			return nil, err
		}
		log.Debug("Load Volumes: %d", len(resp.Volumes))

		for _, volume := range resp.Volumes {
			if volume == nil || volume.VolumeId == nil {
				log.Verbose("Skipping %+v", volume)
				continue
			}

			if volume.Tags == nil {
				log.Verbose("Untagged volume")
				if volume.VolumeId != nil {
					log.Verbose("\t\t%s", *volume.VolumeId)
				}
				continue
			}

			// Get the datacenter from the tags.
			dc := ""
			for _, t := range volume.Tags {
				if t.Key != nil && *t.Key == "DataCenter" && t.Value != nil {
					dc = *t.Value
				}
			}

			if dc == "" || dc != c.Compute.Name() {
				continue
			}

			id := *volume.VolumeId
			log.Debug("Caching volume %s", id)
			v.cache[id] = &volumeCacheEntry{deployed: volume}
		}

		if resp.NextToken != nil {
			token = resp.NextToken
		} else {
			done = true
		}
	}

	return v, nil
}

func (c *volumeCache) find(v *volume) *ec2.Volume {
	e := c.cache[v.Id()]
	if e == nil {
		return nil
	}
	e.configured = v
	return e.deployed
}

func (c *volumeCache) remove(v *volume) {
	log.Debug("Deleting %s from volumeCache", v.Id())
	delete(c.cache, v.Id())
}

func (c *volumeCache) audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	err := aaa.NewAuditWithOptions(flags[0], true, false, false)
	if err != nil {
		return err
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit Object does not exist")
	}

	for _, v := range c.cache {
		switch *v.deployed.State {
		case "in-use", "pending":
			continue
		default:
			a.Audit(aaa.Deployed, "Volume %q is not attached to any instance", *v.deployed.VolumeId)
		}
	}
	return nil
}
