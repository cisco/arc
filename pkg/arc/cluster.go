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

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type ClusterFactory func(resource.Compute, provider.DataCenter, *config.Cluster) (resource.Cluster, error)

var clusterFactories map[string]ClusterFactory

func RegisterClusterFactory(name string, f ClusterFactory) {
	clusterFactories[name] = f
}

func init() {
	clusterFactories = map[string]ClusterFactory{}
}

type Cluster struct {
	*resource.Resources
	*config.Cluster
	compute  resource.Compute
	pods     *pods
	derived_ resource.Cluster
}

// newCluster is the constructor for a cluster object. It returns a non-nil error upon failure.
func newCluster(compute resource.Compute, prov provider.DataCenter, cfg *config.Cluster) (resource.Cluster, error) {
	log.Debug("Initializing Cluster %q", cfg.Name())

	// Validate the config.Cluster object.
	if cfg.Pods == nil {
		return nil, fmt.Errorf("The pods element is missing from the compute configuration")
	}

	factory := clusterFactories[cfg.Name()]
	if factory != nil {
		return factory(compute, prov, cfg)
	}
	return NewDefaultCluster(compute, prov, cfg)
}

// newDefaultCluster is the default cluster contructor.
func NewDefaultCluster(compute resource.Compute, prov provider.DataCenter, cfg *config.Cluster) (*Cluster, error) {
	c := &Cluster{
		Resources: resource.NewResources(),
		Cluster:   cfg,
		compute:   compute,
	}
	c.derived_ = c

	var err error
	c.pods, err = newPods(c, prov, cfg.Pods)
	if err != nil {
		return nil, err
	}
	c.Append(c.pods)
	return c, nil
}

// Compute provides access to Cluster's parent. Compute satisfies the resource.Cluster interface.
func (c *Cluster) Compute() resource.Compute {
	return c.compute
}

// Pods provides access to Compute's child pods. Pods satisfies the resource.Cluster interface.
func (c *Cluster) Pods() resource.Pods {
	return c.pods
}

// Find pod by name. This implies pods are named uniquely. FindPod satisfies the resource.Cluster interface.
func (c *Cluster) FindPod(name string) resource.Pod {
	return c.pods.Find(name)
}

// FindInstance finds an instance from this cluster by name. Instances must be uniquely named.
// The instance name convention is "<pod name>-<instance number>". FindInstance satisfies the
// resource.Cluster interface.
func (c *Cluster) FindInstance(name string) resource.Instance {
	return c.pods.FindInstance(name)
}

// FindInstanceByIP find and instance from this cluster by ip address.
// FindInstaneByIP satisfies the resource.Cluster interface.
func (c *Cluster) FindInstanceByIP(ip string) resource.Instance {
	return c.pods.FindInstanceByIP(ip)
}

func (c *Cluster) Derived() resource.Cluster {
	return c.derived_
}

func (c *Cluster) SetDerived(d resource.Cluster) {
	c.derived_ = d
}

// Route satisfies the embedded resource.Resource interface in resource.Cluster.
// Cluster handles load, create, destroy, start, stop, replace, help, config and
// info requests in order to manage the cluster. All other commands are routed to
// the cluster's pods.
func (c *Cluster) Route(req *route.Request) route.Response {
	log.Route(req, "Cluster %q", c.Name())

	// Route to the appropriate resource. Pod is the only valid sub-resource.
	if req.Top() == "pod" {
		req.Pop()
		if req.Top() == "" {
			podHelp("")
			return route.FAIL
		}
		pod := c.FindPod(req.Top())
		if pod == nil {
			msg.Error("Unknown pod %q.", req.Top())
			return route.FAIL
		}
		return pod.Route(req.Pop())
	}

	// The path should be empty at this point.
	if req.Top() != "" {
		c.help()
		return route.FAIL
	}

	if err := aaa.Authorized(req, "cluster", c.Name()); err != nil {
		msg.Error(err.Error())
		return route.UNAUTHORIZED
	}

	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	switch req.Command() {
	case route.Load:
		return c.RouteInOrder(req)
	case route.Help:
		c.help()
		return route.OK
	case route.Config:
		c.config()
		return route.OK
	case route.Info:
		c.info(req)
		return route.OK
	case route.Create:
		return c.create(req)
	case route.Destroy:
		return c.destroy(req)
	case route.Provision:
		return c.provision(req)
	case route.Start:
		return c.start(req)
	case route.Stop:
		return c.stop(req)
	case route.Restart:
		return c.restart(req)
	case route.Replace:
		return c.replace(req)
	default:
		msg.Error("Unknown cluster command %q.", req.Command().String())
	}
	return route.FAIL
}

