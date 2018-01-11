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

package main

import (
	"fmt"

	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/msg"
	"github.com/cisco/arc/pkg/users"
)

type Audit struct {
	orphanedUsers []*users.User
	orphanedTeams []*users.Team
	rogueTeams    []*users.TeamAudit
}

func NewAudit() *Audit {
	return &Audit{}
}

func (a *Audit) Run(arg string) error {
	switch arg {
	case "users":
		var err error
		err = aaa.NewAudit("User")
		if err != nil {
			return err
		}
		// Users Auditing
		a.orphanedUsers, err = users.FindOrphanedUsers()
		if err != nil {
			return nil
		}
		if a.orphanedUsers != nil {
			for _, v := range a.orphanedUsers {
				if v.Remove {
					msg.Detail("- user: %s, status: removed", v.Name)
					aaa.AuditBuffer["User"].FreeFormAudit("- user: %s, status: removed\n", v.Name)
				} else {
					msg.Detail("- user: %s", v.Name)
					aaa.AuditBuffer["User"].FreeFormAudit("- user: %s\n", v.Name)
				}
			}
		}

		// Teams Auditing
		if err != nil {
			return err
		}
		a.orphanedTeams, err = users.FindOrphanedTeams()
		if err != nil {
			return err
		}
		if a.orphanedTeams != nil {
			for _, v := range a.orphanedTeams {
				msg.Detail("- team: %s", v.Name)
				aaa.AuditBuffer["User"].FreeFormAudit("- team: %s\n", v.Name)
			}
		}
		a.rogueTeams, err = users.FindRogueTeams()
		if err != nil {
			return err
		}
		if a.rogueTeams != nil {
			for _, v := range a.rogueTeams {
				msg.Detail("- team: %s, pod: %s", v.RogueTeam, v.RogueTeamPod)
				aaa.AuditBuffer["User"].FreeFormAudit("- team: %s, pod: %s\n", v.RogueTeam, v.RogueTeamPod)
			}
		}
	}
	return nil
}

func Help() {
	help := `
  With the Audit Command Line tool you can audit the users for all the datacenters specified in 
    /etc/arc/users.json
  The tool determines if there are any orphaned users, rogue users, orphaned teams, or rogue teams.
	The rogue users are caught in pkg/users/teams.go during the newTeam and specifies that a User is undefined.

  TOOL USAGE
    The options of how to use the tool are as follows:
  ----------------------------------------------------
  > Auditing Users
      audit users
`
	fmt.Println(help)
}
