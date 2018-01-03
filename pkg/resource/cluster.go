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

import (
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/route"
)

// StaticCluster provides the interface to the static portion of the
// cluster. This information is provided via config file and is implemented
// by config.Cluster.
type StaticCluster interface {
	Name() string
	SecurityTags() config.SecurityTags
	AuditIgnore() bool
}

// Cluster provides the resource interface used for the common cluster
// object implemented in the arc package.
type Cluster interface {
	Resource
	StaticCluster
	Auditor

	// Compute provides access to Cluster's parent.
	Compute() Compute

	// Pods provides access to Compute's child pods.
	Pods() Pods

	// Find pod by name. This implies pods are named uniquely.
	FindPod(name string) Pod

	// Find instance by name. This implies instances are named uniquely.
	// The name takes the form "<pod name>-<instance number>".
	FindInstance(name string) Instance

	// Find the instance by ip address.
	FindInstanceByIP(ip string) Instance

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
