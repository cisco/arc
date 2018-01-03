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
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/provider"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

type pods struct {
	*resource.Resources
	*config.Pods
	pods map[string]resource.Pod
}

// newPods is the constructor for a pods object. It returns a non-nil error upon failure.
func newPods(cluster *Cluster, provider provider.DataCenter, cfg *config.Pods) (*pods, error) {
	log.Debug("Initializing Pods")

	p := &pods{
		Resources: resource.NewResources(),
		Pods:      cfg,
		pods:      map[string]resource.Pod{},
	}

	for _, conf := range *cfg {
		if p.Find(conf.Name()) != nil {
			return nil, fmt.Errorf("Pod name %q must be unique but is used multiple times", conf.Name())
		}
		pod, err := newPod(cluster, provider, conf)
		if err != nil {
			return nil, err
		}
		p.pods[pod.Name()] = pod
		p.Append(pod)
	}
	return p, nil
}

// Find pod by name. This implies pods are named uniquely.
// Find satisfies the resource.Pods interface.
func (p *pods) Find(name string) resource.Pod {
	return p.pods[name]
}

// FindInstance finds and instance of this pod by name.
// This implies instances are uniquely named.
// The name takes the form "<pod name>-<instance number>".
// FindInstance satisfies the resource.Pods interface.
func (p *pods) FindInstance(name string) resource.Instance {
	for _, pod := range p.pods {
		instance := pod.FindInstance(name)
		if instance != nil {
			return instance
		}
	}
	return nil
}

// FindInstanceByIP finds and instance of this pod by ip address.
// FindInstanceByIP satisfies the resource.Pods interface.
func (p *pods) FindInstanceByIP(ip string) resource.Instance {
	for _, pod := range p.pods {
		instance := pod.FindInstanceByIP(ip)
		if instance != nil {
			return instance
		}
	}
	return nil
}

// Route satisfies the embedded resource.Resource interface in resource.Pods.
func (p *pods) Route(req *route.Request) route.Response {
	log.Route(req, "Pods")

	// Route to the appropriate resource.
	if req.Top() == "pod" {
		req.Pop()
		if req.Top() == "" {
			podHelp("")
			return route.FAIL
		}
		pod := p.Find(req.Top())
		if pod == nil {
			msg.Error("Unknown pod %q.", req.Top())
			return route.FAIL
		}
		return pod.Route(req.Pop())
	}

	// Handle the command.
	switch req.Command() {
	case route.Load, route.Create, route.Provision, route.Start, route.Stop, route.Restart, route.Replace:
		return p.RouteInOrder(req)
	case route.Destroy:
		return p.RouteReverseOrder(req)
	case route.Info:
		return p.info(req)
	}
	return route.FAIL
}

func (p *pods) Audit(flags ...string) error {
	for _, v := range p.pods {
		if err := v.Audit(flags...); err != nil {
			return err
		}
	}
	return nil
}

func (p *pods) info(req *route.Request) route.Response {
	if p.Destroyed() {
		return route.OK
	}
	msg.Info("Pods")
	msg.IndentInc()
	p.RouteInOrder(req)
	msg.IndentDec()
	return route.OK
}
