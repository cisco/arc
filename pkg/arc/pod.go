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

type PodFactory func(resource.Cluster, provider.DataCenter, *config.Pod) (resource.Pod, error)

var podFactories map[string]PodFactory

func RegisterPodFactory(name string, f PodFactory) {
	podFactories[name] = f
}

func init() {
	podFactories = map[string]PodFactory{}
}

type Pod struct {
	*resource.Resources
	*config.Pod
	cluster      resource.Cluster
	instances    *instances
	cnameRecords []resource.DnsRecord
	primaryCName resource.DnsRecord
	derived_     resource.Pod
}

// newPod is the constructor for a pod object. It returns a non-nil error upon failure.
func newPod(cluster resource.Cluster, prov provider.DataCenter, cfg *config.Pod) (resource.Pod, error) {
	log.Debug("Initializing Pod %q", cfg.Name())

	factory := podFactories[cfg.ServerType()]
	if factory != nil {
		return factory(cluster, prov, cfg)
	}
	return NewDefaultPod(cluster, prov, cfg)
}

// newDefaultPod is the default pod contructor.
func NewDefaultPod(cluster resource.Cluster, prov provider.DataCenter, cfg *config.Pod) (*Pod, error) {
	p := &Pod{
		Resources: resource.NewResources(),
		Pod:       cfg,
		cluster:   cluster,
	}
	p.derived_ = p

	// Allocate a config.Instances structure since it isn't part of the config file.
	instancesConfig := config.Instances{}
	for i := 1; i <= cfg.Count(); i++ {
		name := fmt.Sprintf("%s-%02d", cfg.Name(), i)
		conf := config.NewInstance(name, cfg)
		instancesConfig = append(instancesConfig, conf)
	}
	cfg.Instances = &instancesConfig

	var err error
	p.instances, err = newInstances(p, prov, cfg.Instances)
	if err != nil {
		return nil, err
	}
	p.Append(p.instances)
	return p, nil
}

// Cluster provides access to Pod's parent. Cluster satisfies the resource.Pod interface.
func (p *Pod) Cluster() resource.Cluster {
	return p.cluster
}

// Instances provides access to Pod's child instances. Instances satisfies the resource.Pod interface.
func (p *Pod) Instances() resource.Instances {
	return p.instances
}

// FindInstance finds the instance in this pod by name. This implies instances are named uniquely.
// The name takes the form "<pod name>-<instance number>". Find instance satisfies the resource.Pod
// interface.
func (p *Pod) FindInstance(name string) resource.Instance {
	return p.instances.Find(name)
}

// FindInstanceByIP finds the instance in this pod by ip address.  Find instance satisfies the
// resource.Pod interface.
func (p *Pod) FindInstanceByIP(ip string) resource.Instance {
	return p.instances.FindByIP(ip)
}

// DnsCNameRecord returns the configred dns cname record associated with this pod.
// If this pod doesn't have configured cname record, this will return nil.
func (p *Pod) DnsCNameRecords() []resource.DnsRecord {
	return p.cnameRecords
}

func (p *Pod) primaryIndex() int {
	if p.primaryCName == nil {
		return -1
	}
	for n, j := range p.instances.Get() {
		i := j.(resource.Instance)
		if !i.Created() {
			continue
		}
		if i.FQDNMatch(p.primaryCName.DynamicValues()) {
			return n
		}
	}
	return -1
}

func (p *Pod) PrimaryCname() resource.DnsRecord {
	return p.primaryCName
}

// PrimaryInstance returns the instance associated with the cname record for this pod.
// If this pod doesn't have a configured cname record or cannot find a primary instance
// this will return nil.
func (p *Pod) PrimaryInstance() resource.Instance {
	if p.primaryCName == nil {
		return nil
	}
	n := p.primaryIndex()
	if n < 0 {
		return nil
	}
	return p.instances.Get()[n].(resource.Instance)
}

// SecondaryInstances return the instances not associated with the cname record for this pod.
// If this pod doesn't have a configured cname record or cannot find a primary instance
// this will return nil.
func (p *Pod) SecondaryInstances() []resource.Instance {
	if p.primaryCName == nil {
		return nil
	}
	n := p.primaryIndex()
	if n < 0 {
		return nil
	}
	l := len(p.instances.Get())
	if l < 2 {
		return nil
	}

	s := []resource.Instance{}

	for m := (n + 1) % l; m != n; m = (m + 1) % l {
		j := p.instances.Get()[m]
		i := j.(resource.Instance)
		if !i.Created() {
			continue
		}
		s = append(s, i)
	}
	if len(s) < 1 {
		return nil
	}
	return s
}

