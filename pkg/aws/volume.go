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
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

// volume implements the resource.ProviderVolume interface.
type volume struct {
	*config.Volume
	ec2          *ec2.EC2
	compute      *compute
	volumeParams *ec2.BlockDeviceMapping

	instance *instance
	id       string

	volume *ec2.Volume
	cached bool
}

// newVolume is the constructor for a volume object. it returns a non-nil error upon failure.
func newVolume(comp resource.Compute, cfg *config.Volume, p *dataCenterProvider) (resource.ProviderVolume, error) {
	log.Info("Initializing AWS volume %q", cfg.Device())

	c, ok := comp.ProviderCompute().(*compute)
	if !ok {
		return nil, fmt.Errorf("AWS newVolume: Unable to obtain compute")
	}

	// Build the parameters that will be used by the instance
	params := &ec2.BlockDeviceMapping{
		DeviceName: aws.String(cfg.Device()),
		Ebs: &ec2.EbsBlockDevice{
			DeleteOnTermination: aws.Bool(!cfg.Keep()),
			VolumeSize:          aws.Int64(cfg.Size()),
			VolumeType:          aws.String(cfg.Type()),
		},
	}
	if !cfg.Boot() {
		params.Ebs.Encrypted = aws.Bool(true)
	}

	v := &volume{
		Volume:       cfg,
		ec2:          p.ec2,
		compute:      c,
		volumeParams: params,
	}
	return v, nil
}

func (v *volume) Route(req *route.Request) route.Response {
	return route.FAIL
}

func (v *volume) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		return fmt.Errorf("No flag set to find audit object")
	}
	a := aaa.AuditBuffer[flags[0]]
	if a == nil {
		return fmt.Errorf("Audit object does not exist")
	}
	// Configured but not Deployed
	if v.Destroyed() && v.instance != nil {
		a.Audit(aaa.Mismatched, "Instance %q | The volume %s, %s is configured, but not deployed", v.instance.Name(), v.Device(), v.MountPoint())
		// Have the function return early if there aren't volumes deployed to compare with
		return nil
	}
	// Mismatches
	// Encrypted Volumes?
	if v.volume.Encrypted != nil && !*v.volume.Encrypted && !v.Boot() {
		a.Audit(aaa.Mismatched, "Instance %q | Deployed Volume %q is not encrypted", v.instance.Name(), *v.volume.VolumeId)
	}
	// Correct Size
	if v.volume.Size != nil && *v.volume.Size != v.Size() {
		a.Audit(aaa.Mismatched, "Instance %q | Configured Volume Size: %d - Deployed Volume Size: %d", v.instance.Name(), v.Size(), v.volume.Size)
	}
	return nil
}

func (v *volume) associateInstance(i *instance) {
	v.instance = i
}

func (v *volume) associateVolume(volume *ec2.InstanceBlockDeviceMapping) {
	v.id = *volume.Ebs.VolumeId
	log.Debug("Associating volume %q", v.Id())
	v.set(v.compute.volumeCache.find(v))
	if v.volume != nil {
		log.Debug("Using cached volume %s", v.Id())
		v.cached = true
	}
}

func (v *volume) set(volume *ec2.Volume) {
	v.volume = volume
}

func (v *volume) clear() {
	v.volume = nil
	v.id = ""
}

func (v *volume) Load() error {
	// Use a cached value if it exists.
	if v.cached {
		log.Debug("Skipping volume load, cached...")
		return nil
	}

	// A cached value doesn't exist, so lets ask AWS.
	if v.instance == nil || v.Id() == "" {
		log.Debug("Volume is not associated with an instance, skipping load.")
		return nil
	}
	params := &ec2.DescribeVolumesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("volume-id"),
				Values: []*string{aws.String(v.Id())},
			},
		},
		VolumeIds: []*string{aws.String(v.Id())},
	}
	resp, err := v.ec2.DescribeVolumes(params)
	if err != nil {
		return err
	}
	if resp == nil || resp.Volumes == nil || len(resp.Volumes) != 1 || v.Id() != *resp.Volumes[0].VolumeId {
		if resp != nil {
			log.Error("Unusable DescribeVolumesOutput: %+v", resp)
		}
		return fmt.Errorf("DescribeVolumes returned unusable result")
	}
	v.set(resp.Volumes[0])
	return nil
}

