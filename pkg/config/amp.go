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
	"github.com/cisco/arc/pkg/log"
	"github.com/cisco/arc/pkg/msg"
)

type Amp struct {
	Name_         string         `json:"name"`
	Notifications *Notifications `json:"notifications"`
	Provider      *Provider      `json:"provider"`
	SecurityTags_ SecurityTags   `json:"security_tags"`
	Storage       *Storage       `json:"storage"`
	KeyManagement *KeyManagement `json:"key_management"`
}

func NewAmp(dc string) (*Amp, error) {
	log.Debug("Unmarshal the corresponding json file.")
	file := fmt.Sprintf(env.Lookup("ROOT")+"/etc/arc/%s.json", dc)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	a := &Amp{}
	if err := json.Unmarshal(data, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *Amp) Name() string {
	return a.Name_
}

func (a *Amp) SecurityTags() SecurityTags {
	return a.SecurityTags_
}

func (a *Amp) PrintLocal() {
	msg.Detail("%-20s\t%s", "name", a.Name())
	if a.Provider != nil {
		a.Provider.Print()
	}
	if a.SecurityTags_ != nil {
		a.SecurityTags_.Print()
	}
}

func (a *Amp) Print() {
	msg.Info("Amp Config")
	msg.IndentInc()
	a.PrintLocal()
	if a.Storage != nil {
		a.Storage.Print()
	}
	if a.KeyManagement != nil {
		a.KeyManagement.Print()
	}
	msg.IndentDec()
}
