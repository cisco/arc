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

package spark

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jbogarin/go-cisco-spark/ciscospark"

	"github.com/cisco/arc/pkg/log"
)

type sparkError struct {
	string
}

func (s sparkError) Error() string {
	return s.string
}

type Client struct {
	room        string
	messageType messageType
	sparkClient *ciscospark.Client
}

type messageType int

const (
	Text messageType = iota
	Markdown
	Html
)

func New(token, room string, messageType messageType) (*Client, error) {
	if token == "" {
		return nil, sparkError{"No token available for use"}
	}
	if room == "" {
		return nil, sparkError{"No room to message"}
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	sparkClient := ciscospark.NewClient(client)
	sparkClient.Authorization = "Bearer " + token

	return &Client{
		room:        room,
		messageType: messageType,
		sparkClient: sparkClient,
	}, nil
}

func (c *Client) Get(n int) ([]string, error) {
	m := []string{}
	params := &ciscospark.MessageQueryParams{
		Max:    n,
		RoomID: c.room,
	}
	messages, _, err := c.sparkClient.Messages.Get(params)
	if err != nil {
		return nil, err
	}
	for i := len(messages) - 1; i >= 0; i-- {
		m = append(m, fmt.Sprintf("User: %s\nMessage: %s\n\n", messages[i].PersonEmail, messages[i].Text))
	}
	return m, nil
}

func (c *Client) Write(p []byte) (int, error) {
	if len(p) < 1 {
		return 0, sparkError{"Nothing to send to spark"}
	}
	// Message to send to spark
	s := string(p)
	m := &ciscospark.MessageRequest{
		RoomID: c.room,
	}
	switch c.messageType {
	case Text:
		m.Text = s
	case Markdown, Html:
		m.MarkDown = s
	}
	var err error
	for i := 0; i < 3; i++ {
		_, _, err = c.sparkClient.Messages.Post(m)
		switch err {
		case nil:
			return len(p), nil
		default:
			if strings.Contains(err.Error(), " 502 ") {
				log.Debug("Retrying the spark message")
				time.Sleep(3 * time.Second)
				continue
			}
		}
	}
	return 0, err
}