// PkgName returns the name of the servertype rpm or deb associated with this pod.
func (p *Pod) PkgName() string {
	switch {
	case strings.HasPrefix(p.Image(), "centos"):
		return fmt.Sprintf("servertype-%s-1.0.0-%s.x86_64.rpm", p.ServerType(), p.Version())
	case strings.HasPrefix(p.Image(), "ucxn"):
		return fmt.Sprintf("servertype-%s-1.0.0-%s.x86_64.rpm", p.ServerType(), p.Version())
	case strings.HasPrefix(p.Image(), "ubuntu"):
		return fmt.Sprintf("servertype-%s_1.0.0-%s_amd64.deb", p.ServerType(), p.Version())
	}
	return ""
}

func (p *Pod) Derived() resource.Pod {
	return p.derived_
}

func (p *Pod) SetDerived(d resource.Pod) {
	p.derived_ = d
}

// Route satisfies the embedded resource.Resource interface in resource.Pod.
// Pod handles load, create, destroy, start, stop, replace, help, config and
// info requests in order to manage the pod. All other commands are routed to
// the pod's instances.
func (p *Pod) Route(req *route.Request) route.Response {
	log.Route(req, "Pod %q", p.Name())

	// Route to the appropriate resource. Instance is the only valid sub-resource.
	if req.Top() == "instance" {
		req.Pop()
		if req.Top() == "" {
			instanceHelp("")
			return route.FAIL
		}
		instance := p.FindInstance(req.Top())
		if instance == nil {
			msg.Error("Unknown instance %q.", req.Top())
			return route.FAIL
		}
		return instance.Route(req.Pop())
	}

	// The path should be empty at this point.
	if req.Top() != "" {
		p.help()
		return route.FAIL
	}

	if err := aaa.Authorized(req, "pod", p.Name()); err != nil {
		msg.Error(err.Error())
		return route.UNAUTHORIZED
	}

	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load:
		return p.load(req)
	case route.Help:
		p.help()
		return route.OK
	case route.Config:
		p.config()
		return route.OK
	case route.Info:
		p.info(req)
		return route.OK
	case route.Create:
		return p.create(req)
	case route.Destroy:
		return p.destroy(req)
	case route.Provision:
		return p.provision(req)
	case route.Start:
		return p.start(req)
	case route.Stop:
		return p.stop(req)
	case route.Restart:
		return p.restart(req)
	case route.Replace:
		return p.replace(req)
	default:
		msg.Error("Unknown pod command %q.", req.Command().String())
	}
	return route.FAIL
}

func (p *Pod) routeToChildren(req *route.Request) route.Response {
	if !req.Flag("podonly") {
		return p.RouteInOrder(req)
	}
	return route.OK
}

func (p *Pod) routeReverseToChildren(req *route.Request) route.Response {
	if !req.Flag("podonly") {
		return p.RouteReverseOrder(req)
	}
	return route.OK
}

func (p *Pod) load(req *route.Request) route.Response {
	// Set the cname records here rather than in new since the dns subsystem
	// is allocated after the datacenter subsystem.
	dns := p.Cluster().Compute().DataCenter().Dns()
	if dns != nil {
		p.cnameRecords = dns.CNameRecords().FindByPod(p.Name())
		p.primaryCName = dns.CNameRecords().Find(p.Name())
	}
	return p.RouteInOrder(req)
}

func (p *Pod) help() {
	podHelp(p.Name())
}

func podHelp(n string) {
	name := ""
	if n != "" {
		name = " " + n
	}
	commands := []help.Command{
		{route.Create.String(), fmt.Sprintf("create%s pod", name)},
		{route.Provision.String(), fmt.Sprintf("provision%s pod", name)},
		{route.Provision.String() + " users", fmt.Sprintf("update%s pod users", name)},
		{route.Start.String(), fmt.Sprintf("start%s pod", name)},
		{route.Stop.String(), fmt.Sprintf("stop%s pod", name)},
		{route.Restart.String(), fmt.Sprintf("restart%s pod", name)},
		{route.Replace.String(), fmt.Sprintf("replace%s pod", name)},
		{route.Audit.String(), fmt.Sprintf("audit%s pod", name)},
		{route.Destroy.String(), fmt.Sprintf("destroy%s pod", name)},
		{route.Config.String(), fmt.Sprintf("provide the%s pod configuration", name)},
		{route.Info.String(), fmt.Sprintf("provide information about allocated%s pod", name)},
		{route.Help.String(), "provide this help"},
	}
	if name == "" {
		name = " [name]"
	}
	help.Print(fmt.Sprintf("pod%s", name), commands)
}

