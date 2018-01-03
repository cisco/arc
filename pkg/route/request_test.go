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

package route

import (
	"strings"
	"testing"
)

func check(t *testing.T, req *Request, l int, c Command, f int) {
	p := req.Path()
	if p == nil {
		t.Fatalf("Expected path, got nil\n")
	}
	if len(p.path) != l {
		t.Errorf("Expected path len %d, got %q, %q, %q\n", l, p.path, req.command, req.flags)
	}
	cmd := req.Command()
	if cmd != c {
		t.Errorf("Expected %q, got %q\n", c.String(), cmd.String())
	}
	if len(req.flags.flags) != f {
		t.Errorf("Expects flags len %d, got %q\n", f, req.flags)
	}
}

func TestNewRequest(t *testing.T) {
	req := NewRequest("dc", "user", "time")
	if req == nil {
		t.Fatalf("Expected req, got nil\n")
	}
	dc := req.DataCenter()
	if dc != "dc" {
		t.Errorf("Expected %q, got %q\n", "dc", dc)
	}
	user := req.UserId()
	if user != "user" {
		t.Errorf("Expected %q, got %q\n", "user", user)
	}
	time := req.Time()
	if time != "time" {
		t.Errorf("Expected %q, got %q\n", "time", time)
	}
	if !req.Flags().Empty() {
		t.Errorf("Expects no flags, got %q\n", req.flags)
	}
	p := req.Path()
	if p == nil {
		t.Fatalf("Expected path, got nil\n")
	}
	if len(p.path) > 0 {
		t.Error("Expected empty path, got %q\n", p.path)
	}
	cmd := req.Command()
	if cmd != None {
		t.Errorf("Expected %q, got %q\n", "None", cmd)
	}
	if !req.Flags().Empty() {
		t.Errorf("Expects no flags, got %q\n", req.flags)
	}
	if req.TestFlag() {
		t.Errorf("Test flag shouldn't be set\n", req.flags)
	}
}

func TestCloneRequest(t *testing.T) {
	orig := NewRequest("dc", "user", "time")
	if orig == nil {
		t.Fatalf("Expected req, got nil\n")
	}
	req := orig.Clone(Load)
	if req == nil {
		t.Fatalf("Expected req, got nil\n")
	}
	dc := req.DataCenter()
	if dc != "dc" {
		t.Errorf("Expected %q, got %q\n", "dc", dc)
	}
	user := req.UserId()
	if user != "user" {
		t.Errorf("Expected %q, got %q\n", "user", user)
	}
	time := req.Time()
	if time != "time" {
		t.Errorf("Expected %q, got %q\n", "time", time)
	}
	if !req.Flags().Empty() {
		t.Errorf("Expects no flags, got %q\n", req.flags)
	}
	p := req.Path()
	if p == nil {
		t.Fatalf("Expected path, got nil\n")
	}
	if len(p.path) > 0 {
		t.Error("Expected empty path, got %q\n", p.path)
	}
	cmd := req.Command()
	if cmd != Load {
		t.Errorf("Expected %q, got %q\n", Load.String(), cmd.String())
	}
	if !req.Flags().Empty() {
		t.Errorf("Expects no flags, got %q\n", req.flags)
	}
}

func TestRequestParseEmpty(t *testing.T) {
	req := NewRequest("dc", "user", "time")
	if req == nil {
		t.Fatalf("Expected req, got nil\n")
	}
	req.Parse([]string{""})
	check(t, req, 0, None, 0)
	req.Parse(nil)
	check(t, req, 0, None, 0)
}

func TestRequestParsePathOnly(t *testing.T) {
	req := NewRequest("dc", "user", "time")
	if req == nil {
		t.Fatalf("Expected req, got nil\n")
	}
	req.Parse(strings.Split("foo bar beh", " "))
	check(t, req, 3, None, 0)
}

func TestRequestParseCommandOnly(t *testing.T) {
	req := NewRequest("dc", "user", "time")
	if req == nil {
		t.Fatalf("Expected req, got nil\n")
	}
	req.Parse([]string{Create.String()})
	check(t, req, 0, Create, 0)
}

func TestRequestParseCommandFlags(t *testing.T) {
	req := NewRequest("dc", "user", "time")
	if req == nil {
		t.Fatalf("Expected req, got nil\n")
	}
	req.Parse(strings.Split(Provision.String()+" noreboot bootstrap and other stuff", " "))
	check(t, req, 0, Provision, 5)
}

func TestRequestParse(t *testing.T) {
	req := NewRequest("dc", "user", "time")
	if req == nil {
		t.Fatalf("Expected req, got nil\n")
	}
	req.Parse(strings.Split("secgroup common "+Create.String()+" norules", " "))
	check(t, req, 2, Create, 1)
}
