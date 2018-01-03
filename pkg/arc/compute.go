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

type auditResource int

const (
	inst auditResource = iota
	eip
	vol
)

type compute struct {
	*resource.Resources
	*config.Compute
	dc              *dataCenter
	providerCompute resource.ProviderCompute
	keypair         *keypair
	clusters        *clusters
}

// newCompute is the constructor for a compute object. It returns a non-nil error upon failure.
func newCompute(dc *dataCenter, prov provider.DataCenter, cfg *config.Compute) (*compute, error) {
	log.Debug("Initializing Compute")

	// Validate the config.Compute object.
	if cfg.Clusters == nil {
		return nil, fmt.Errorf("The clusters element is missing from the compute configuration")
	}

	c := &compute{
		Resources: resource.NewResources(),
		Compute:   cfg,
		dc:        dc,
	}

	var err error

	// Delegate the provider specific compute behavior to the resource.ProviderCompute object.
	c.providerCompute, err = prov.NewCompute(cfg)
	if err != nil {
		return nil, err
	}

	c.keypair, err = newKeyPair(prov)
	if err != nil {
		return nil, err
	}
	c.Append(c.keypair)

	c.clusters, err = newClusters(c, prov, cfg.Clusters)
	if err != nil {
		return nil, err
	}
	c.Append(c.clusters)
	return c, nil
}

// AuditVolumes identifies any volumes that have been deployed but are not in the configuration.
func (c *compute) AuditVolumes(flags ...string) error {
	return c.providerCompute.AuditVolumes(flags...)
}

// AuditEIP identifies any elastic IPs that have been allocated but are associated with anything.
func (c *compute) AuditEIP(flags ...string) error {
	return c.providerCompute.AuditEIP(flags...)
}

// AuditInstances identifies any instances that have been deployed but are not in the configuration.
func (c *compute) AuditInstances(flags ...string) error {
	return c.providerCompute.AuditInstances(flags...)
}

// auditHelper is a function that handles getting the audits completed
func (c *compute) auditHelper(req *route.Request, auditType auditResource) route.Response {
	// Skip if the test flag is set.
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}
	switch auditType {
	case inst:
		top := req.Top()
		if req.Pop().Top() != "" {
			req.Path().Push(top)
			return c.clusters.Route(req)
		}
		if err := c.Audit("Instance"); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case eip:
		if err := c.AuditEIP("EIP"); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	case vol:
		if err := c.AuditVolumes("Volume"); err != nil {
			msg.Error(err.Error())
			return route.FAIL
		}
	}
	return route.OK
}

// auditAll preforms an audit on all the resources it needs to audit
func (c *compute) auditAll(req *route.Request) route.Response {
	if resp := c.auditHelper(req, inst); resp != route.OK {
		return resp
	}
	if resp := c.auditHelper(req, vol); resp != route.OK {
		return resp
	}
	if resp := c.auditHelper(req, eip); resp != route.OK {
		return resp
	}
	return route.OK
}

// DataCenter satisfies the resource.Compute interface and provides access
// to compute's parent.
func (c *compute) DataCenter() resource.DataCenter {
	return c.dc
}

// KeyPair satisfies the resource.Compute interface and provides access
// to compute's keypair.
func (c *compute) KeyPair() resource.KeyPair {
	return c.keypair
}

// Clusters satisfies the resource.Compute interface and provides access
// to compute's clusters.
func (c *compute) Clusters() resource.Clusters {
	return c.clusters
}

// Find cluster by name. This implies clusters are named uniquely.
func (c *compute) FindCluster(name string) resource.Cluster {
	return c.clusters.Find(name)
}

// Find pod by name. This implies pods are named uniquely.
func (c *compute) FindPod(name string) resource.Pod {
	return c.clusters.FindPod(name)
}

// Find instance by name. This implies instances are named uniquely.
// The name takes the form "<pod name>-<instance number>".
func (c *compute) FindInstance(name string) resource.Instance {
	return c.clusters.FindInstance(name)
}

// Find the instance by ip address.
func (c *compute) FindInstanceByIP(ip string) resource.Instance {
	return c.clusters.FindInstanceByIP(ip)
}

// ProviderCompute provides access to the provider specific compute.
func (c *compute) ProviderCompute() resource.ProviderCompute {
	return c.providerCompute
}

// Route satisfies the embedded resource.Resource interface in resource.Compute.
// Compute terminates load, create, destroy, help, config and info commands.
// All other commands are routed to compute's children.
func (c *compute) Route(req *route.Request) route.Response {
	log.Route(req, "Compute")

	switch req.Top() {
	case "":
		break
	case "keypair":
		return c.keypair.Route(req.Pop())
	case "cluster":
		return c.clusters.Route(req.Pop())
	case "pod":
		return c.clusters.Route(req)
	case "instance":
		if req.Command() != route.Audit {
			return c.clusters.Route(req)
		}
		return c.auditHelper(req, inst)
	case "volume":
		if req.Command() != route.Audit {
			msg.Error("No command %q found for volumes", req.Command())
			return route.FAIL
		}
		return c.auditHelper(req, vol)
	case "eip":
		if req.Command() != route.Audit {
			msg.Error("No command %q found for elastic IPs", req.Command())
			return route.FAIL
		}
		return c.auditHelper(req, eip)
	default:
		c.help()
		return route.FAIL
	}

	// Skip if the test flag is set
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
	case route.Audit:
		return c.auditAll(req)
	default:
		msg.Error("Unknown compute command '%s'.", req.Command())
	}
	return route.FAIL
}

func (c *compute) Audit(flags ...string) error {
	if len(flags) == 0 || flags[0] == "" {
		fmt.Errorf("No flag set to find audit object")
	}
	err := aaa.NewAudit(flags[0])
	if err != nil {
		return err
	}
	if err := c.AuditInstances(flags...); err != nil {
		return err
	}
	return c.clusters.Audit(flags...)
}

func (c *compute) help() {
	commands := []help.Command{
		{route.Config.String(), "show the compute configuration"},
		{route.Info.String(), "show information about allocated compute resource"},
		{route.Help.String(), "show this help"},
	}
	help.Print("compute", commands)
}

func (c *compute) config() {
	c.Compute.Print()
}

func (c *compute) info(req *route.Request) {
	if c.Destroyed() {
		return
	}
	msg.Info("Compute")
	msg.IndentInc()
	c.Compute.PrintLocal()
	c.RouteInOrder(req)
	msg.IndentDec()
}
