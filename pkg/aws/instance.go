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
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

// instance implements the resource.ProviderInstance interface.
type instance struct {
	*config.Instance
	provider *dataCenterProvider
	ec2      *ec2.EC2

	compute   *compute
	network   *network
	subnet    *subnet
	secgroups []*securityGroup

	id             string
	imageId        string
	keyname        string
	roleIdentifier *roleIdentifier
	volumes        []*volume
	instance       *ec2.Instance
	cached         bool
}

// newInstance constructs the aws instance.
func newInstance(in resource.Instance, cfg *config.Instance, p *dataCenterProvider) (*instance, error) {
	log.Debug("Initializing AWS Instance %q", cfg.Name())

	// Type assertion to get underlying aws compute and network.
	c, ok := in.Pod().Cluster().Compute().ProviderCompute().(*compute)
	if !ok {
		return nil, fmt.Errorf("AWS newInstance: Unable to obtain prov compute associated with instance %s", cfg.Name())
	}
	n, ok := in.Network().ProviderNetwork().(*network)
	if !ok {
		return nil, fmt.Errorf("AWS newInstance: Unable to obtain prov network associated with instance %s", cfg.Name())
	}

	// Type assertion to get underlying aws subnet for this instance.
	s, ok := in.Subnet().ProviderSubnet().(*subnet)
	if !ok {
		return nil, fmt.Errorf("AWS newInstance: Unable to obtain subnet %s associated with instance %s", in.Subnet().Name(), cfg.Name())
	}

	// Type assertions to get underlying aws securityGroups associated with this instance.
	secgroups := []*securityGroup{}
	for _, sg := range in.SecurityGroups() {
		secgroup, ok := sg.ProviderSecurityGroup().(*securityGroup)
		if !ok {
			return nil, fmt.Errorf("AWS newInstance: Unable to obtain security group %s associated with instance %s", sg.Name(), cfg.Name())
		}
		secgroups = append(secgroups, secgroup)
	}

	// Get the id of the image to be used with the instance.
	imageId := p.images[cfg.Image()]
	if imageId == "" {
		return nil, fmt.Errorf("newInstance: Unknown image %s", cfg.Image())
	}

	// Get the name of the keypair that will be installed for the root user.
	keyname := in.KeyPair().Name()

	volumes := []*volume{}
	for _, v := range in.ProviderVolumes() {
		volumes = append(volumes, v.(*volume))
	}

	i := &instance{
		Instance:  cfg,
		provider:  p,
		ec2:       p.ec2,
		compute:   c,
		network:   n,
		subnet:    s,
		secgroups: secgroups,
		imageId:   imageId,
		keyname:   keyname,
		volumes:   volumes,
	}
	i.set(c.instanceCache.find(i))
	if i.instance != nil {
		i.cached = true
	}

	for _, v := range in.ProviderVolumes() {
		v.(*volume).associateInstance(i)
	}

	return i, nil
}

func (i *instance) Route(req *route.Request) route.Response {
	log.Route(req, "AWS Instance %q", i.Name())

	switch req.Command() {
	case route.Load:
		if req.Flag("reload") {
			// Clear the cached value
			log.Debug("Clearing cached value for %s", i.Name())
			i.cached = false
		}
		if err := i.load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	case route.Create:
		return i.create(req)
	case route.Destroy:
		return i.destroy(req)
	case route.Start:
		return i.start(req)
	case route.Stop:
		return i.stop(req)
	case route.Restart:
		return i.restart(req)
	}
	return route.FAIL
}

func (i *instance) Created() bool {
	if i.instance == nil || i.State() == "shutting-down" || i.State() == "terminated" {
		return false
	}
	return true
}

func (i *instance) Destroyed() bool {
	return i.instance == nil
}

func (i *instance) Id() string {
	return i.id
}

func (i *instance) ImageId() string {
	return i.imageId
}

func (i *instance) KeyName() string {
	return i.keyname
}

func (i *instance) State() string {
	if i.instance == nil || i.instance.State == nil || i.instance.State.Name == nil {
		return ""
	}
	return *i.instance.State.Name
}