func (v *volume) reload(test func() bool, m string) bool {
	// Clear the cached value
	log.Debug("Clearing cached value for %s", v.Id())
	v.cached = false

	return msg.Wait(
		fmt.Sprintf("Waiting for Volume %q to %s", v.Device(), m), //title
		fmt.Sprintf("Volume %q failed to %s", v.Id(), m),          //err
		300,  // duration
		test, // test()
		func() bool { // load()
			if err := v.Load(); err != nil {
				msg.Error(err.Error())
				return false
			}
			return true
		},
	)
}

func (v *volume) Created() bool {
	return v.volume != nil
}

func (v *volume) Destroyed() bool {
	return v.volume == nil
}

func (v *volume) Id() string {
	return v.id
}

func (v *volume) State() string {
	if !v.Created() {
		return ""
	}
	return *v.volume.State
}

// Attached returns true if the volume is attached to an instance.
func (v *volume) Attached() bool {
	if v.State() == "in-use" {
		return true
	}
	return false
}

// Detached returns true if the volume is detached from an instance.
func (v *volume) Detached() bool {
	if v.State() == "available" {
		return true
	}
	return false
}

// Detach this volume from the associated instance.
func (v *volume) Detach() error {
	if v.Destroyed() {
		return nil
	}
	log.Debug("Detaching volume %q from %q", v.Device(), v.MountPoint())
	params := &ec2.DetachVolumeInput{
		Device:     aws.String(v.Device()),
		InstanceId: aws.String(v.instance.Id()),
		VolumeId:   aws.String(v.Id()),
	}
	_, err := v.ec2.DetachVolume(params)
	if err != nil {
		return err
	}
	if !v.reload(v.Detached, "detach") {
		return fmt.Errorf("Unable to detach volume %q", v.Id())
	}
	log.Debug("Volume detached from %q with instance %q", v.Id(), v.instance.Id())
	aaa.Accounting("Volume detached: %s, %s, %s", v.Device(), v.MountPoint(), v.Id())
	return nil
}

// Attach this volume to the associated instance.
func (v *volume) Attach() error {
	if v.Destroyed() {
		return nil
	}
	log.Debug("Attaching volume %q to %q", v.Device(), v.MountPoint())
	params := &ec2.AttachVolumeInput{
		Device:     aws.String(v.Device()),
		InstanceId: aws.String(v.instance.Id()),
		VolumeId:   aws.String(v.Id()),
	}
	_, err := v.ec2.AttachVolume(params)
	if err != nil {
		return err
	}
	if !v.reload(v.Attached, "attach") {
		return fmt.Errorf("Unable to detach volume %q", v.Id())
	}
	log.Debug("Volume attached to volume %q with instance %q", v.Id(), v.instance.Id())
	aaa.Accounting("Volume attached: %s, %s, %s", v.Device(), v.MountPoint(), v.Id())
	return nil
}

// Destroy this volume. Detach must be called prior to this otherwise this will
// return an error.
func (v *volume) Destroy() error {
	if v.Destroyed() {
		return nil
	}
	log.Debug("Destroying volume %q, %q", v.Device(), v.MountPoint())
	params := &ec2.DeleteVolumeInput{
		VolumeId: aws.String(v.Id()),
	}
	if _, err := v.ec2.DeleteVolume(params); err != nil {
		return err
	}
	v.clear()
	aaa.Accounting("Volume destroyed: %s%s, %s", v.Device(), v.MountPoint(), v.Id())
	v.compute.volumeCache.remove(v)

	return nil
}

func (v *volume) Reset() {
	v.clear()
}

func (v *volume) SetTags(t map[string]string) error {
	return setTags(v.ec2, t, v.Id())
}

func (v *volume) Info() {
	if v.Destroyed() {
		return
	}
	msg.Info("Volume")
	msg.Detail("%-20s\t%s", "Id", v.Id())
	msg.Detail("%-20s\t%s", "State", v.State())
	msg.Detail("%-20s\t%s", "Instance Id", v.instance.Id())
	msg.Detail("%-20s\t%t", "Encrypted", *v.volume.Encrypted)
	msg.Detail("%-20s\t%d", "Size", *v.volume.Size)
	msg.Detail("%-20s\t%s", "Type", *v.volume.VolumeType)
	printTags(v.volume.Tags)
}
