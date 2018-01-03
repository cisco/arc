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
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/resource"
)

//---------------------------------------------------------------------------

type cType uint

const (
	// Run a script on the local machine.
	Local = iota

	// Remote is a command that combines a copy and a sudo command.
	Remote

	// Copy a file from the local machine to the target instance.
	Copy

	// Run a script on the target machine as with super user priviledges.
	Sudo

	// Output a message to the console
	Message
)

// The Command structure.
type Command struct {
	// Type is required in a command set and not needed it an individual run command.
	// These are Local, Remote, Copy and Sudo.
	Type cType

	// Instance is required in an individual run command and not needed in a command set.
	Instance resource.Instance

	// The description of the command. This will be send as a message to the console.
	Desc string

	// The source of the command. For local and sudo type of commands, this is the
	// path to the command to be executed. For copy and remote commands this is the
	// path of the script to copy to the target instance.
	Src string

	// The destination of the command. This is unneeded for local and sudo commands.
	// For remote and copy commands, this is the location on the target instance
	// where the script will be copied to
	Dest string

	// The arguments passed to the command. This is unneeded for the copy command
	Args []string

	asRoot bool
}

//---------------------------------------------------------------------------

// Run runs a set of commands. Remote and copy commands are directed to
// the given instance. If the instance is nil, the set of commands should
// all be local. It returns true on success. Command output is send as messages
// to the console.
func Run(c []Command, i resource.Instance) bool {
	return runCommands(c, i, false)
}

// RunQuiet runs a set of commands the same as Run, however messaging
// to the console is disabled for the run.
func RunQuiet(c []Command, i resource.Instance) bool {
	msg.Quiet(true)
	defer msg.Quiet(false)
	return runCommands(c, i, false)
}

// RunAsRoot runs a set of commands as root. Remote and copy commands are directed to
// the given instance. If the instance is nil, the set of commands should
// all be local. It returns true on success. Command output is send as messages
// to the console.
func RunAsRoot(c []Command, i resource.Instance) bool {
	return runCommands(c, i, true)
}

// RunQuietAsRoot runs a set of commands the same as RunAsRoot, however messaging
// to the console is disabled for the run.
func RunQuietAsRoot(c []Command, i resource.Instance) bool {
	msg.Quiet(true)
	defer msg.Quiet(false)
	return runCommands(c, i, true)
}

// RunWithOutput runs a set of commands. Remote and copy commands are directed to
// the given instance. If the instance is nil, the set of commands should
// all be local. The command output and an error are returned.
func RunWithOutput(c []Command, i resource.Instance) ([]byte, error) {
	return runCommandsWithOutput(c, i, false)
}

// RunWithOutputAsRoot runs a set of commands as root. Remote and copy commands are directed to
// the given instance. If the instance is nil, the set of commands should
// all be local. The command output and an error are returned.
func RunWithOutputAsRoot(c []Command, i resource.Instance) ([]byte, error) {
	return runCommandsWithOutput(c, i, true)
}

//---------------------------------------------------------------------------

// RunLocalWithOutput runs the given command as a local command.
// The command output and an error are returned.
func RunLocalWithOutput(c Command) ([]byte, error) {
	return runLocal(c)
}

// RunLocal runs the given command as a local command.
// It returns true on success. Command output is send as messages
// to the console.
func RunLocal(c Command) bool {
	return runCommand(c, RunLocalWithOutput)
}

//---------------------------------------------------------------------------

// RunRemoteWithOutput copies the source script from the given command to
// the destination on the instance. It then runs the script on the instance.
// The command output and an error are returned.
func RunRemoteWithOutput(c Command) ([]byte, error) {
	return runCommandWithOutput(c, runRemote)
}

// RunRemote copies the source script from the given command to the destination
// on the instance. It then runs the script on the instance. It returns true on
// success. Command output is send as messages to the console.
func RunRemote(c Command) bool {
	return runCommand(c, RunRemoteWithOutput)
}

//---------------------------------------------------------------------------

// RunSudoWithOutput runs the given command on the instance under sudo.
// The command output and an error are returned.
func RunSudoWithOutput(c Command) ([]byte, error) {
	return runCommandWithOutput(c, runSudo)
}

// RunSudo runs the given command on the instance under sudo.
// It returns true on success. Command output is send as messages
// to the console.
func RunSudo(c Command) bool {
	return runCommand(c, RunSudoWithOutput)
}

//---------------------------------------------------------------------------

// CopyToWithOutput copies the give source script to the destination on the
// given instance. The command output and an error are returned.
func CopyToWithOutput(c Command) ([]byte, error) {
	return runCommandWithOutput(c, copyTo)
}

// CopyTo copies the give source script to the destination on the
// given instance. It returns true on success. Command output is send as
// messages to the console.
func CopyTo(c Command) bool {
	return runCommand(c, CopyToWithOutput)
}

//---------------------------------------------------------------------------

func runCommands(c []Command, i resource.Instance, r bool) bool {
	if output, err := runCommandsWithOutput(c, i, r); err != nil {
		if output == nil {
			msg.Error(err.Error())
		} else {
			msg.Error("%s\n%s", err.Error(), output)
		}
		return false
	}
	return true
}

