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

package config

import "github.com/cisco/arc/pkg/msg"

// Instances is a collection of Instance objects.
type Instances []*Instance

// Print provides a user friendly way to view the instances configuration.
func (i *Instances) Print() {
	msg.Info("Instances Config")
	msg.IndentInc()
	for _, instance := range *i {
		instance.Print()
	}
	msg.IndentDec()
}

// The configuration of the instance object. An instance isn't part of the
// actual config file, but it is provided for consistency. The instance
// contains a name, and a pointer to the pod config.
type Instance struct {
	*Pod
	name string
}

// NewInstance creates a new instance config object. It requires the
// name of the instance and a reference to the pod containing the instance.
func NewInstance(name string, p *Pod) *Instance {
	return &Instance{
		Pod:  p,
		name: name,
	}
}

// Name satisfies the resource.StaticInstance interface. Instance names must be unique.
func (i *Instance) Name() string {
	return i.name
}

// ServerType satisfies the resource.StaticInstance interface.
func (i *Instance) ServerType() string {
	return i.Pod.ServerType()
}

// Version satisfies the resource.StaticInterface interface. This is the version of the
// servertype package used to provision this instance.
func (i *Instance) Version() string {
	return i.Pod.Version()
}

// Image satisfies the resource.StaticInstance interface. This returns the image used for
// this instance. The image is the base OS image (centos6, centos7, ubuntu14, etc)
func (i *Instance) Image() string {
	return i.Pod.Image()
}

// InstanceType satisfies the resource.StaticInstance interface. The instance type specifies
// the cloud providers machine type which defines the virtual cpus, memory and disk available
// to this instance. For example, in AWS m4.large is an instance type.
func (i *Instance) InstanceType() string {
	return i.Pod.InstanceType()
}

// Role satisfies the resource.StaticInstance interface. The role is optional and allows
// instances in this pod to acquire an IAM role in order to interact with AWS
// programmatically.
func (i *Instance) Role() string {
	return i.Pod.Role()
}

// SubnetGroupName satisfies the resource.StaticInterface interface. This instance
// will be placed on the associated subnet.
func (i *Instance) SubnetGroupName() string {
	return i.Pod.SubnetGroup()
}

// SecurityGroupNames satisifies the resource.StaticInstance interface. This instance
// will be associated with this list of security groups.
func (i *Instance) SecurityGroupNames() []string {
	return i.Pod.SecurityGroups()
}

// Teams satisfies the resource.StaticInstance interface. The instance will have the users in the given teams setup.
func (i *Instance) Teams() []string {
	return i.Pod.Teams()
}

func (i *Instance) Volumes() *Volumes {
	return i.Pod.Volumes
}

// PrintLocal provides a user friendly way to view the configuration local to the instance object.
func (i *Instance) PrintLocal() {
	msg.Info("Instance Config")
	msg.Detail("%-20s\t%s", "name", i.Name())
	msg.Detail("%-20s\t%s", "servertype", i.ServerType())
	msg.Detail("%-20s\t%s", "version", i.Version())
	msg.Detail("%-20s\t%s", "image", i.Image())
	msg.Detail("%-20s\t%s", "type", i.InstanceType())
	msg.Detail("%-20s\t%s", "subnet_group", i.SubnetGroupName())
	groups, sep := "", ""
	for _, group := range i.SecurityGroupNames() {
		groups += sep + group
		sep = ", "
	}
	msg.Detail("%-20s\t%s", "security_groups", groups)
	teams, sep := "", ""
	for _, team := range i.Teams() {
		teams += sep + team
		sep = ", "
	}
	msg.Detail("%-20s\t%s", "teams", teams)
}

// Print provides a user friendly way to view a instance configuration.
func (i *Instance) Print() {
	i.PrintLocal()
	msg.IndentInc()
	if i.Volumes() != nil {
		i.Volumes().Print()
	}
	msg.IndentDec()
}
