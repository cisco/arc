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

package command

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/ssh"
)

type client struct {
	client *ssh.Client
}

func newClient(i resource.Instance, asRoot bool) (*client, error) {

	// Create the underlying ssh client.
	cl, err := ssh.NewClient()
	if err != nil {
		return nil, err
	}
	s := &client{
		client: cl,
	}

	// Find the jump host unless we are the jump host
	var bastion resource.Instance
	if i.Pod().ServerType() != "bastion" {
		pod := i.Pod().Cluster().Compute().FindPod("bastion")
		for _, b := range pod.Instances().GetInstances() {
			if b.State() == "running" {
				bastion = b
				break
			}
		}
		if bastion == nil {
			return nil, fmt.Errorf("Cannot find a running bastion server")
		}
	}

	instanceUser := env.Lookup("SSH_USER")
	if asRoot {
		instanceUser = i.RootUser()
	}

	// The bastion user is always the ssh user. If you can jump thru the
	// bastion it has been provisioned with all users so there is no need to
	// use the root user.
	bastionUser := env.Lookup("SSH_USER")

	count, max := 0, 600
	for ; count < max; count++ {
		if bastion != nil {
			log.Info("Creating ssh connection to %s - %s, via %s - %s",
				i.Name(), i.PrivateIPAddress(), bastion.Name(), bastion.PublicIPAddress())

			// The instance which we are trying to connect to does not have
			// a public ip address, so we need to jump thru the bastion.
			log.Debug("Instance ssh user: %s", instanceUser)
			instanceAddr := ssh.NewAddress(instanceUser, i.PrivateIPAddress())

			log.Debug("Bastion ssh user: %s", bastionUser)
			bastionAddr := ssh.NewAddress(bastionUser, bastion.PublicIPAddress())
			err := s.client.JumpConnect(bastionAddr, instanceAddr)
			if err == nil {
				break
			}
			log.Verbose("%v", err)
		} else {
			log.Info("Creating ssh connection to %s - %s", i.Name(), i.PublicIPAddress())

			// The instance we are trying to connect to has a public ip address.
			// We will connect directly to it.
			log.Debug("Instance ssh user: %s", instanceUser)
			instanceAddr := ssh.NewAddress(instanceUser, i.PublicIPAddress())
			err := s.client.DirectConnect(instanceAddr)
			if err == nil {
				break
			}
			log.Verbose("%v", err)
		}
		if count == 0 {
			msg.Detail("Waiting for ssh to connect to %s", i.Name())
			msg.Raw(msg.Tab())
		}
		if count%120 < 60 {
			msg.Raw(".")
		} else {
			msg.Raw("\b \b")
		}
		time.Sleep(time.Second)
	}
	if count > 0 {
		msg.Raw("\n")
	}
	if count == max {
		return nil, fmt.Errorf("Failed to connect to %s", i.Name())
	}

	return s, nil
}

func (s *client) close() error {
	if s.client == nil {
		return nil
	}
	return s.client.Close()
}

func (s *client) copy(src, dest string) ([]byte, error) {
	if s.client == nil {
		return nil, fmt.Errorf("Client does not exist")
	}
	msg.Detail("Copying file '%s' to '%s'", filepath.Base(src), dest)
	return s.client.Copy(src, dest)
}

func (s *client) run(cmd string) ([]byte, error) {
	if s.client == nil {
		return nil, fmt.Errorf("Client does not exist")
	}
	msg.Detail("Running command '%s'", cmd)
	return s.client.Run(cmd)
}

func (s *client) sudo(cmd string) ([]byte, error) {
	if s.client == nil {
		return nil, fmt.Errorf("Client does not exist")
	}
	msg.Detail("Running sudo command '%s'", cmd)
	return s.client.Sudo(cmd)
}
