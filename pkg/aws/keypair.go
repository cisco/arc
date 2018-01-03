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
	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
)

// keypair implements the resource.ProviderKeyPair interface.
type keypair struct {
	*config.KeyPair
	ec2         *ec2.EC2
	fingerprint string
}

// newKeyPair constructs the aws keypair.
func newKeyPair(cfg *config.KeyPair, c *ec2.EC2) (resource.ProviderKeyPair, error) {
	log.Debug("Initializing AWS KeyPair %q.", cfg.Name())
	return &keypair{
		KeyPair: cfg,
		ec2:     c,
	}, nil
}

func (k *keypair) Route(req *route.Request) route.Response {
	log.Route(req, "AWS KeyPair %q", k.Name())

	switch req.Command() {
	case route.Create:
		return k.create(req)
	case route.Destroy:
		return k.destroy(req)
	case route.Info:
		k.info()
		return route.OK
	}
	return route.FAIL
}

func (k *keypair) FingerPrint() string {
	return k.fingerprint
}

func (k *keypair) Created() bool {
	return k.FingerPrint() != ""
}

func (k *keypair) Destroyed() bool {
	return !k.Created()
}

func (k *keypair) set(fingerprint string) {
	k.fingerprint = fingerprint
}

func (k *keypair) clear() {
	k.fingerprint = ""
}

func (k *keypair) Load() error {
	params := &ec2.DescribeKeyPairsInput{
		Filters: []*ec2.Filter{
			{
				Name: aws.String("key-name"),
				Values: []*string{
					aws.String(k.Name()),
				},
			},
		},
		KeyNames: []*string{
			aws.String(k.Name()),
		},
	}
	keypairs, err := k.ec2.DescribeKeyPairs(params)
	if err != nil {
		// Unlike most other aws resources, DescribeKeyPairs will fail if the resource
		// doesn't exist. This is expected if the resource hasn't been created yet so return OK.
		return nil
	}
	for _, key := range keypairs.KeyPairs {
		if key.KeyName != nil && k.Name() == *key.KeyName {
			k.set(*key.KeyFingerprint)
			return nil
		}
	}
	return nil
}

func (k *keypair) create(req *route.Request) route.Response {
	msg.Info("KeyPair Creation: %s", k.Name())
	if k.Created() {
		msg.Detail("KeyPair exists, skipping...")
		return route.OK
	}

	params := &ec2.ImportKeyPairInput{
		KeyName:           aws.String(k.Name()),
		PublicKeyMaterial: []byte(k.KeyMaterial()),
	}
	key, err := k.ec2.ImportKeyPair(params)
	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if key.KeyFingerprint == nil {
		msg.Error("KeyPair missing fingerprint")
		return route.FAIL
	}
	k.set(*key.KeyFingerprint)

	msg.Detail("Created %s", k.Name())
	aaa.Accounting("KeyPair created: %s, %s", k.Name(), k.FingerPrint())
	return route.OK
}

func (k *keypair) destroy(req *route.Request) route.Response {
	msg.Info("KeyPair Destruction: %s", k.Name())
	if k.Destroyed() {
		msg.Detail("KeyPair does not exist, skipping...")
		return route.OK
	}

	params := &ec2.DeleteKeyPairInput{
		KeyName: aws.String(k.Name()),
	}
	if _, err := k.ec2.DeleteKeyPair(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	k.clear()

	msg.Detail("Destroyed: %s", k.Name())
	aaa.Accounting("KeyPair Destroyed: %s", k.Name())
	return route.OK
}

func (k *keypair) info() {
	if k.Destroyed() {
		return
	}
	msg.Info("KeyPair")
	msg.Detail("%-20s\t%s", "name", k.Name())
	msg.Detail("%-20s\t%s", "fingerprint", k.FingerPrint())
}
