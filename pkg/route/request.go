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

import "fmt"

type Request struct {
	datacenter string
	userId     string
	time       string

	path    *Path
	command Command
	flags   *Flags
}

func NewRequest(datacenter string, userId string, time string) *Request {
	return &Request{
		datacenter: datacenter,
		userId:     userId,
		time:       time,
		path:       NewPath(),
		command:    None,
		flags:      NewFlags(),
	}
}

func (r *Request) Clone(c Command) *Request {
	return &Request{
		datacenter: r.datacenter,
		userId:     r.userId,
		time:       r.time,
		path:       NewPath(),
		command:    c,
		flags:      r.Flags().Clone(),
	}
}

func (r *Request) Parse(params []string) {
	for i, s := range params {
		if s == "" {
			continue
		}
		c := s2c[s]
		switch c {
		case None:
			r.path.Append(s)
		default:
			r.command = c
			r.flags.Set(params[i+1:])
			return
		}
	}
}

func (r *Request) DataCenter() string {
	return r.datacenter
}

func (r *Request) UserId() string {
	return r.userId
}

func (r *Request) Time() string {
	return r.time
}

func (r *Request) Path() *Path {
	return r.path
}

func (r *Request) Command() Command {
	return r.command
}

func (r *Request) SetCommand(c Command) {
	r.command = c
}

func (r *Request) Flags() *Flags {
	return r.flags
}

func (r *Request) Flag(s string) bool {
	return r.flags.isSet(s)
}

func (r *Request) TestFlag() bool {
	return r.flags.isSet("test")
}

func (r *Request) String() string {
	return fmt.Sprintf("%s %s %s", r.path.path, r.command, r.flags.flags)
}

func (r *Request) Top() string {
	return r.path.Top()
}

func (r *Request) Pop() *Request {
	r.path.Pop()
	return r
}
