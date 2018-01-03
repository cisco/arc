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

type Flags struct {
	flags []string
}

func NewFlags() *Flags {
	return &Flags{
		flags: []string{},
	}
}

func (f *Flags) Clone() *Flags {
	return &Flags{f.flags}
}

func (f *Flags) isSet(s string) bool {
	for _, t := range f.flags {
		if s == t {
			return true
		}
	}
	return false
}

func (f *Flags) Set(s []string) {
	f.flags = s
}

func (f *Flags) Get() []string {
	return f.flags
}

func (f *Flags) Append(s string) *Flags {
	if !f.isSet(s) {
		f.flags = append(f.flags, s)
	}
	return f
}

func (f *Flags) Remove(s string) *Flags {
	newflags := []string{}
	for _, t := range f.flags {
		if s != t {
			newflags = append(newflags, t)
		}
	}
	f.flags = newflags
	return f
}

func (f *Flags) Empty() bool {
	return len(f.flags) == 0
}