func (p *Pod) config() {
	p.Pod.Print()
}

func (p *Pod) info(req *route.Request) {
	if p.Destroyed() {
		return
	}
	msg.Info("Pod")
	msg.Detail("%-20s\t%s", "name", p.Name())
	msg.Detail("%-20s\t%s", "servertype", p.ServerType())
	msg.Detail("%-20s\t%s", "version", p.Version())
	msg.Detail("%-20s\t%s", "package name", p.PackageName())
	msg.Detail("%-20s\t%s", "image", p.Image())
	msg.Detail("%-20s\t%s", "type", p.InstanceType())
	securityGroups, sep := "", ""
	for _, securityGroup := range p.SecurityGroups() {
		securityGroups += sep + securityGroup
		sep = ", "
	}
	msg.Detail("%-20s\t%s", "security_groups", securityGroups)
	msg.Detail("%-20s\t%d", "count", p.Count())
	if p.DnsCNameRecords() != nil {
		msg.Detail("")
		for _, record := range p.DnsCNameRecords() {
			msg.Detail("%-20s\t%s", "dns cname record", record.Id())
		}
		if i := p.PrimaryInstance(); i != nil {
			msg.Detail("%-20s\t%s", "primary", i.Name())
		}
		if s := p.SecondaryInstances(); s != nil {
			secondary := ""
			for _, i := range s {
				secondary += i.Name() + " "
			}
			msg.Detail("%-20s\t%s", "secondary", secondary)
		}
	}
	msg.IndentInc()
	p.RouteInOrder(req)
	msg.IndentDec()
}

// Create

