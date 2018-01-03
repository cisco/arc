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

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cisco/arc/pkg/msg"
)

// Pods is a collection of Pod objects.
type Pods []*Pod

// Print provides a user friendly way to view the pods configuration.
func (p *Pods) Print() {
	msg.Info("Pods Config")
	msg.IndentInc()
	for _, pod := range *p {
		pod.Print()
	}
	msg.IndentDec()
}

// The configuration of the pod object. It contains a name, a servertype,
// the version of the servertype, the base image, the machine type, the associated
// subnet group, the associated security groups, the count being the number of instances
// created, and the list of volume templates to use for each instance.
type Pod struct {
	Name_           string   `json:"pod"`
	ServerType_     string   `json:"servertype"`
	Version_        int      `json:"version"`
	Image_          string   `json:"image"`
	InstanceType_   string   `json:"type"`
	Role_           string   `json:"role"`
	SubnetGroup_    string   `json:"subnet_group"`
	SecurityGroups_ []string `json:"security_groups"`
	Count_          int      `json:"count"`
	Teams_          []string `json:"teams"`
	Volumes         *Volumes `json:"volumes"`
	Instances       *Instances
}

// Name satisfies the resource.StaticPod interface. Pod names must be unique.
func (p *Pod) Name() string {
	return p.Name_
}

// ServerType satisfies the resource.StaticPod interface.
func (p *Pod) ServerType() string {
	return p.ServerType_
}

// Version satisfies the resource.StaticPod interface. This is the version of the
// servertype package used to provision each instance in this pod.
func (p *Pod) Version() string {
	return strconv.Itoa(p.Version_)
}

// PackageName satisfies the resource.StaticPod interface. This is a shortcut method
// provided to give the servertype package name.
func (p *Pod) PackageName() string {
	n := ""
	switch {
	case strings.HasPrefix(p.Image(), "centos"):
		n = fmt.Sprintf("servertype-%s-1.0.0-%s.x86_64.rpm", p.ServerType(), p.Version())
	case strings.HasPrefix(p.Image(), "ucxn"):
		n = fmt.Sprintf("servertype-%s-1.0.0-%s.x86_64.rpm", p.ServerType(), p.Version())
	case strings.HasPrefix(p.Image(), "qualys"):
		n = fmt.Sprintf("servertype-%s-1.0.0-%s.x86_64.rpm", p.ServerType(), p.Version())
	case strings.HasPrefix(p.Image(), "ubuntu"):
		n = fmt.Sprintf("servertype-%s_1.0.0-%s_amd64.deb", p.ServerType(), p.Version())
	}
	return n
}

// Image satisfies the resource.StaticPod interface. This returns the image used for
// each instance. The image is the base OS image (centos6, centos7, ubuntu, etc)
func (p *Pod) Image() string {
	return p.Image_
}

// InstanceType satisfies the resource.StaticPod interface. The instance type specifies
// the cloud providers machine type which defines the virtual cpus, memory and disk available
// to the instance. For example, in AWS m4.large is an instance type.
func (p *Pod) InstanceType() string {
	return p.InstanceType_
}

// Role satisfies the resource.StaticPod interface. The role is optional and allows
// instances in this pod to acquire an IAM role in order to interact with AWS
// programmatically.
func (p *Pod) Role() string {
	return p.Role_
}

// SubnetGroupName satisfies the resource.StaticPod interface. All instances in the pod
// will be placed on the associated subnet.
func (p *Pod) SubnetGroup() string {
	return p.SubnetGroup_
}

// SecurityGroups satisifies the resource.StaticPod interface. All instances in the pod
// will be associated with this list of security groups.
func (p *Pod) SecurityGroups() []string {
	return p.SecurityGroups_
}

// Count satisfies the resource.StaticPod interface. The pod will have count number of instances.
func (p *Pod) Count() int {
	return p.Count_
}

// Teams satisfies the resource.StaticPod interface. The pod will have the users in the given teams setup.
func (p *Pod) Teams() []string {
	return p.Teams_
}

// PrintLocal provides a user friendly way to view the configuration local to the pod object.
func (p *Pod) PrintLocal() {
	msg.Detail("%-20s\t%s", "name", p.Name())
	msg.Detail("%-20s\t%s", "servertype", p.ServerType())
	msg.Detail("%-20s\t%s", "version", p.Version())
	msg.Detail("%-20s\t%s", "image", p.Image())
	msg.Detail("%-20s\t%s", "type", p.InstanceType())
	msg.Detail("%-20s\t%s", "subnet_group", p.SubnetGroup())
}

// Print provides a user friendly way to view a pod configuration.
func (p *Pod) Print() {
	msg.Info("Pod Config")
	p.PrintLocal()
	groups, sep := "", ""
	for _, group := range p.SecurityGroups() {
		groups += sep + group
		sep = ", "
	}
	msg.Detail("%-20s\t%s", "security_groups", groups)
	teams, sep := "", ""
	for _, team := range p.Teams() {
		teams += sep + team
		sep = ", "
	}
	msg.Detail("%-20s\t%s", "teams", teams)
	msg.Detail("%-20s\t%d", "count", p.Count())
	msg.IndentInc()
	if p.Volumes != nil {
		p.Volumes.Print()
	}
	msg.IndentDec()
}