func (i *instance) Started() bool {
	return i.State() == "running"
}

func (i *instance) Stopped() bool {
	return i.State() == "stopped"
}

func (i *instance) PrivateIPAddress() string {
	if i.instance == nil || i.instance.PrivateIpAddress == nil {
		return ""
	}
	return *i.instance.PrivateIpAddress

}

func (i *instance) PublicIPAddress() string {
	if i.instance == nil || i.instance.PublicIpAddress == nil {
		return ""
	}
	return *i.instance.PublicIpAddress
}

func (i *instance) SetTags(t map[string]string) error {
	return setTags(i.ec2, t, i.Id())
}

func (i *instance) Audit(flags ...string) error {
	log.Debug("Instance %s Audit", i.Name())
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find the audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit object doesn't exist")
	}
	// Configured but not deployed
	if i.Destroyed() {
		a.Audit(aaa.Configured, "%s", i.Name())
		return nil
	}
	// Mismatched
	i.compareImage(a)
	i.compareInstanceType(a)
	i.compareRole(a)
	i.compareSecgroups(a)
	i.compareSubnet(a)
	i.compareVolumes(a)
	return nil
}

func (i *instance) compareImage(a *aaa.Audit) {
	d := ""
	if i.instance.ImageId != nil {
		d = i.provider.findById(*i.instance.ImageId)
	}
	if i.instance.ImageId != nil && d != "" && d != i.Image() {
		a.Audit(aaa.Mismatched, "Instance %q | Configured Image: %q - Deployed Image: %q", i.Name(), i.Image(), d)
	}
}

func (i *instance) compareInstanceType(a *aaa.Audit) {
	if i.instance.InstanceType != nil && *i.instance.InstanceType != i.Instance.InstanceType() {
		a.Audit(aaa.Mismatched, "Instance %q | Configured Instance Type: %q - Deployed Instance Type: %q", i.Name(), i.Instance.InstanceType(), *i.instance.InstanceType)
	}
}

func (i *instance) compareRole(a *aaa.Audit) {
	// Role Configured but not deployed
	if i.instance.IamInstanceProfile == nil && i.Instance.Role() != "" {
		a.Audit(aaa.Mismatched, "Instance %q | Configured Role: %q - No Role Deployed", i.Name(), i.Instance.Role())
	}
	if i.instance.IamInstanceProfile != nil && i.instance.IamInstanceProfile.Arn != nil {
		if i.Instance.Role() == "" {
			// Role Deployed but not configured
			a.Audit(aaa.Mismatched, "Instance %q | No Configured Role - Deployed Role: %q", i.Name(), filepath.Base(*i.instance.IamInstanceProfile.Arn))
		} else if i.Instance.Role() != "" && i.Instance.Role() != filepath.Base(*i.instance.IamInstanceProfile.Arn) {
			// Roles do not match
			a.Audit(aaa.Mismatched, "Instance %q | Configured Role: %q - Deployed Role: %q", i.Name(), i.Instance.Role(), filepath.Base(*i.instance.IamInstanceProfile.Arn))
		}
	}
}

func (i *instance) compareSecgroups(a *aaa.Audit) {
	found := false
	if i.instance.SecurityGroups != nil {
		// Secgroup Deployed but not Configured
		for _, d := range i.instance.SecurityGroups {
			log.Debug("Instance Audit - SG - DNC | %q", *d.GroupName)
			found = false
			for _, c := range i.SecurityGroupNames() {
				if d.GroupName != nil && *d.GroupName == c {
					found = true
					break
				}
			}
			if !found {
				a.Audit(aaa.Mismatched, "Instance %q | This instance is a member of Secgroup %q but isn't configured to be", i.Name(), *d.GroupName)
				found = false
			}
		}
		// Secgroup Configured but not Deployed
		for _, c := range i.SecurityGroupNames() {
			log.Debug("Instance Audit - SG - CND | %q", c)
			found = false
			for _, d := range i.instance.SecurityGroups {
				if d.GroupName != nil && *d.GroupName == c {
					found = true
					break
				}
			}
			if !found {
				a.Audit(aaa.Mismatched, "Instance %q | This instance is configured to be a member of Secgroup %q but isn't", i.Name(), c)
				found = false
			}
		}
	}
}

