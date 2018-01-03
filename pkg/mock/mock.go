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

package mock

import (
	"os"

	"github.com/cisco/arc/pkg/config"
	"github.com/cisco/arc/pkg/help"
	"github.com/cisco/arc/pkg/resource"
	"github.com/cisco/arc/pkg/route"
)

//---------------------------------------------------------------------------
// module visibility

var mockedResource map[string]resource.Resource

func init() {
	mockedResource = map[string]resource.Resource{}
}

func set(s string, r resource.Resource) {
	mockedResource[s] = r
}

func Get(s string) resource.Resource {
	return mockedResource[s]
}

//---------------------------------------------------------------------------
// mock - public interface

type mock struct {
	name  string
	state map[string]bool
}

func newMock(name string, cfg *config.Provider) *mock {
	m := &mock{
		name: name,
		state: map[string]bool{
			"route":     true,
			"created":   false,
			"destroyed": true,
		},
	}
	for k, v := range cfg.Data {
		if v == "yes" {
			m.Set(k, true)
		}
		if v == "no" {
			m.Set(k, false)
		}
	}
	return m
}

func (m *mock) Set(s string, b bool) {
	m.state[s] = b
}

func (m *mock) Get(s string) bool {
	return m.state[s]
}

func (m *mock) Route(req *route.Request) route.Response {
	resp := route.FAIL
	if m.Get("route") {
		resp = route.OK
	}
	return resp
}

func (m *mock) Load() error {
	return nil
}

func (m *mock) Created() bool {
	created := os.Getenv("created")
	if created != "" {
		if created == "yes" {
			return true
		} else {
			return false
		}
	}
	return m.Get("created")
}

func (m *mock) Destroyed() bool {
	destroyed := os.Getenv("destroyed")
	if destroyed != "" {
		if destroyed == "yes" {
			return true
		} else {
			return false
		}
	}
	return m.Get("destroyed")
}

func (m *mock) CanRoute(*route.Request) bool {
	return false
}

func (m *mock) HelpCommands() []help.Command {
	return []help.Command{}
}
