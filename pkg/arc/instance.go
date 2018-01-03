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

package arc

import (
	"fmt"
	"strings"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type InstanceFactory func(
	resource.Pod, resource.Subnet, resource.KeyPair,
	provider.DataCenter, *config.Instance,
) (resource.Instance, error)

var instanceFactories map[string]InstanceFactory

func RegisterInstanceFactory(name string, f InstanceFactory) {
	instanceFactories[name] = f
}

func init() {
	instanceFactories = map[string]InstanceFactory{}
}

type Instance struct {
	*config.Instance
	pod              resource.Pod
	net              resource.Network
	subnet           resource.Subnet
	secgroups        []resource.SecurityGroup
	keypair          resource.KeyPair
	role             *role
	volumes          *volumes
	eip              *elasticIP
	dns              *dns
	providerInstance resource.ProviderInstance
	privateARecord   *dnsRecord
	publicARecord    *dnsRecord
	derived_         resource.Instance
}

// newInstance is the constructor for a instance object. It returns a non-nil error upon failure.
func newInstance(
	pod resource.Pod, subnet resource.Subnet, keypair resource.KeyPair,
	prov provider.DataCenter, cfg *config.Instance,
) (resource.Instance, error) {
	log.Debug("Initializing Instance %q", cfg.Name())

	factory := instanceFactories[cfg.ServerType()]
	if factory != nil {
		return factory(pod, subnet, keypair, prov, cfg)
	}
	return NewDefaultInstance(pod, subnet, keypair, prov, cfg)
}

// newDefaultInstance is the default instance contructor.
func NewDefaultInstance(
	pod resource.Pod, subnet resource.Subnet, keypair resource.KeyPair,
	prov provider.DataCenter, cfg *config.Instance,
) (*Instance, error) {

	// The references to the network and compute resources.
	net := pod.Cluster().Compute().DataCenter().Network()
	compute := pod.Cluster().Compute()

	// The list of references to it's security group resources.
	secgroups := []resource.SecurityGroup{}
	for _, secgroupName := range cfg.SecurityGroupNames() {
		secgroup := net.SecurityGroups().Find(secgroupName)
		if secgroup == nil {
			return nil, fmt.Errorf("newInstance, unknown secgroup name %s", secgroupName)
		}
		secgroups = append(secgroups, secgroup)
	}

	i := &Instance{
		Instance:  cfg,
		pod:       pod,
		net:       net,
		subnet:    subnet,
		secgroups: secgroups,
		keypair:   keypair,
	}

	var err error
	i.role, err = newRole(i, prov, cfg.Role())
	if err != nil {
		return nil, err
	}
	i.volumes, err = newVolumes(compute, prov, cfg.Volumes())
	if err != nil {
		return nil, err
	}

	// Create underlying provider implementation of the instance.
	i.providerInstance, err = prov.NewInstance(i, cfg)
	if err != nil {
		return nil, err
	}

	if i.subnet.Access() == "public_elastic" {
		var err error
		i.eip, err = newElasticIP(i, prov)
		if err != nil {
			return nil, err
		}
	}

	// The dns records need to be allocated during the load step, since
	// the dns subsystem is allocated after the datacenter subsystem is
	// allocated. This function is in the context of the datacenter
	// subsystem.

	return i, nil
}

// Created returns true is the instance has been created in the provider.
// This satisfies the resource.Resource interface.
func (i *Instance) Created() bool {
	return i.providerInstance.Created()
}

// Destroyed returns true is the instance has not been created in the provider.
// This satisfies the resource.Resource interface.
func (i *Instance) Destroyed() bool {
	return i.providerInstance.Destroyed()
}

// Pod provides access to Instance's parent. Pod satisfies the resource.Instance interface.
func (i *Instance) Pod() resource.Pod {
	return i.pod
}

// Network provides access to the network to which instance is associated.
func (i *Instance) Network() resource.Network {
	return i.net
}

// Subnet provides access to the subnet to which instance is allocated.
func (i *Instance) Subnet() resource.Subnet {
	return i.subnet
}

// SecurityGroups provides access to the security groups to which instance is associated.
func (i *Instance) SecurityGroups() []resource.SecurityGroup {
	return i.secgroups
}

// KeyPair provides access to the keypair that will be assigned to this instance.
func (i *Instance) KeyPair() resource.KeyPair {
	return i.keypair
}

// Dns provides access to the dns associated with the datacenter.
func (i *Instance) Dns() resource.Dns {
	return i.dns
}

// PrivateHostname returns the hostname (without the domain name) associated with the
// private ip address of the instance.
func (i *Instance) PrivateHostname() string {
	return i.Name() + "-internal"
}

// PrivateFQDN returns the FQDN associated with the private ip address of the instance.
func (i *Instance) PrivateFQDN() string {
	return i.PrivateHostname() + "." + i.dns.Domain()
}

// PublicHostname returns the hostname (without the domain name) associated with the
// public ip address of the instance.
func (i *Instance) PublicHostname() string {
	switch i.subnet.Access() {
	case "public":
		return i.Name()
	case "public_elastic":
		return i.Name()
	}
	return ""
}

// PublicFQDN returns the FQDN associated with the public ip address of the instance.
// If this instance doesn't have a public ip address (most don't) this will return
// an empty string "".
func (i *Instance) PublicFQDN() string {
	switch i.subnet.Access() {
	case "public", "public_elastic":
		return i.PublicHostname() + "." + i.dns.Domain()
	}
	return ""
}

// PrivateDnsARecord returns the dns a record associated with this instance.
func (i *Instance) PrivateDnsARecord() resource.DnsRecord {
	return i.privateARecord
}

// PublicDnsARecord return the dns a record associated with the public ip address of this
// instance. If this instance doesn't have a public ip address (most don't) this will
// return nil.
func (i *Instance) PublicDnsARecord() resource.DnsRecord {
	return i.publicARecord
}

// FQDNMatch returns true if the given list of fqdns contains a match to either the
// private or public fqdn.
func (i *Instance) FQDNMatch(fqdns []string) bool {
	for _, fqdn := range fqdns {
		// SCPLATFORM-4668: Normalize the fqdn, private fqdn and public fqdn so we can compare them correctly.
		if !strings.HasSuffix(fqdn, ".") {
			fqdn += "."
		}
		privateFqdn := i.PrivateFQDN()
		if !strings.HasSuffix(privateFqdn, ".") {
			privateFqdn += "."
		}
		publicFqdn := i.PublicFQDN()
		if !strings.HasSuffix(publicFqdn, ".") {
			publicFqdn += "."
		}
		if fqdn == privateFqdn || fqdn == publicFqdn {
			return true
		}
	}
	return false
}

func (i *Instance) Id() string {
	if i.providerInstance == nil {
		return ""
	}
	return i.providerInstance.Id()
}

func (i *Instance) ImageId() string {
	if i.providerInstance == nil {
		return ""
	}
	return i.providerInstance.ImageId()
}

func (i *Instance) KeyName() string {
	if i.providerInstance == nil {
		return ""
	}
	return i.providerInstance.KeyName()
}

func (i *Instance) State() string {
	if i.providerInstance == nil {
		return ""
	}
	return i.providerInstance.State()
}

func (i *Instance) Started() bool {
	if i.providerInstance == nil {
		return false
	}
	return i.providerInstance.Started()
}

func (i *Instance) Stopped() bool {
	if i.providerInstance == nil {
		return false
	}
	return i.providerInstance.Stopped()
}

func (i *Instance) PrivateIPAddress() string {
	if i.providerInstance == nil {
		return ""
	}
	return i.providerInstance.PrivateIPAddress()
}

func (i *Instance) PublicIPAddress() string {
	if i.providerInstance == nil {
		return ""
	}
	if i.eip != nil && i.eip.Id() != "" {
		return i.eip.IpAddress()
	}
	return i.providerInstance.PublicIPAddress()
}

func (i *Instance) SetTags(t map[string]string) error {
	if i.providerInstance == nil {
		return fmt.Errorf("providerInstance not created")
	}
	return i.providerInstance.SetTags(t)
}

func (i *Instance) RootUser() string {
	switch {
	case strings.HasPrefix(i.Image(), "centos"):
		return "centos"
	case strings.HasPrefix(i.Image(), "ucxn"):
		return "root"
	case strings.HasPrefix(i.Image(), "ubuntu"):
		return "ubuntu"
	}
	return "root"
}

func (i *Instance) ProviderVolumes() []resource.ProviderVolume {
	pv := []resource.ProviderVolume{}
	for _, r := range i.volumes.Get() {
		v := r.(resource.Volume)
		pv = append(pv, v.ProviderVolume())
	}
	return pv
}

func (i *Instance) ProviderRole() resource.ProviderRole {
	r := i.role.ProviderRole()
	return r
}

func (i *Instance) Derived() resource.Instance {
	if i.derived_ != nil {
		return i.derived_
	}
	return i
}

func (i *Instance) SetDerived(d resource.Instance) {
	i.derived_ = d
}

func (i *Instance) Route(req *route.Request) route.Response {
	log.Route(req, "Instance %q", i.Name())

	if req.Top() != "" {
		i.help()
		return route.FAIL
	}

	// Skip if the test flag is set.
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load:
		return i.load(req)
	case route.Help:
		i.help()
		return route.OK
	case route.Config:
		i.config()
		return route.OK
	case route.Info:
		i.info()
		return route.OK
	case route.Create:
		msg.Info("Instance Creation: %s", i.Name())
		if i.Created() {
			msg.Detail("Instance exists, skipping...")
			return route.OK
		}
		if resp := i.create(req); resp != route.OK {
			return resp
		}
		if req.Flag("noprovision") {
			return route.OK
		}
		req.Flags().Append("initial")
		resp := i.provision(req)
		req.Flags().Remove("initial")
		return resp
	case route.Destroy:
		// See instance_destroy.go
		return i.destroy(req)
	case route.Provision:
		// See instance_provision.go
		return i.provision(req)
	case route.Start:
		// See instance_restart.go
		return i.start(req)
	case route.Stop:
		// See instance_restart.go
		return i.stop(req)
	case route.Restart:
		// See instance_restart.go
		return i.restart(req)
	case route.Replace:
		// See instance_replace.go
		return i.replace(req)
	case route.Audit:
		// See instance_audit.go
		err := aaa.NewAudit("Instance")
		if err != nil {
			msg.Error(err.Error())
		}
		if err := i.Audit("Instance"); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
		return route.OK
	default:
		msg.Error("Unknown instance command %q.", req.Command().String())
	}
	return route.FAIL
}

