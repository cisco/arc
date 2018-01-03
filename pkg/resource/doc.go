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

/*
Package resource provides the interfaces to the resources common to cloud
providers such as OpenStack, Amazon's AWS, Google's GCP, and Microsoft's
Azure. The orgranization of these resources is opinionated such that it
matches the needs of the Spark Call project.

Arc

The arc resource tree begins with the Arc resource as it's root, it contains
a datacenter resource and a dns resource.

The following files implement Arc.

	arc.Arc
		Implements the resource.Arc interface.
		Embeds resource.Resources since it is a collection of the datacenter and dns resources.
		Embeds config.Arc which provides the StaticArc interface.

	config.Arc
 		Implements the resource.StaticArc interface.
		The arc configuration file is unmarshalled into config.Arc

DataCenter

The purpose of the datacenter is to collect and manage the resources necessary
to standup instances to do real work. The datacenter resource contains network
resources and compute resources. It also has it's own provider separate from
the dns provider.

Network

The network resource contains the resources necessary to create a network
for the datacenter. TODO

SubnetGroups, SubnetGroup and Subnet

The subnet groups are a concept specific to spark call. A subnet group is a
collection of related subnets whose purpose is to create an extended  subnet
across availability zones in order to provide highly available instances.

For example if a network has three availability zones, and a subnet group name
"public" was defined, such as

	"network": {
		"cidr": "192.168.0.0/16",
		"availability_zones": [ "az1", "az2", "az3" ],
		"subnet_groups": [
		{
			"name":   "public",
			"cidr":   "192.168.10.0/24",
			"access": "public"
		},

The "public" subnet group would contain three subnets, each one located in a separate
availability zone, one subnet with the starting cidr of 192.168.10.0, the next with
192.168.10.1 and the next with 192.168.10.2.

Subnet groups are also intended to be the destionation of a security rule. This means
that a security group can specify "subnet:public" and a three security rules will be
created, one for each subnet.

SecurityGroups, SecurityGroup, SecurityRules and SecurityRule

A security group is a collection of security rules which allow ingress and egress
network traffic to flow to given destination.

Compute

TODO.

KeyPair

TODO.

Clusters and Cluster

TODO.

Pods and Pod

TODO.

Instances and Instance

TODO.

Dns

TODO.

DnsRecords and DnsRecord

TODO.

*/
package resource
