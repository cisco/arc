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

package ssh

import (
	"testing"
)

func TestAddressEmpty(t *testing.T) {
	addr := NewAddress("", "")

	if addr.User != "" {
		t.Error("Expected empty user, got ", addr.User)
	}
	if addr.Addr != ":22" {
		t.Error("Expected addr ':22', got ", addr.Addr)
	}
}

func TestAddressNoPort(t *testing.T) {
	addr := NewAddress("alice", "10.0.0.4")

	if addr.User != "alice" {
		t.Error("Expected user 'alice' user, got ", addr.User)
	}
	if addr.Addr != "10.0.0.4:22" {
		t.Error("Expected addr ':22', got ", addr.Addr)
	}
}

func TestFullAddress(t *testing.T) {
	addr := NewAddress("alice", "10.0.0.4:2222")

	if addr.User != "alice" {
		t.Error("Expected user 'alice' user, got ", addr.User)
	}
	if addr.Addr != "10.0.0.4:2222" {
		t.Error("Expected addr '10.0.0.4:2222', got ", addr.Addr)
	}
}
