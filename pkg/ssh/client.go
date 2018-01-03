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

package ssh

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Client struct {
	agentConn  net.Conn
	config     *ssh.ClientConfig
	jumpConn   *ssh.Client
	remoteConn net.Conn
	client     *ssh.Client
}

// NewClient is a the Client constructor. It connects to the running ssh-agent and
// sets up the user name. The user name can be overriden using the SSH_USER environment
// variable.
func NewClient() (*Client, error) {
	client := &Client{}

	// Connect to the ssh-agent
	authMethods := []ssh.AuthMethod{}
	agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, err
	}
	authMethods = append(authMethods, ssh.PublicKeysCallback(agent.NewClient(agentConn).Signers))

	// Find the user name, prefer SSH_USER environment variable to user name.
	username := os.Getenv("SSH_USER")
	if username == "" {
		u, err := user.Current()
		if err != nil {
			agentConn.Close()
			return nil, err
		}
		username = u.Username
	}

	client.agentConn = agentConn
	client.config = &ssh.ClientConfig{
		User:            username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return client, nil
}

// DirectConnect establishes an authenticated ssh connection to the given address.
// If the connection succeeds it will need to be closed.
func (c *Client) DirectConnect(addr Address) error {
	if c.client != nil {
		return ClientError{"Connect: client already connected"}
	}

	if addr.User != "" {
		c.config.User = addr.User
	}

	client, err := ssh.Dial("tcp", addr.Addr, c.config)
	if err != nil {
		return err
	}
	c.client = client

	return nil
}

// JumpConnect established an authenticated ssh connection via a jump
// host (aka bastion), meaning that this will first connect to the
// jump host, then connect to the remote address. If the connection
// succeeds it will need to be closed.
func (c *Client) JumpConnect(jumpAddr, remoteAddr Address) error {
	if c.client != nil {
		return ClientError{"Connect: client already connected"}
	}

	// Connect to the jump host (aka bastion).
	jumpCfg := *c.config
	if jumpAddr.User != "" {
		jumpCfg.User = jumpAddr.User
	}

	jumpConn, err := ssh.Dial("tcp", jumpAddr.Addr, &jumpCfg)
	if err != nil {
		return err
	}

	// Connect to the remote host via the jump host.
	remoteConn, err := jumpConn.Dial("tcp", remoteAddr.Addr)
	if err != nil {
		jumpConn.Close()
		return err
	}

	remoteCfg := *c.config
	if remoteAddr.User != "" {
		remoteCfg.User = remoteAddr.User
	}

	// Directly connect to the remote host using the underlying remoteConn as the transport.
	clientConn, ncChan, reqChan, err := ssh.NewClientConn(remoteConn, remoteAddr.Addr, &remoteCfg)
	if err != nil {
		jumpConn.Close()
		remoteConn.Close()
		return err
	}

	c.jumpConn = jumpConn
	c.remoteConn = remoteConn
	c.client = ssh.NewClient(clientConn, ncChan, reqChan)
	return nil
}

// Close releases the resources used by the Client.
func (c *Client) Close() error {
	if c.client != nil {
		c.client.Close()
	}
	if c.remoteConn != nil {
		c.remoteConn.Close()
	}
	if c.jumpConn != nil {
		c.jumpConn.Close()
	}
	if c.agentConn != nil {
		c.agentConn.Close()
	}
	return nil
}

// Run executes the given command on the remote host. The returned byte slice is
// the combined stout and stderr output from the command.
func (c *Client) Run(cmd string) ([]byte, error) {
	if c.client == nil {
		return nil, ClientError{"Run: client not connected"}
	}
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return session.CombinedOutput(cmd)
}

// Sudo executes the given command on the remote host as a sudo command. The returned byte slice is
// the combined stout and stderr output from the command.
func (c *Client) Sudo(cmd string) ([]byte, error) {
	if c.client == nil {
		return nil, ClientError{"RunPty: client not connected"}
	}
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm", 80, 24, modes); err != nil {
		return nil, err
	}
	return session.CombinedOutput("sudo " + cmd)
}

// Copy uses the scp protocol to copy the given srcfile from the local host to
// the destFile on the remote host.
// The scp protocol is explain here:
//  	 https://blogs.oracle.com/janp/entry/how_the_scp_protocol_works
//
// The returned byte slice is the combined stout and stderr output from the command.
func (c *Client) Copy(srcFile, destFile string) ([]byte, error) {
	if c.client == nil {
		return nil, ClientError{"Copy: client not connected"}
	}
	session, err := c.client.NewSession()
	if err != nil {
		return nil, err
	}
	defer session.Close()

	src, err := os.Open(srcFile)
	if err != nil {
		return nil, err
	}
	defer src.Close()

	stat, err := src.Stat()
	if err != nil {
		return nil, err
	}
	mode := stat.Mode()
	if !mode.IsRegular() {
		return nil, ClientError{"Copy: File " + srcFile + " is not a regular file"}
	}
	perm := mode.Perm()

	go func() {
		w, _ := session.StdinPipe()
		fmt.Fprintln(w, "C"+fmt.Sprintf("%#o", uint32(perm)), stat.Size(), filepath.Base(destFile))
		if stat.Size() > 0 {
			io.Copy(w, src)
		}
		fmt.Fprint(w, "\x00")
		w.Close()
	}()

	return session.CombinedOutput("scp -qt " + filepath.Dir(destFile))
}