func (i *Instance) load(req *route.Request) route.Response {
	// See instance_dns.go for instance dns implementation

	// Initialize dns, since the dns subsystem is created after the datacenter subsystem.
	if i.dns == nil {
		var ok bool
		i.dns, ok = i.Pod().Cluster().Compute().DataCenter().Dns().(*dns)
		if !ok {
			msg.Error("Instance load, failed to initialize dns")
			return route.FAIL
		}
	}

	// Load the instance
	if i.providerInstance.Route(req) != route.OK {
		return route.FAIL
	}

	// Create the dns a records
	if err := i.newDnsARecords(); err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	// Load the dns a records
	if i.Created() && i.loadDnsARecords(req) != route.OK {
		return route.FAIL
	}

	// Load the elastic IP
	if i.eip != nil {
		if err := i.eip.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	}

	// Load the role
	if i.role.Name() != "" {
		if err := i.role.Load(); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	}

	return route.OK
}

func (i *Instance) reload(req *route.Request, test func() bool, m string) bool {
	req.Flags().Append("reload")
	defer req.Flags().Remove("reload")
	return msg.Wait(
		fmt.Sprintf("Waiting for Instance %s, %s to %s", i.Name(), i.Id(), m), //title
		fmt.Sprintf("Instance %s, %s failed to %s", i.Name(), i.Id(), m),      // err
		300,  // duration
		test, // test()
		func() bool {
			return i.load(req) == route.OK
		},
	)
}