func (c *Cluster) routeToChildren(req *route.Request) route.Response {
	if !req.Flag("clusteronly") {
		return c.RouteInOrder(req)
	}
	return route.OK
}

func (c *Cluster) routeReverseToChildren(req *route.Request) route.Response {
	if !req.Flag("clusteronly") {
		return c.RouteReverseOrder(req)
	}
	return route.OK
}

func (c *Cluster) help() {
	commands := []help.Command{
		{Name: route.Create.String(), Desc: fmt.Sprintf("create %s cluster", c.Name())},
		{Name: route.Provision.String(), Desc: fmt.Sprintf("provision %s cluster", c.Name())},
		{Name: route.Provision.String() + " users", Desc: fmt.Sprintf("update %s cluster users", c.Name())},
		{Name: route.Start.String(), Desc: fmt.Sprintf("start %s cluster", c.Name())},
		{Name: route.Stop.String(), Desc: fmt.Sprintf("stop %s cluster", c.Name())},
		{Name: route.Restart.String(), Desc: fmt.Sprintf("restart %s cluster", c.Name())},
		{Name: route.Replace.String(), Desc: fmt.Sprintf("replace %s cluster", c.Name())},
		{Name: route.Audit.String(), Desc: fmt.Sprintf("audit %s cluster", c.Name())},
		{Name: route.Destroy.String(), Desc: fmt.Sprintf("destroy %s cluster", c.Name())},
		{Name: route.Config.String(), Desc: fmt.Sprintf("provide the %s cluster configuration", c.Name())},
		{Name: route.Info.String(), Desc: fmt.Sprintf("provide information about allocated %s cluster", c.Name())},
		{Name: route.Help.String(), Desc: "provide this help"},
	}
	help.Print(fmt.Sprintf("cluster %s", c.Name()), commands)
}

func (c *Cluster) config() {
	c.Cluster.Print()
}

func (c *Cluster) info(req *route.Request) {
	if c.Destroyed() {
		return
	}
	msg.Info("Cluster")
	msg.Detail("%-20s\t%s", "name", c.Name())
	msg.IndentInc()
	c.RouteInOrder(req)
	msg.IndentDec()
}

// Create