func (i *instance) compareSubnet(a *aaa.Audit) {
	if i.instance.SubnetId != nil && *i.instance.SubnetId != i.subnet.Id() {
		d := i.network.subnetCache.findById(*i.instance.SubnetId)
		depName := ""
		cfgName := i.subnet.Name()
		if d.Tags != nil {
			for _, v := range d.Tags {
				if *v.Key == "Name" {
					depName = *v.Value
				}
			}
		}
		a.Audit(aaa.Mismatched, "Instance %q | Configured Subnet: %q - Deployed Subnet: %q", i.Name(), cfgName, depName)
	}
}

func (i *instance) compareVolumes(a *aaa.Audit) {
	if i.instance.BlockDeviceMappings == nil {
		return
	}
	// Correct Number of Volumes
	if len(i.instance.BlockDeviceMappings) != len(i.volumes) {
		a.Audit(aaa.Mismatched, "Instance %q | Number of Configured Volumes: %d - Number of Deployed Volumes: %d", i.Name(), len(i.volumes), len(i.instance.BlockDeviceMappings))
	}
}

func (i *instance) set(instance *ec2.Instance) {
	if instance == nil || instance.InstanceId == nil {
		return
	}
	i.instance = instance
	i.id = *instance.InstanceId
	for _, v := range i.volumes {
		for _, b := range instance.BlockDeviceMappings {
			if v.Created() {
				continue
			}
			if b.DeviceName != nil && v.Device() == *b.DeviceName {
				v.associateVolume(b)
			}
		}
	}
}

func (i *instance) clear() {
	i.instance = nil
	i.id = ""
}

func (i *instance) load() error {

	// Use a cached value if it exists.
	if i.cached {
		log.Debug("Skipping instance load, cached...")
		if err := i.loadVolumes(); err != nil {
			return err
		}
		return nil
	}

	var token *string
	done := false

	for !done {
		params := &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name: aws.String("vpc-id"),
					Values: []*string{
						aws.String(i.network.vpc.id()),
					},
				},
			},
			NextToken: token,
		}
		if i.Id() != "" {
			params.Filters = append(params.Filters, &ec2.Filter{
				Name: aws.String("instance-id"),
				Values: []*string{
					aws.String(i.Id()),
				},
			})
		} else {
			params.Filters = append(params.Filters, &ec2.Filter{
				Name: aws.String("tag:Name"),
				Values: []*string{
					aws.String(i.Name()),
				},
			})
		}

		resp, err := i.ec2.DescribeInstances(params)
		if err != nil {
			return err
		}

		for _, reservation := range resp.Reservations {
			for _, instance := range reservation.Instances {
				// See if the id's match.
				if i.Id() != "" && instance.InstanceId != nil {
					if i.Id() == *instance.InstanceId {
						i.set(instance)
						log.Debug("Loaded %s, %s, %s -- via id", i.Name(), i.Id(), i.State())
						if err := i.loadVolumes(); err != nil {
							return err
						}
						return nil
					}
				}
				if instance.Tags == nil {
					break
				}
				// If the state is "shutting-down" or "terminated" skip.
				if instance.State != nil && instance.State.Name != nil &&
					(*instance.State.Name == "shutting-down" || *instance.State.Name == "terminated") {
					log.Debug("Skipping instance state %s.", *instance.State.Name)
					continue
				}
				// See if the names match.
				for _, tag := range instance.Tags {
					if tag != nil && tag.Key != nil && *tag.Key == "Name" && tag.Value != nil && *tag.Value == i.Name() {
						i.set(instance)
						log.Debug("Loaded %s, %s, %s -- via name", i.Name(), i.Id(), i.State())
						if err := i.loadVolumes(); err != nil {
							return err
						}
						return nil
					}
				}
			}
		}
		if resp.NextToken != nil {
			token = resp.NextToken
		} else {
			done = true
		}
	}
	log.Debug("Didn't find %s, skipping...", i.Name())
	return nil
}