func (p *Pod) create(req *route.Request) route.Response {
	msg.Info("Pod Creation: %s", p.Name())
	if p.Created() && !req.Flag("podonly") {
		msg.Detail("Pod exists, skipping...")
		return route.OK
	}
	if resp := p.Derived().PreCreate(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().Create(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().PostCreate(req); resp != route.OK {
		return resp
	}
	msg.Detail("Pod Created: %s", p.Name())
	aaa.Accounting("Pod created: %s", p.Name())
	return route.OK
}

func (p *Pod) PreCreate(req *route.Request) route.Response {
	return route.OK
}

func (p *Pod) Create(req *route.Request) route.Response {
	return p.routeToChildren(req)
}

func (p *Pod) PostCreate(req *route.Request) route.Response {
	// Create the cname record for this pod if it exists.
	if p.cnameRecords == nil {
		return route.OK
	}
	for _, record := range p.cnameRecords {
		if resp := record.Route(req); resp != route.OK {
			return resp
		}
	}
	return route.OK
}

// Destroy

func (p *Pod) destroy(req *route.Request) route.Response {
	msg.Info("Pod Destruction: %s", p.Name())
	if p.Destroyed() {
		msg.Detail("Pod does not exist, skipping...")
		return route.OK
	}
	if resp := p.Derived().PreDestroy(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().Destroy(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().PostDestroy(req); resp != route.OK {
		return resp
	}
	msg.Detail("Pod Destroyed: %s", p.Name())
	aaa.Accounting("Pod destroyed: %s", p.Name())
	return route.OK
}

func (p *Pod) PreDestroy(req *route.Request) route.Response {
	// Destroy the cname records for this pod if it exists.
	if p.cnameRecords == nil {
		return route.OK
	}
	for _, record := range p.cnameRecords {
		if resp := record.Route(req); resp != route.OK {
			return resp
		}
	}
	return route.OK
}

func (p *Pod) Destroy(req *route.Request) route.Response {
	return p.routeReverseToChildren(req)
}

func (p *Pod) PostDestroy(req *route.Request) route.Response {
	return route.OK
}

// Provision

func (p *Pod) provision(req *route.Request) route.Response {
	msg.Info("Pod Provision: %s", p.Name())
	if p.Destroyed() {
		msg.Detail("Pod does not exist, skipping...")
		return route.OK
	}
	if resp := p.Derived().PreProvision(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().Provision(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().PostProvision(req); resp != route.OK {
		return resp
	}
	msg.Detail("Pod Provisioned: %s", p.Name())
	aaa.Accounting("Pod provisioned: %s", p.Name())
	return route.OK
}

func (p *Pod) PreProvision(req *route.Request) route.Response {
	return route.OK
}

func (p *Pod) Provision(req *route.Request) route.Response {
	return p.routeToChildren(req)
}

func (p *Pod) PostProvision(req *route.Request) route.Response {
	return route.OK
}

// Start

func (p *Pod) start(req *route.Request) route.Response {
	msg.Info("Pod Start: %s", p.Name())
	if p.Destroyed() {
		msg.Detail("Pod does not exist, skipping...")
		return route.OK
	}
	if resp := p.Derived().PreStart(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().Start(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().PostStart(req); resp != route.OK {
		return resp
	}
	msg.Detail("Pod Started: %s", p.Name())
	aaa.Accounting("Pod started: %s", p.Name())
	return route.OK
}

func (p *Pod) PreStart(req *route.Request) route.Response {
	return route.OK
}

func (p *Pod) Start(req *route.Request) route.Response {
	return p.routeToChildren(req)
}

func (p *Pod) PostStart(req *route.Request) route.Response {
	return route.OK
}

// Stop

func (p *Pod) stop(req *route.Request) route.Response {
	msg.Info("Pod Stop: %s", p.Name())
	if p.Destroyed() {
		msg.Detail("Pod does not exist, skipping...")
		return route.OK
	}
	if resp := p.Derived().PreStop(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().Stop(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().PostStop(req); resp != route.OK {
		return resp
	}
	msg.Detail("Pod Stopped: %s", p.Name())
	aaa.Accounting("Pod stopped: %s", p.Name())
	return route.OK
}

func (p *Pod) PreStop(req *route.Request) route.Response {
	return route.OK
}

func (p *Pod) Stop(req *route.Request) route.Response {
	return p.routeReverseToChildren(req)
}

func (p *Pod) PostStop(req *route.Request) route.Response {
	return route.OK
}

// Restart

func (p *Pod) restart(req *route.Request) route.Response {
	msg.Info("Pod Restart: %s", p.Name())
	if p.Destroyed() {
		msg.Detail("Pod does not exist, skipping...")
		return route.OK
	}
	if resp := p.Derived().PreRestart(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().Restart(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().PostRestart(req); resp != route.OK {
		return resp
	}
	msg.Detail("Pod Restarted: %s", p.Name())
	aaa.Accounting("Pod restarted: %s", p.Name())
	return route.OK
}

func (p *Pod) PreRestart(req *route.Request) route.Response {
	return route.OK
}

func (p *Pod) Restart(req *route.Request) route.Response {
	return p.routeToChildren(req)
}

func (p *Pod) PostRestart(req *route.Request) route.Response {
	return route.OK
}

// Replace

func (p *Pod) replace(req *route.Request) route.Response {
	msg.Info("Pod Replace: %s", p.Name())
	if p.Destroyed() {
		msg.Detail("Pod does not exist, skipping...")
		return route.OK
	}
	if resp := p.Derived().PreReplace(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().Replace(req); resp != route.OK {
		return resp
	}
	if resp := p.Derived().PostReplace(req); resp != route.OK {
		return resp
	}
	msg.Detail("Pod Replaced: %s", p.Name())
	aaa.Accounting("Pod replaced: %s", p.Name())
	return route.OK
}

func (p *Pod) PreReplace(req *route.Request) route.Response {
	return route.OK
}

func (p *Pod) Replace(req *route.Request) route.Response {
	return p.routeToChildren(req)
}

func (p *Pod) PostReplace(req *route.Request) route.Response {
	return route.OK
}

// Audit

func (p *Pod) Audit(flags ...string) error {
	if err := p.Derived().PreAudit(flags...); err != nil {
		return err
	}
	if err := p.Derived().MidAudit(flags...); err != nil {
		return err
	}
	if err := p.Derived().PostAudit(flags...); err != nil {
		return err
	}
	return nil
}

func (p *Pod) PreAudit(flags ...string) error {
	return nil
}

func (p *Pod) MidAudit(flags ...string) error {
	return p.instances.Audit(flags...)
}

func (p *Pod) PostAudit(flags ...string) error {
	return nil
}
