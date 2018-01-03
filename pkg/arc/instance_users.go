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
	"os"
	"strconv"

	"github.com/cisco/arc/pkg/command"
	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/route"
	"github.com/cisco/arc/pkg/users"
)

func (i *Instance) configureUsers(req *route.Request, asRoot bool) route.Response {
	var err error
	commands := []command.Command{}

	commands = i.copyUserScripts(commands)
	commands = i.setupGroups(commands)
	commands, err = i.setupUsers(commands)

	if err != nil {
		msg.Error(err.Error())
		return route.FAIL
	}

	f := command.RunQuiet
	if asRoot {
		f = command.RunQuietAsRoot
	}
	if !f(commands, i) {
		return route.FAIL
	}
	return route.OK
}

func (i *Instance) copyUserScripts(commands []command.Command) []command.Command {
	commands = append(commands,
		command.Command{
			Type: command.Message,
			Dest: "Info",
			Desc: "Installing user scripts",
		},
	)
	dir := "/usr/lib/arc/users/"
	scripts := []string{"setup_group", "setup_user", "add_user_to_groups", "setup_ssh", "setup_sudo"}
	for _, script := range scripts {
		commands = append(commands,
			command.Command{
				Type: command.Copy,
				Desc: script,
				Src:  dir + script,
			},
			command.Command{
				Type: command.Message,
				Desc: "Script installed: " + script,
			},
		)
	}
	return commands
}

func (i *Instance) setupGroups(commands []command.Command) []command.Command {
	commands = append(commands,
		command.Command{
			Type: command.Message,
			Dest: "Info",
			Desc: "Creating groups",
		},
	)
	for _, group := range users.Groups {
		c := command.Command{
			Type: command.Sudo,
			Desc: "setup group \"" + group.Name + "\"",
			Src:  "/usr/lib/arc/users/setup_group",
			Args: []string{group.Name, strconv.Itoa(group.Gid)},
		}
		if group.Remove {
			c.Desc = "remove group \"" + group.Name + "\""
			c.Args = append(c.Args, "remove")
		}
		commands = append(commands, c,
			command.Command{
				Type: command.Message,
				Desc: "Created group: " + group.Name,
			},
		)
	}
	return commands
}

func (i *Instance) setupUsers(commands []command.Command) ([]command.Command, error) {
	for _, name := range i.Teams() {
		team := users.Teams[name]
		if team == nil {
			return commands, fmt.Errorf("Unknown Team %1", name)
		}
		commands = append(commands,
			command.Command{
				Type: command.Message,
				Dest: "Info",
				Desc: "Creating team: " + name,
			},
		)
		for _, user := range team.Users {
			if user.Remove {
				commands = i.removeUser(user, commands)
			} else {
				var err error
				commands, err = i.addUser(user, commands, team.Sudo)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return commands, nil
}

func (i *Instance) addUser(user *users.User, commands []command.Command, sudo bool) ([]command.Command, error) {
	groupArgs := []string{user.Name}
	groupArgs = append(groupArgs, user.Groups...)

	authFile := env.Lookup("ARC") + "/" + "authorized_keys." + user.Name
	file, err := os.OpenFile(authFile, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	for _, key := range user.SshKeys {
		if _, err := file.WriteString(key + "\n"); err != nil {
			file.Close()
			return nil, err
		}
	}
	if err := file.Close(); err != nil {
		return nil, err
	}

	commands = append(commands,
		command.Command{
			Type: command.Sudo,
			Desc: "create user \"" + user.Name + "\" account",
			Src:  "/usr/lib/arc/users/setup_user",
			Args: []string{user.Name, strconv.Itoa(user.Uid)},
		},
		command.Command{
			Type: command.Sudo,
			Desc: "add user \"" + user.Name + "\" to groups",
			Src:  "/usr/lib/arc/users/add_user_to_groups",
			Args: groupArgs,
		},
		command.Command{
			Type: command.Copy,
			Desc: "authorized_keys",
			Src:  authFile,
			Dest: "/tmp/authorized_keys." + user.Name,
		},
		command.Command{
			Type: command.Sudo,
			Desc: "setup user \"" + user.Name + "\" ssh",
			Src:  "/usr/lib/arc/users/setup_ssh",
			Args: []string{user.Name},
		},
	)
	if sudo {
		c := command.Command{
			Type: command.Sudo,
			Desc: "add sudo access for  user \"" + user.Name + "\"",
			Src:  "/usr/lib/arc/users/setup_sudo",
			Args: []string{user.Name},
		}
		commands = append(commands, c)
	}
	commands = append(commands,
		command.Command{
			Type: command.Message,
			Desc: "Created user: " + user.Name,
		},
	)
	return commands, nil
}

func (i *Instance) removeUser(user *users.User, commands []command.Command) []command.Command {
	commands = append(commands,
		command.Command{
			Type: command.Sudo,
			Desc: "remove user \"" + user.Name + "\"",
			Src:  "/usr/lib/arc/users/setup_user",
			Args: []string{user.Name, "remove"},
		},
		command.Command{
			Type: command.Sudo,
			Desc: "remove sudo access for  user \"" + user.Name + "\"",
			Src:  "/usr/lib/arc/users/setup_sudo",
			Args: []string{user.Name, "remove"},
		},
		command.Command{
			Type: command.Message,
			Desc: "Removed user: " + user.Name,
		},
	)
	return commands
}
