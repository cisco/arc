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
	provider *storageProvider
	s3       *s3.S3

	storage *storage

	access string
	bucket *s3.Bucket
	arn    *arn
}

func newBucket(bkt resource.Bucket, cfg *config.Bucket, prov *storageProvider) (resource.ProviderBucket, error) {
	log.Debug("Initializing AWS Bucket %q", cfg.Name())

	b := &bucket{
		Bucket:   cfg,
		provider: prov,
		storage:  bkt.Storage().ProviderStorage().(*storage),
		s3:       prov.s3[cfg.Region()],
	}
	b.arn = newIamRole(prov.number, b.Role())

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

func (b *bucket) enableVersioning() error {
	log.Debug("Enabling Bucket Versioning")
	params := &s3.PutBucketVersioningInput{
		Bucket: aws.String(b.Name()),
		VersioningConfiguration: &s3.VersioningConfiguration{
			MFADelete: aws.String("Disabled"),
			Status:    aws.String("Enabled"),
		},
	}
	_, err := b.s3.PutBucketVersioning(params)
	if err != nil {
		return err
	}
	return nil
}

func (b *bucket) EnableEncryption(key resource.EncryptionKey) error {
	log.Debug("Enabling Bucket Encryption")
	bucketKey := key.ProviderEncryptionKey().(*encryptionKey)
	params := &s3.PutBucketEncryptionInput{
		Bucket: aws.String(b.Name()),
		ServerSideEncryptionConfiguration: &s3.ServerSideEncryptionConfiguration{
			Rules: []*s3.ServerSideEncryptionRule{
				{
					ApplyServerSideEncryptionByDefault: &s3.ServerSideEncryptionByDefault{
						KMSMasterKeyID: bucketKey.encryptionKey.TargetKeyId,
						SSEAlgorithm:   aws.String("aws:kms"),
					},
				},
			},
		},
	}
	_, err := b.s3.PutBucketEncryption(params)
	if err != nil {
		return err
	}
	msg.Detail("Bucket encryption: enabled")
	return nil
}

func (b *bucket) EnableReplication(key resource.EncryptionKey) error {
	log.Debug("Enabling Bucket Replication")
	if b.Role() == "" || b.Destination() == "" {
		return fmt.Errorf("No Role or Destination found for replication")
	}
	params := &s3.PutBucketReplicationInput{
		Bucket: aws.String(b.Name()),
		ReplicationConfiguration: &s3.ReplicationConfiguration{
			Role: aws.String(b.arn.String()),
			Rules: []*s3.ReplicationRule{
				{
					Destination: &s3.Destination{
						Bucket: aws.String((&arn{
							service:    "s3",
							relativeId: b.Destination(),
						}).String()),
					},
					Prefix: aws.String(""),
					Status: aws.String("Enabled"),
				},
			},
		},
	}

	if key != nil {
		log.Debug("Enabling Encrypted Bucket Replication")
		params.ReplicationConfiguration.Rules[0].SourceSelectionCriteria = &s3.SourceSelectionCriteria{
			SseKmsEncryptedObjects: &s3.SseKmsEncryptedObjects{
				Status: aws.String("Enabled"),
			},
		}
		bucketKey := key.ProviderEncryptionKey().(*encryptionKey)
		params.ReplicationConfiguration.Rules[0].Destination.EncryptionConfiguration = &s3.EncryptionConfiguration{
			ReplicaKmsKeyID: bucketKey.encryptionKey.AliasArn,
		}
	}

	_, err := b.s3.PutBucketReplication(params)
	if err != nil {
		return err
	}
	return nil
}

func (b *bucket) Route(req *route.Request) route.Response {
	log.Route(req, "AWS Bucket %q", b.Name())
	return route.OK
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
	if err := b.Load(); err != nil {
		return err
	}
	msg.Detail("Bucket created: %s", b.Name())
	err = b.enableVersioning()
	if err != nil {
		return err
	}
	msg.Detail("Bucket versioning: enabled")

	return nil
}

func (b *bucket) Load() error {
	if bucket := b.storage.bucketCache.find(b); bucket != nil {
		log.Debug("Skipping Bucket load, cached...")
		b.set(bucket)
		return nil
	}
	params := &s3.ListBucketsInput{}
	resp, err := b.s3.ListBuckets(params)
	if err != nil {
		return err
	}
	for _, bucket := range resp.Buckets {
		if aws.StringValue(bucket.Name) == b.Name() {
			b.bucket = bucket
		}
	}
	return nil
}
func (b *bucket) Destroy(flags ...string) error {
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
	b.clear()
	return nil
}

func (b *bucket) Provision(flags ...string) error {
	msg.Info("Bucket Provision")
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
