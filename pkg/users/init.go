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
	"encoding/json"
	"io/ioutil"

	conf "github.com/cisco/arc/pkg/config"
)

var Users map[string]*User
var Groups []*Group
var Teams map[string]*Team
var DataCenters []*conf.Arc

func Init(name string) error {
	file, err := ioutil.ReadFile(name)
	if err != nil {
		return err
	}

	cfg := &config{}
	if err := json.Unmarshal(file, cfg); err != nil {
		return err
	}

	Users = newUsers(cfg.UsersConfig)
	Groups = newGroups(cfg.GroupsConfig)
	Teams, err = newTeams(cfg.TeamsConfig)
	if err != nil {
		return err
	}
	for _, t := range Teams {
		if err := t.Populate(); err != nil {
			return err
		}
	}
	DataCenters, err = newDataCenters(cfg.DataCentersConfig)
	if err != nil {
		return err
	}

	return nil
}