func runCommandsWithOutput(c []Command, i resource.Instance, r bool) ([]byte, error) {
	var cl *client
	var err error
	if i != nil {
		cl, err = newClient(i, r)
		if err != nil {
			return nil, err
		}
		defer cl.close()
	}

	for _, command := range c {
		command.Instance = i
		command.asRoot = r
		if output, err := commandRouter(command, cl); err != nil {
			return output, err
		}
	}
	return nil, nil
}

func commandRouter(c Command, cl *client) ([]byte, error) {
	switch c.Type {
	case Local:
		return runLocal(c)
	case Remote:
		if c.Instance == nil {
			return nil, fmt.Errorf("The instance must be defined for a remote command")
		}
		return runRemote(c, cl)
	case Sudo:
		if c.Instance == nil {
			return nil, fmt.Errorf("The instance must be defined for a sudo command")
		}
		return runSudo(c, cl)
	case Copy:
		if c.Instance == nil {
			return nil, fmt.Errorf("The instance must be defined for a copy command")
		}
		return copyTo(c, cl)
	case Message:
		if msg.GetQuiet() {
			msg.Quiet(false)
			defer msg.Quiet(true)
		}
		switch c.Dest {
		case "Error":
			msg.Error(c.Desc)
		case "Warn":
			msg.Warn(c.Desc)
		case "Info":
			msg.Info(c.Desc)
		case "Detail":
			msg.Detail(c.Desc)
		case "Raw":
			msg.Raw(c.Desc)
		case "":
			msg.Detail(c.Desc)
		}
		return nil, nil
	}
	return nil, fmt.Errorf("Unknown target command target %d", c.Type)
}

//---------------------------------------------------------------------------

type commandFunc func(Command) ([]byte, error)
type commandSshFunc func(Command, *client) ([]byte, error)

func runCommandWithOutput(c Command, f commandSshFunc) ([]byte, error) {
	cl, err := newClient(c.Instance, c.asRoot)
	if err != nil {
		return nil, err
	}
	defer cl.close()
	return f(c, cl)
}

func runCommand(c Command, f commandFunc) bool {
	if output, err := f(c); err != nil {
		if output == nil {
			msg.Error(err.Error())
		} else {
			msg.Error("%s\n%s", err.Error(), output)
		}
		return false
	}
	return true
}

//---------------------------------------------------------------------------

func (c Command) arcSrc() bool {
	return strings.HasPrefix(c.Src, "/usr/lib/arc")
}

func runLocal(c Command) ([]byte, error) {
	msg.Info("Local command: %s", c.Desc)
	output, err := local(c)
	if err != nil {
		return output, err
	}
	return nil, nil
}

func runRemote(c Command, cl *client) ([]byte, error) {
	msg.Info("Remote command: %s on %s", c.Desc, c.Instance.Name())
	output, err := copyto(c, cl)
	if err != nil {
		return output, err
	}
	d := c
	if d.Dest != "" {
		d.Src = d.Dest
	}
	output, err = sudo(d, cl)
	if err != nil {
		return output, err
	}
	return nil, nil
}

func runSudo(c Command, cl *client) ([]byte, error) {
	msg.Info("Sudo command: %s on %s", c.Desc, c.Instance.Name())
	output, err := sudo(c, cl)
	if err != nil {
		return output, err
	}
	return nil, nil
}

func copyTo(c Command, cl *client) ([]byte, error) {
	msg.Info("Copy command: %s to %s", c.Desc, c.Instance.Name())
	output, err := copyto(c, cl)
	if err != nil {
		return output, err
	}
	return nil, nil
}

//---------------------------------------------------------------------------

func local(c Command) ([]byte, error) {
	cmd := ""
	if c.arcSrc() {
		cmd += env.Lookup("ROOT")
	}
	cmd += c.Src
	args := ""
	for _, arg := range c.Args {
		if args == "" {
			args = arg
		} else {
			args += " " + arg
		}
	}
	msg.Detail("Running command '%s %s'", cmd, args)
	output, err := exec.Command(cmd, c.Args...).CombinedOutput()
	log.Verbose("%s", output)
	return output, err
}

func sudo(c Command, cl *client) ([]byte, error) {
	cmd := c.Src
	if len(c.Args) > 0 {
		args := ""
		for _, a := range c.Args {
			args += " " + a
		}
		cmd += args
	}
	output, err := cl.sudo(cmd)
	log.Verbose("%s", output)
	return output, err
}

func copyto(c Command, cl *client) ([]byte, error) {
	// Let's make sure the directory we are copying the file to exists
	// before the copy.
	dir := filepath.Dir(c.Dest)
	if dir != "." && dir != "/tmp" {
		quiet := msg.GetQuiet()
		if !quiet {
			msg.Quiet(true)
		}
		output, err := sudo(Command{Src: "/bin/mkdir -p " + dir}, cl)
		if !quiet {
			msg.Quiet(false)
		}
		if err != nil {
			return output, err
		}
	}

	if c.Dest == "" {
		c.Dest = c.Src
	}
	src := ""
	if c.arcSrc() {
		src += env.Lookup("ROOT")
	}
	src += c.Src
	output, err := cl.copy(src, c.Dest)
	log.Verbose("%s", output)
	return output, err
}
