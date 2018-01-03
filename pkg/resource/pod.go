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

package resource

import "github.com/cisco/arc/pkg/route"

// StaticPod provides the interface to the static portion of the
// pod. This information is provided via config file and is implemented
// by config.Pod.
type StaticPod interface {
	Name() string
	ServerType() string
	Version() string
	PackageName() string
	Image() string
	InstanceType() string
	Role() string
	SubnetGroup() string
	SecurityGroups() []string
	Teams() []string
	Count() int
}

// Pod provides the resource interface used for the common pod
// object implemented in the arc package.
type Pod interface {
	Resource
	StaticPod
	Auditor

	// Cluster provides access to Pod's parent.
	Cluster() Cluster

	// Instances provides access to Pod's child instances.
	Instances() Instances

	// Find instance by name. This implies instances are named uniquely.
	// The name takes the form "<pod name>-<instance number>".
	FindInstance(name string) Instance

	// Find the instance by ip address.
	FindInstanceByIP(ip string) Instance

	// Dervied returns the base pod.
	Derived() Pod

	// DnsCNameRecords returns the configred dns cname records associated with this pod.
	// If this pod doesn't have configured cname records, this will return nil.
	DnsCNameRecords() []DnsRecord

	// PrimaryCname returns the primaru cname record for this pod.
	// If this pod doesn't have configured cname records, this will return nil.
	PrimaryCname() DnsRecord

	// PrimaryInstance returns the instance associated with the cname record for this pod.
	// If this pod doesn't have a configured cname record or cannot find a primary instance
	// this will return nil.
	PrimaryInstance() Instance

	// SecondaryInstances return the instances not associated with the cname reocrd for this pod.
	// If this pod doesn't have a configured cname record, cannot find a primary instance
	// or has a single instance this will return nil.
	SecondaryInstances() []Instance

	// PkgName returns the name of the servertype rpm or deb associated with this pod.
	PkgName() string

	// Creator
	PreCreate(req *route.Request) route.Response
	Create(req *route.Request) route.Response
	PostCreate(req *route.Request) route.Response

	// Destroyer
	PreDestroy(req *route.Request) route.Response
	Destroy(req *route.Request) route.Response
	PostDestroy(req *route.Request) route.Response

	// Provisioner
	PreProvision(req *route.Request) route.Response
	Provision(req *route.Request) route.Response
	PostProvision(req *route.Request) route.Response

	// Starter
	PreStart(req *route.Request) route.Response
	Start(req *route.Request) route.Response
	PostStart(req *route.Request) route.Response

	// Stopper
	PreStop(req *route.Request) route.Response
	Stop(req *route.Request) route.Response
	PostStop(req *route.Request) route.Response

	// Restarter
	PreRestart(req *route.Request) route.Response
	Restart(req *route.Request) route.Response
	PostRestart(req *route.Request) route.Response

	// Replacer
	PreReplace(req *route.Request) route.Response
	Replace(req *route.Request) route.Response
	PostReplace(req *route.Request) route.Response

	// Auditor
	AuditOverride
}
