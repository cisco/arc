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

type clusters struct {
	*resource.Resources
	*config.Clusters
	clusters map[string]resource.Cluster
}

// newClusters is the constructor for a clusters object. It returns a non-nil error upon failure.
func newClusters(compute *compute, prov provider.DataCenter, cfg *config.Clusters) (*clusters, error) {
	log.Debug("Initializing Clusters")

	c := &clusters{
		Resources: resource.NewResources(),
		Clusters:  cfg,
		clusters:  map[string]resource.Cluster{},
	}

	for _, conf := range *cfg {
		if c.Find(conf.Name()) != nil {
			return nil, fmt.Errorf("Cluster name %q must be unique but is used multiple times", conf.Name())
		}
		cluster, err := newCluster(compute, prov, conf)
		if err != nil {
			return nil, err
		}
		c.clusters[cluster.Name()] = cluster
		c.Append(cluster)
	}
	return c, nil
}

// Find satisfies the resource.Clusters interface and provides a way
// to search for a specific cluster. This assumes cluster names are unique.
func (c *clusters) Find(name string) resource.Cluster {
	return c.clusters[name]
}

// FindPod satisfies the resource.Clusters interface and provides a way
// to search for a specific pod. This assumes pod names are unique.
func (c *clusters) FindPod(name string) resource.Pod {
	for _, cluster := range c.clusters {
		pod := cluster.FindPod(name)
		if pod != nil {
			return pod
		}
	}
	return nil
}

// FindInstance satisfies the resource.Clusters interface and provides a way
// to search for a specific instance. This assumes instance names are unique.
func (c *clusters) FindInstance(name string) resource.Instance {
	for _, cluster := range c.clusters {
		instance := cluster.FindInstance(name)
		if instance != nil {
			return instance
		}
	}
	return nil
}

// FindInstanceByIP satisfies the resource.Clusters interface and provides a way
// to search for a specific instance by ip address. This assumes an ip address
// assigned to an instance is unique.
func (c *clusters) FindInstanceByIP(ip string) resource.Instance {
	for _, cluster := range c.clusters {
		instance := cluster.FindInstanceByIP(ip)
		if instance != nil {
			return instance
		}
	}
	return nil
}

// Route satisfies the embedded resource.Resource interface in resource.Clusters.
func (c *clusters) Route(req *route.Request) route.Response {
	log.Route(req, "Clusters")

	// Route to the appropriate resource. First try follow the resource path.
	switch req.Top() {
	case "pod":
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
	case "instance":
		req.Pop()
		if req.Top() == "" {
			instanceHelp("")
			return route.FAIL
		}
		instance := c.FindInstance(req.Top())
		if instance == nil {
			msg.Error("Unknown instance %q.", req.Top())
			return route.FAIL
		}
		return instance.Route(req.Pop())
	}

	// Is the resource the name of a cluster?
	if cluster := c.Find(req.Top()); cluster != nil {
		return cluster.Route(req.Pop())
	}
	if req.Top() != "" {
		msg.Error("Unknown cluster %q.", req.Top())
		return route.FAIL
	}

	// Skip if the test flag is set.
	if req.TestFlag() {
		msg.Detail("Test. Skipping...")
		return route.OK
	}

	// Handle the command
	switch req.Command() {
	case route.Load:
		return c.RouteInOrder(req)
	case route.Help:
		c.help()
		return route.OK
	case route.Info:
		c.info(req)
		return route.OK
	case route.Config:
		c.config()
		return route.OK
	case route.Provision:
		return c.RouteInOrder(req)
	default:
		msg.Error("Unknown cluster command %q.", req.Command().String())
	}
	return route.FAIL
}

func (c *clusters) Audit(flags ...string) error {
	for _, v := range c.clusters {
		if err := v.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (c *clusters) help() {
	commands := []help.Command{
		{Name: "'name'", Desc: "manage named cluster"},
		{Name: route.Config.String(), Desc: "provide the clusters configuration"},
		{Name: route.Info.String(), Desc: "provide information about allocated clusters"},
		{Name: route.Help.String(), Desc: "provide this help"},
	}
	help.Print("cluster", commands)
}

func (c *clusters) config() {
	c.Clusters.Print()
}

func (c *clusters) info(req *route.Request) {
	if c.Destroyed() {
		return
	}
	msg.Info("Clusters")
	msg.IndentInc()
	c.RouteInOrder(req)
	msg.IndentDec()
}
