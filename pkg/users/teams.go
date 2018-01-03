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
	"strings"
)

// Team provides a way to organize users.
type Team struct {
	Name     string
	Sudo     bool
	Users    []*User
	subTeams []string
}

func newTeam(cfg teamConfig) (*Team, error) {
	t := &Team{
		Name:     cfg.Name,
		Sudo:     cfg.Sudo,
		Users:    []*User{},
		subTeams: []string{},
	}
	for _, n := range cfg.Users {
		user := strings.Split(n, ":")
		if len(user) == 2 && user[0] == "team" {
			if s := t.FindTeam(user[1]); s != "" {
				continue
			}
			t.subTeams = append(t.subTeams, user[1])
			continue
		}
		u := Users[n]
		if u == nil {
			return nil, usersError{"User " + n + " is not defined."}
		}
		t.Users = append(t.Users, u)
	}
	return t, nil
}

func newTeams(cfg teamsConfig) (map[string]*Team, error) {
	m := map[string]*Team{}
	for _, c := range cfg {
		t, err := newTeam(c)
		if err != nil {
			return nil, err
		}
		m[t.Name] = t
	}
	return m, nil
}

func (t *Team) FindUser(name string) *User {
	for _, u := range t.Users {
		if u.Name == name {
			return u
		}
	}
	return nil
}

func (t *Team) FindTeam(name string) string {
	for _, team := range t.subTeams {
		if team == name {
			return team
		}
	}
	return ""
}

func (t *Team) Populate() error {
	for i := 0; i < len(t.subTeams); i++ {
		s := t.subTeams[i]
		team := Teams[s]
		if team == nil {
			return usersError{"Team " + s + " is not defined."}
		}
		for _, v := range team.subTeams {
			if st := t.FindTeam(v); st == "" {
				t.subTeams = append(t.subTeams, v)
			}
		}
		for _, u := range team.Users {
			if user := t.FindUser(u.Name); user != nil {
				continue
			}
			t.Users = append(t.Users, u)
		}
	}
	return nil
}
