//
// Copyright (c) 2018, Cisco Systems
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

package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

type containerService struct {
	*config.ContainerService
	ecs *ecs.ECS

	cluster *ecs.Cluster
}

func newContainerService(cfg *config.ContainerService, p *containerServiceProvider) (resource.ProviderContainerService, error) {
	log.Debug("Initializaing AWS Container Cluster %q", cfg.Name())
	return &containerService{
		ContainerService: cfg,
		ecs:              p.ecs,
	}, nil
}

func (cs *containerService) State() string {
	if cs.Destroyed() || cs.cluster.Status == nil {
		return ""
	}
	return *cs.cluster.Status
}

func (cs *containerService) set(c *ecs.Cluster) {
	if c == nil {
		return
	}
	cs.cluster = c
}

func (cs *containerService) clear() {
	cs.cluster = nil
}

func (cs *containerService) Load() error {
	log.Debug("Loading AWS Container Cluster: %q", cs.Name())

	params := &ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(cs.Name())},
	}
	resp, err := cs.ecs.DescribeClusters(params)
	if err != nil {
		return err
	}
	if len(resp.Clusters) == 0 {
		return nil
	}
	if len(resp.Clusters) != 1 {
		return fmt.Errorf("Internal error: multiple AWS ECS clusters named %q", cs.Name())
	}
	cs.set(resp.Clusters[0])

	return nil
}

func (cs *containerService) Create(flags ...string) error {
	log.Debug("Creating AWS Container Cluster: %q", cs.Name())
	if cs.Created() {
		return nil
	}
	// TODO
	return nil
}

func (cs *containerService) Created() bool {
	return cs.cluster != nil
}

func (cs *containerService) Destroy(flags ...string) error {
	log.Debug("Creating AWS Container Cluster: %q", cs.Name())
	if cs.Destroyed() {
		return nil
	}
	// TODO
	return nil
}

func (cs *containerService) Destroyed() bool {
	return cs.cluster == nil
}

func (cs *containerService) Provision(flags ...string) error {
	log.Debug("Provisioning AWS Container Cluster: %q", cs.Name())
	if cs.Destroyed() {
		return nil
	}
	// TODO
	return nil
}

func (cs *containerService) Audit(flags ...string) error {
	log.Debug("Audting AWS Container Cluster: %q", cs.Name())
	if cs.Destroyed() {
		return nil
	}
	// TODO
	return nil
}

func (cs *containerService) Info() {
	if cs.Destroyed() {
		return
	}
	if cs.cluster.ClusterName != nil {
		msg.Detail("%-40s\t%s", "name", *cs.cluster.ClusterName)
	}
	if cs.cluster.ClusterArn != nil {
		msg.Detail("%-40s\t%s", "arn", *cs.cluster.ClusterArn)
	}
	if cs.cluster.Status != nil {
		msg.Detail("%-40s\t%s", "status", *cs.cluster.Status)
	}
	if cs.cluster.RegisteredContainerInstancesCount != nil {
		msg.Detail("%-40s\t%d", "registered instances", *cs.cluster.RegisteredContainerInstancesCount)
	}
	if cs.cluster.RunningTasksCount != nil {
		msg.Detail("%-40s\t%d", "running tasks", *cs.cluster.RunningTasksCount)
	}
	if cs.cluster.PendingTasksCount != nil {
		msg.Detail("%-40s\t%d", "pending tasks", *cs.cluster.PendingTasksCount)
	}
	if cs.cluster.ActiveServicesCount != nil {
		msg.Detail("%-40s\t%d", "active services", *cs.cluster.ActiveServicesCount)
	}
	for _, s := range cs.cluster.Statistics {
		if s.Name != nil && s.Value != nil {
			msg.Detail("%-40s\t%d", *s.Name, *s.Value)
		}
	}
}
