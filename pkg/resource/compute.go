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

// StaticCompuite provides the interface to the static portion of the
// compute object. This information is provided via config file and is implemented
// by config.Compute.
type StaticCompute interface {
	Name() string
	BootstrapVersion() int
	DeployVersion() int
	SecretsVersion() int
	AideVersion() int
}

// DyanmicCompute provides the interface to the dynamic portion of compute.
type DynamicCompute interface {

	// AuditVolumes identifies any volumes that have been deployed but are not in the configuration.
	AuditVolumes(flags ...string) error

	// AuditEIP identifies any elastic IPs that have been allocated but are associated with anything.
	AuditEIP(flags ...string) error

	// AuditInstance identifies any instances that have been deployed but are not in the configuration.
	AuditInstances(flags ...string) error
}

// Compute provides the resource interface used for the common compute
// object implemented in the arc package. It contains a DataCenter method
// to access it's parent, and the KeyPair and Clusters methods to access
// it's children.
type Compute interface {
	Resource
	StaticCompute
	DynamicCompute

	// DataCenter provides access to Compute's parent.
	DataCenter() DataCenter

	// KeyPair provides access to Compute's child keypair.
	KeyPair() KeyPair

	// Clusters provides access to Compute's child clusters.
	Clusters() Clusters

	// Find cluster by name. This implies clusters are named uniquely.
	FindCluster(name string) Cluster

	// Find pod by name. This implies pods are named uniquely.
	FindPod(name string) Pod

	// Find instance by name. This implies instances are named uniquely.
	// The name takes the form "<pod name>-<instance number>".
	FindInstance(name string) Instance

	// Find the instance by ip address.
	FindInstanceByIP(ip string) Instance

	// ProviderCompute provides access to the provider specific compute.
	ProviderCompute() ProviderCompute
}

// ProviderCompute provides a resource interface for the provider supplied compute.
type ProviderCompute interface {
	DynamicCompute
}
