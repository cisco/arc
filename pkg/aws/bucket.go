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
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type bucket struct {
	*config.Bucket
	storage *storage
	s3      *s3.S3

	access string
	bucket *s3.Bucket
}

func newBucket(bkt resource.Bucket, cfg *config.Bucket, s *s3.S3) (resource.ProviderBucket, error) {
	log.Debug("Initializing AWS Bucket %q", cfg.Name())

	b := &bucket{
		Bucket:  cfg,
		storage: bkt.Storage().ProviderStorage().(*storage),
		s3:      s,
	}
	b.set(b.storage.bucketCache.find(b))

	return b, nil
}

func (b *bucket) SetTags(tags map[string]string) error {
	log.Debug("Tagging bucket %q", b.Name())
	log.Debug("Tags = %v+", tags)
	tagSet := []*s3.Tag{}
	for k, v := range tags {
		tag := &s3.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}
		tagSet = append(tagSet, tag)
	}
	for k, v := range b.SecurityTags() {
		tag := &s3.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}
		tagSet = append(tagSet, tag)
	}
	params := &s3.PutBucketTaggingInput{
		Bucket: aws.String(b.Name()),
		Tagging: &s3.Tagging{
			TagSet: tagSet,
		},
	}
	_, err := b.s3.PutBucketTagging(params)
	if err != nil {
		return err
	}
	return nil
}

func (b *bucket) Info() {
	if b.bucket == nil {
		return
	}
	msg.Info("Bucket")
	params := &s3.GetBucketAclInput{
		Bucket: aws.String(b.Name()),
	}
	msg.Detail("%-20s\t%s", "name", aws.StringValue(b.bucket.Name))
	msg.Detail("%-20s\t%+v", "date created", b.bucket.CreationDate)
	resp, err := b.s3.GetBucketAcl(params)
	if err != nil {
		msg.Error(err.Error())
		return
	}
	msg.Detail("Permissions")
	for _, v := range resp.Grants {
		msg.IndentInc()
		msg.Detail("%-20s\t%s", "grantee", aws.StringValue(v.Grantee.Type))
		msg.Detail("%-20s\t%s", "permission", aws.StringValue(v.Permission))
		msg.IndentDec()
	}
}

func (b *bucket) Route(req *route.Request) route.Response {
	log.Route(req, "AWS Bucket %q", b.Name())
	switch req.Command() {
	case route.Info:
		b.Info()
		return route.OK
	case route.Config:
		b.Print()
		return route.OK
	case route.Create:
		if err := b.Create(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		if req.Flag("noprovision") {
			return route.OK
		}
	case route.Destroy:
		if err := b.destroy(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	}
	return route.FAIL
}

func (b *bucket) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit Object does not exist")
	}
	if b.bucket == nil {
		a.Audit(aaa.Configured, "%s", b.Name())
	}
	return nil
}

func (b *bucket) set(bucket *s3.Bucket) {
	b.bucket = bucket
}

func (b *bucket) clear() {
	b.bucket = nil
}

func (b *bucket) Created() bool {
	return b.bucket != nil
}

func (b *bucket) Destroyed() bool {
	return b.bucket == nil
}

func (b *bucket) Create(flags ...string) error {
	msg.Info("Bucket Create: %s", b.Name())
	msg.Detail("Bucket Region: %s", b.Region())
	params := &s3.CreateBucketInput{
		ACL:    aws.String("private"),
		Bucket: aws.String(b.Name()),
	}
	if b.Region() != "us-east-1" {
		params.CreateBucketConfiguration = &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(b.Region()),
		}
	}
	_, err := b.s3.CreateBucket(params)
	if err != nil {
		return err
	}
	msg.Detail("Bucket created: %s", b.Name())
	return nil
}

func (b *bucket) destroy(flags ...string) error {
	msg.Info("Bucket Deletion: %s", b.Name())
	msg.Detail("Bucket Region: %s", b.Region())
	params := &s3.DeleteBucketInput{
		Bucket: aws.String(b.Name()),
	}
	_, err := b.s3.DeleteBucket(params)
	if err != nil {
		return err
	}
	msg.Detail("Bucket deleted: %s", b.Name())
	return nil
}