func (i *instance) loadVolumes() error {
	for _, v := range i.volumes {
		if err := v.Load(); err != nil {
			return err
		}
		log.Debug("Volume %q loaded, %s", v.Id(), v.State())
	}
	return nil
}

func (i *instance) reload(test func() bool, m string) bool {
	// Clear the cached value
	log.Debug("Clearing cached value for %s", i.Name())
	i.cached = false
	for _, v := range i.volumes {
		log.Debug("Clearing cached value for volume %s", v.Id())
		v.cached = false
	}

	return msg.Wait(
		fmt.Sprintf("Waiting for Instance %s, %s to %s", i.Name(), i.Id(), m), //title
		fmt.Sprintf("Instance %s, %s failed to %s", i.Name(), i.Id(), m),      // err
		300,  // duration
		test, // test()
		func() bool {
			if err := i.load(); err != nil {
				msg.Error(err.Error())
				return false
			}
			return true
		},
	)
}

func (i *instance) reloadStarted() bool {
	return i.reload(i.Started, "start")
}

func (i *instance) reloadStopped() bool {
	return i.reload(i.Stopped, "stop")
}

func (i *instance) create(req *route.Request) route.Response {
	if i.Created() {
		return route.OK
	}

	securityGroupIds := []*string{}
	for _, sg := range i.secgroups {
		securityGroupIds = append(securityGroupIds, aws.String(sg.Id()))
	}
	volumes := []*ec2.BlockDeviceMapping{}
	for _, v := range i.volumes {
		if req.Flag("preserve_volume") {
			if !v.Preserve() {
				volumes = append(volumes, v.volumeParams)
			}
		} else {
			volumes = append(volumes, v.volumeParams)
		}
	}

	params := &ec2.RunInstancesInput{
		BlockDeviceMappings: volumes,
		ImageId:             aws.String(i.ImageId()),
		MaxCount:            aws.Int64(1),
		MinCount:            aws.Int64(1),
		InstanceType:        aws.String(i.InstanceType()),
		KeyName:             aws.String(i.KeyName()),
		SubnetId:            aws.String(i.subnet.Id()),
		SecurityGroupIds:    securityGroupIds,
	}

	reservation, err := i.ec2.RunInstances(params)
	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if reservation == nil || len(reservation.Instances) != 1 {
		msg.Error("Failed to create instance %s", i.Name())
		return route.FAIL
	}
	i.set(reservation.Instances[0])
	if !i.reloadStarted() {
		return route.FAIL
	}
	return route.OK
}

func (i *instance) destroy(req *route.Request) route.Response {
	if i.Destroyed() {
		return route.OK
	}

	params := &ec2.TerminateInstancesInput{
		InstanceIds: []*string{
			aws.String(i.Id()),
		},
	}
	if _, err := i.ec2.TerminateInstances(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	i.clear()
	i.compute.instanceCache.remove(i)

	return route.OK
}

func (i *instance) start(req *route.Request) route.Response {
	if i.Destroyed() {
		return route.OK
	}
	params := &ec2.StartInstancesInput{
		InstanceIds: []*string{
			aws.String(i.Id()),
		},
	}
	if _, err := i.ec2.StartInstances(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if !i.reloadStarted() {
		return route.FAIL
	}
	return route.OK
}

func (i *instance) stop(req *route.Request) route.Response {
	if i.Destroyed() {
		return route.OK
	}
	params := &ec2.StopInstancesInput{
		InstanceIds: []*string{
			aws.String(i.Id()),
		},
	}
	if _, err := i.ec2.StopInstances(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if !i.reloadStopped() {
		return route.FAIL
	}
	return route.OK

}

func (i *instance) restart(req *route.Request) route.Response {
	if i.Destroyed() {
		return route.OK
	}
	params := &ec2.RebootInstancesInput{
		InstanceIds: []*string{
			aws.String(i.Id()),
		},
	}
	if _, err := i.ec2.RebootInstances(params); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}
	if !i.reloadStarted() {
		return route.FAIL
	}
	return route.OK

}
