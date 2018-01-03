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

import "testing"

func TestNewPath(t *testing.T) {
	p := NewPath()
	if p == nil {
		t.Fatalf("Failed to create new Path\n")
	}
	if len(p.path) != 0 {
		t.Errorf("Expected empty path, got %p\n", p.path)
	}
}

func TestPathPop(t *testing.T) {
	p := NewPath()
	if p == nil {
		t.Fatalf("Failed to create new Path\n")
	}

	s := p.Pop().Top()
	if s != "" {
		t.Errorf("Expected empty string, got %q\n", s)
	}

	s = p.Push("one").Pop().Top()
	if len(p.path) != 0 {
		t.Errorf("Expected path of length 0, got %q\n", p.path)
	}

	s = p.Push("one").Push("two").Push("three").Push("four").Pop().Top()
	if len(p.path) != 3 {
		t.Errorf("Expected path of length 4, got %q\n", p.path)
	}
	if s != "three" {
		t.Errorf("Expected top to be %q, got %q\n", "three", s)
	}
}

func TestPathPush(t *testing.T) {
	p := NewPath()
	if p == nil {
		t.Fatalf("Failed to create new Path\n")
	}

	p.Push("one")
	if len(p.path) != 1 {
		t.Errorf("Expected path of length 1, got %q\n", p.path)
	}

	s := p.Top()
	if s != "one" {
		t.Errorf("Expected top to be %q, got %q\n", "one", s)
	}

	p.Push("two").Push("three").Push("four")

	if len(p.path) != 4 {
		t.Errorf("Expected path of length 4, got %q\n", p.path)
	}

	s = p.Top()
	if s != "four" {
		t.Errorf("Expected top to be %q, got %q\n", "four", s)
	}
}

func TestPathAppend(t *testing.T) {
	p := NewPath()
	if p == nil {
		t.Fatalf("Failed to create new Path\n")
	}

	p.Append("one")
	if len(p.path) != 1 {
		t.Errorf("Expected path of length 1, got %q\n", p.path)
	}

	s := p.Top()
	if s != "one" {
		t.Errorf("Expected top to be %q, got %q\n", "one", s)
	}

	p.Append("two").Append("three").Append("four")

	if len(p.path) != 4 {
		t.Errorf("Expected path of length 4, got %q\n", p.path)
	}

	s = p.Top()
	if s != "one" {
		t.Errorf("Expected top to be %q, got %q\n", "one", s)
	}
}
