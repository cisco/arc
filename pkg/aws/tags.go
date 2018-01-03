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
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/route"
)

func createTags(c *ec2.EC2, name string, rsrcId string, req *route.Request) error {
	log.Debug("Tagging %s - %s", name, rsrcId)

	params := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(rsrcId),
		},
		Tags: []*ec2.Tag{
			{
				Key:   aws.String("Name"),
				Value: aws.String(name),
			},
			{
				Key:   aws.String("Created By"),
				Value: aws.String(req.UserId()),
			},
			{
				Key:   aws.String("Last Modified"),
				Value: aws.String(req.Time()),
			},
			{
				Key:   aws.String("DataCenter"),
				Value: aws.String(req.DataCenter()),
			},
		},
	}

	_, err := c.CreateTags(params)
	return err
}

func printTags(tags []*ec2.Tag) {
	if tags == nil {
		return
	}
	if os.Getenv("verbose") != "yes" {
		return
	}
	msg.IndentInc()
	msg.Info("Tags")
	for _, t := range tags {
		if t == nil {
			continue
		}
		key := *t.Key
		value := *t.Value
		if key != "Name" {
			msg.Detail("%-20s\t%s", key, value)
		}
	}
	msg.IndentDec()
}

func setTags(c *ec2.EC2, t map[string]string, id string) error {
	tags := []*ec2.Tag{}
	for k, v := range t {
		tags = append(tags, &ec2.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
		log.Debug("Create tag %s: %s", k, v)
	}
	params := &ec2.CreateTagsInput{
		Resources: []*string{
			aws.String(id),
		},
		Tags: tags,
	}
	_, err := c.CreateTags(params)
	return err
}
