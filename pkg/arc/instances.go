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

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type instances struct {
	*resource.Resources
	*config.Instances
	pod       *Pod
	instances map[string]resource.Instance
}

// newInstancess is the constructor for a instances object. It returns a non-nil error upon failure.
func newInstances(pod *Pod, prov provider.DataCenter, cfg *config.Instances) (*instances, error) {
	log.Debug("Initializing Instances")

	i := &instances{
		Resources: resource.NewResources(),
		pod:       pod,
		instances: map[string]resource.Instance{},
	}

	// The reference to the network resource.
	net := pod.Cluster().Compute().DataCenter().Network()

	// The availability zones available to these instances.
	availabilityZones := net.AvailabilityZones()

	// The subnet group associated with these instances.
	subnetGroup := net.SubnetGroups().Find(pod.SubnetGroup())
	if subnetGroup == nil {
		return nil, fmt.Errorf("Cannot find subnet group %s configured for pod %s", pod.SubnetGroup(), pod.Name())
	}

	// The keypair to be used with these instances.
	keypair := pod.Cluster().Compute().KeyPair()

	n := 0
	for _, conf := range *cfg {
		// Ensure the instance is uniquely named.
		if i.Find(conf.Name()) != nil {
			return nil, fmt.Errorf("Instance name %q must be unique but is used multiple times", conf.Name())
		}

		// The availability zone for this instance. Chosing via round robin. Always starting at 0.
		az := availabilityZones[n%len(availabilityZones)]

		// Get the subnet associated with the AZ.
		subnetName := pod.SubnetGroup() + "-" + az
		subnet := subnetGroup.Find(subnetName)
		if subnet == nil {
			return nil, fmt.Errorf("Cannot find subnet %s configured for instance %s", subnetName, conf.Name())
		}

		instance, err := newInstance(pod, subnet, keypair, prov, conf)
		if err != nil {
			return nil, err
		}
		i.instances[instance.Name()] = instance
		i.Append(instance)

		n++
	}
	return i, nil
}

// GetInstances returns a map of all instances indexed by instance name.
func (i *instances) GetInstances() map[string]resource.Instance {
	return i.instances
}

// Find instance by name. This implies instances are named uniquely.
// The name takes the form "<pod name>-<instance number>".
func (i *instances) Find(name string) resource.Instance {
	return i.instances[name]
}

// Find the instance by ip address.
func (i *instances) FindByIP(ip string) resource.Instance {
	for _, instance := range i.instances {
		if instance.PrivateIPAddress() == ip || instance.PublicIPAddress() == ip {
			return instance
		}
	}
	return nil
}

// Route satisfies the embedded resource.Resource interface in resource.Pods.
func (i *instances) Route(req *route.Request) route.Response {
	log.Route(req, "Instances")

	switch req.Top() {
	case "instance":
		req.Pop()
		if req.Top() == "" {
			instanceHelp("")
			return route.FAIL
		}
		instance := i.Find(req.Top())
		if instance == nil {
			msg.Error("Unknown instance %q.", req.Top())
			return route.FAIL
		}
		return instance.Route(req.Pop())
	}

	switch req.Command() {
	case route.Load, route.Create, route.Provision, route.Start, route.Stop, route.Restart, route.Replace:
		return i.RouteInOrder(req)
	case route.Destroy:
		return i.RouteReverseOrder(req)
	case route.Help:
		i.help()
		return route.OK
	case route.Config:
		i.config()
		return route.OK
	case route.Info:
		i.info(req)
		return route.OK
	}
	return route.FAIL
}

func (i *instances) Audit(flags ...string) error {
	for _, v := range i.instances {
		if err := v.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (i *instances) help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: "create all instances"},
		{Name: route.Provision.String(), Desc: "provision all instances"},
		{Name: route.Start.String(), Desc: "start all instances"},
		{Name: route.Stop.String(), Desc: "stop all instances"},
		{Name: route.Restart.String(), Desc: "restart all instances"},
		{Name: route.Replace.String(), Desc: "replace all instances"},
		{Name: route.Audit.String(), Desc: "audit all instances"},
		{Name: route.Destroy.String(), Desc: "destroy all instances"},
		{Name: "'name'", Desc: "manage named instance"},
		{Name: route.Config.String(), Desc: "provide the instances configuration"},
		{Name: route.Info.String(), Desc: "provide information about allocated instances"},
		{Name: route.Help.String(), Desc: "provide this help"},
	}
	help.Print("pods command", commands)
}

func (i *instances) config() {
	i.Instances.Print()
}

func (i *instances) info(req *route.Request) {
	if i.Destroyed() {
		return
	}
	msg.Info("Instances")
	msg.IndentInc()
	i.RouteInOrder(req)
	msg.IndentDec()
}