func (c *Cluster) create(req *route.Request) route.Response {
	msg.Info("Cluster Creation: %s", c.Name())
	if c.Created() {
		msg.Detail("Cluster exists, skipping...")
		return route.OK
	}
	if resp := c.Derived().PreCreate(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().Create(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().PostCreate(req); resp != route.OK {
		return resp
	}
	msg.Detail("Cluster Created: %s", c.Name())
	aaa.Accounting("Cluster created: %s", c.Name())
	return route.OK
}

func (c *Cluster) PreCreate(req *route.Request) route.Response {
	return route.OK
}

func (c *Cluster) Create(req *route.Request) route.Response {
	return c.routeToChildren(req)
}

func (c *Cluster) PostCreate(req *route.Request) route.Response {
	return route.OK
}

// Destroy

func (c *Cluster) destroy(req *route.Request) route.Response {
	msg.Info("Cluster Destruction: %s", c.Name())
	if c.Destroyed() {
		msg.Detail("Cluster does not exist, skipping...")
		return route.OK
	}
	if resp := c.Derived().PreDestroy(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().Destroy(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().PostDestroy(req); resp != route.OK {
		return resp
	}
	msg.Detail("Cluster Destroyed: %s", c.Name())
	aaa.Accounting("Cluster destroyed: %s", c.Name())
	return route.OK
}

func (c *Cluster) PreDestroy(req *route.Request) route.Response {
	return route.OK
}

func (c *Cluster) Destroy(req *route.Request) route.Response {
	return c.routeReverseToChildren(req)
}

func (c *Cluster) PostDestroy(req *route.Request) route.Response {
	return route.OK
}

// Provision

func (c *Cluster) provision(req *route.Request) route.Response {
	msg.Info("Cluster Provision: %s", c.Name())
	if c.Destroyed() {
		msg.Detail("Cluster does not exist, skipping...")
		return route.OK
	}
	if resp := c.Derived().PreProvision(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().Provision(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().PostProvision(req); resp != route.OK {
		return resp
	}
	msg.Detail("Cluster Provisioned: %s", c.Name())
	aaa.Accounting("Cluster provisioned: %s", c.Name())
	return route.OK
}

func (c *Cluster) PreProvision(req *route.Request) route.Response {
	return route.OK
}

func (c *Cluster) Provision(req *route.Request) route.Response {
	return c.routeToChildren(req)
}

func (c *Cluster) PostProvision(req *route.Request) route.Response {
	return route.OK
}

// Start

func (c *Cluster) start(req *route.Request) route.Response {
	msg.Info("Cluster Start: %s", c.Name())
	if c.Destroyed() {
		msg.Detail("Cluster does not exist, skipping...")
		return route.OK
	}
	if resp := c.Derived().PreStart(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().Start(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().PostStart(req); resp != route.OK {
		return resp
	}
	msg.Detail("Cluster Started: %s", c.Name())
	aaa.Accounting("Cluster started: %s", c.Name())
	return route.OK
}

func (c *Cluster) PreStart(req *route.Request) route.Response {
	return route.OK
}

func (c *Cluster) Start(req *route.Request) route.Response {
	return c.routeToChildren(req)
}

func (c *Cluster) PostStart(req *route.Request) route.Response {
	return route.OK
}

// Stop

func (c *Cluster) stop(req *route.Request) route.Response {
	msg.Info("Cluster Stop: %s", c.Name())
	if c.Destroyed() {
		msg.Detail("Cluster does not exist, skipping...")
		return route.OK
	}
	if resp := c.Derived().PreStop(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().Stop(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().PostStop(req); resp != route.OK {
		return resp
	}
	msg.Detail("Cluster Stopped: %s", c.Name())
	aaa.Accounting("Cluster stopped: %s", c.Name())
	return route.OK
}

func (c *Cluster) PreStop(req *route.Request) route.Response {
	return route.OK
}

func (c *Cluster) Stop(req *route.Request) route.Response {
	return c.routeReverseToChildren(req)
}

func (c *Cluster) PostStop(req *route.Request) route.Response {
	return route.OK
}

// Restart

func (c *Cluster) restart(req *route.Request) route.Response {
	msg.Info("Cluster Restart: %s", c.Name())
	if c.Destroyed() {
		msg.Detail("Cluster does not exist, skipping...")
		return route.OK
	}
	if resp := c.Derived().PreRestart(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().Restart(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().PostRestart(req); resp != route.OK {
		return resp
	}
	msg.Detail("Cluster Restarted: %s", c.Name())
	aaa.Accounting("Cluster restarted: %s", c.Name())
	return route.OK
}

func (c *Cluster) PreRestart(req *route.Request) route.Response {
	return route.OK
}

func (c *Cluster) Restart(req *route.Request) route.Response {
	return c.routeToChildren(req)
}

func (c *Cluster) PostRestart(req *route.Request) route.Response {
	return route.OK
}

// Replace

func (c *Cluster) replace(req *route.Request) route.Response {
	msg.Info("Cluster Replace: %s", c.Name())
	if c.Destroyed() {
		msg.Detail("Cluster does not exist, skipping...")
		return route.OK
	}
	if resp := c.Derived().PreReplace(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().Replace(req); resp != route.OK {
		return resp
	}
	if resp := c.Derived().PostReplace(req); resp != route.OK {
		return resp
	}
	msg.Detail("Cluster Replaced: %s", c.Name())
	aaa.Accounting("Cluster replaced: %s", c.Name())
	return route.OK
}

func (c *Cluster) PreReplace(req *route.Request) route.Response {
	return route.OK
}

func (c *Cluster) Replace(req *route.Request) route.Response {
	return c.routeToChildren(req)
}

func (c *Cluster) PostReplace(req *route.Request) route.Response {
	return route.OK
}

// Audit

func (c *Cluster) Audit(flags ...string) error {
	if err := c.Derived().PreAudit(flags...); err != nil {
		return err
	}
	if err := c.Derived().MidAudit(flags...); err != nil {
		return err
	}
	if err := c.Derived().PostAudit(flags...); err != nil {
		return err
	}
	return nil
}

func (c *Cluster) PreAudit(flags ...string) error {
	return nil
}

func (c *Cluster) MidAudit(flags ...string) error {
	if err := c.pods.Audit(flags...); err != nil {
		return err
	}
	return nil
}

func (c *Cluster) PostAudit(flags ...string) error {
	return nil
}
