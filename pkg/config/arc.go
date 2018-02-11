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

package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/msg"
)

// The configuration of the arc object. It has a name, a
// datacenter element and a dns element.
type Arc struct {
	Name_           string           `json:"name"`
	Title_          string           `json:"title"`
	Provider        *Provider        `json:"provider"`
	Notifications   *Notifications   `json:"notifications"`
	DataCenter      *DataCenter      `json:"datacenter"`
	DatabaseService *DatabaseService `json:"database_service"`
	Dns             *Dns             `json:"dns"`
}

func NewArc(dc string) (*Arc, error) {
	file := fmt.Sprintf(env.Lookup("ROOT")+"/etc/arc/%s.json", dc)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	a := &Arc{}
	if err := json.Unmarshal(data, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Name satisfies the resource.StaticArc interface.
func (a *Arc) Name() string {
	return a.Name_
}

// Title satisfies the resource.StaticArc interface.
func (a *Arc) Title() string {
	return a.Title_
}

// PrintLocal provides a user friendly way to view the configuration local to the arc object.
// This is a shallow print.
func (a *Arc) PrintLocal() {
	msg.Info("Arc Config")
	msg.Detail("%-20s\t%s", "name", a.Name())
	msg.Detail("%-20s\t%s", "title", a.Title())
}

// Print provides a user friendly way to view the arc configuration.
// This is a deep print.
func (a *Arc) Print() {
	a.PrintLocal()
	msg.IndentInc()
	if a.Provider != nil {
		a.Provider.Print()
	}
	if a.DataCenter != nil {
		a.DataCenter.Print()
	}
	if a.DatabaseService != nil {
		a.DatabaseService.Print()
	}
	if a.Dns != nil {
		a.Dns.Print()
	}
	msg.IndentDec()
}
