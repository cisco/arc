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

package users

import (
	"github.com/cisco/arc/pkg/aaa"
	"github.com/cisco/arc/pkg/msg"
)

type TeamAudit struct {
	RogueTeam    string
	RogueTeamPod string
}

// FindOrphanedUsers finds any users that are not a part of a team.
func FindOrphanedUsers() ([]*User, error) {
	msg.Info("Find Orphaned Users")
	aaa.AuditBuffer["Users"].FreeFormAudit("Find Orphaned Users\n")

	found := false
	var orphanedUsers []*User
	for _, usr := range Users {
		found = false
		for _, t := range Teams {
			for _, u := range t.Users {
				if usr.Name == u.Name {
					found = true
				}
			}
		}
		if !found {
			orphanedUsers = append(orphanedUsers, usr)
		}
	}
	return orphanedUsers, nil
}

// FindOrphanedTeams finds any teams that are not a part of a pod
func FindOrphanedTeams() ([]*Team, error) {
	msg.Info("Find Orphaned Teams")
	aaa.AuditBuffer["Users"].FreeFormAudit("\n> Find Orphaned Teams\n")

	var orphanedTeams []*Team
	var foundTeams map[string]bool
	foundTeams = make(map[string]bool)

	for _, team := range Teams {
		foundTeams[team.Name] = false
	}

	for _, team := range Teams {
		for _, dc := range DataCenters {
			for _, cluster := range *dc.DataCenter.Compute.Clusters {
				for _, pod := range *cluster.Pods {
					for _, t := range pod.Teams_ {
						if team.Name == t {
							foundTeams[team.Name] = true
							for _, v := range team.subTeams {
								foundTeams[v] = true
							}
						}
					}
				}
			}
		}
	}
	for team, found := range foundTeams {
		if !found {
			orphanedTeams = append(orphanedTeams, Teams[team])
		}
	}

	return orphanedTeams, nil
}

// FindRogueTeams finds any teams that are a part of a pod, but aren't a defined team.
func FindRogueTeams() ([]*TeamAudit, error) {
	msg.Info("Find Rogue Teams")
	aaa.AuditBuffer["Users"].FreeFormAudit("\n> Find Rogue Teams\n")

	var rogueTeams []*TeamAudit
	for _, dc := range DataCenters {
		for _, cluster := range *dc.DataCenter.Compute.Clusters {
			for _, pod := range *cluster.Pods {
				for _, t := range pod.Teams_ {
					if team := Teams[t]; team == nil {
						rogueTeams = append(rogueTeams, &TeamAudit{
							RogueTeam:    t,
							RogueTeamPod: pod.Name(),
						})
					}
				}
			}
		}
	}

	return rogueTeams, nil
}