func (i *Instance) reloadStarted(req *route.Request) bool {
	return i.reload(req, i.Started, "start")
}

func (i *Instance) reloadStopped(req *route.Request) bool {
	return i.reload(req, i.Stopped, "stop")
}

func (i *Instance) help() {
	instanceHelp(i.Name())
}

func instanceHelp(n string) {
	name := ""
	if n != "" {
		name = " " + n
	}
	commands := []help.Command{
		{route.Create.String(), fmt.Sprintf("create%s instance", name)},
		{route.Provision.String(), fmt.Sprintf("provision%s instance", name)},
		{route.Provision.String() + " users", fmt.Sprintf("update%s instance users", name)},
		{route.Start.String(), fmt.Sprintf("start%s instance", name)},
		{route.Stop.String(), fmt.Sprintf("stop%s instance", name)},
		{route.Restart.String(), fmt.Sprintf("restart%s instance", name)},
		{route.Replace.String(), fmt.Sprintf("replace%s instance", name)},
		{route.Audit.String(), fmt.Sprintf("audit%s instance", name)},
		{route.Destroy.String(), fmt.Sprintf("destroy%s instance", name)},
		{route.Config.String(), fmt.Sprintf("provide the%s instance configuration", name)},
		{route.Info.String(), fmt.Sprintf("provide information about allocated%s instance", name)},
		{route.Help.String(), "provide this help"},
	}
	if name == "" {
		name = " [name]"
	}
	help.Print(fmt.Sprintf("instance%s", name), commands)
}

func (i *Instance) config() {
	i.Instance.Print()
}

func (i *Instance) info() {
	if i.Destroyed() {
		return
	}
	msg.Info("Instance")
	msg.Detail("%-20s\t%s", "name", i.Name())
	msg.Detail("%-20s\t%s", "id", i.Id())
	msg.Detail("%-20s\t%s", "image id", i.ImageId())
	msg.Detail("%-20s\t%s", "state", i.State())
	msg.Detail("%-20s\t%s", "private ip address", i.PrivateIPAddress())
	msg.Detail("%-20s\t%s", "private dns a record", i.privateARecord.Id())
	if i.PublicIPAddress() != "" {
		msg.Detail("%-20s\t%s", "public ip address", i.PublicIPAddress())
	}
	if i.publicARecord != nil {
		msg.Detail("%-20s\t%s", "public dns a record", i.publicARecord.Id())
	}
	if i.role.Name() != "" {
		msg.Detail("%-20s\t%s", "role", i.role.Name())
	}
	if i.volumes != nil {
		i.volumes.info()
	}
}
