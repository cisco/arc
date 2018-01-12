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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

	"github.com/cisco/arc/pkg/spark"
)

type Spark struct {
	Token string            `json:"token"`
	Rooms map[string]string `json:"rooms"`
}

// New fills the Spark struct with the information provided in the configuration file
func NewSpark() (*Spark, error) {
	file := os.Getenv("HOME") + "/.spark"

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	s := &Spark{}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return s, nil
}

// Run returns a nonzero value when it fails. Run also is where the command is parsed to
// what action to take whether it be sending or getting messages.
func (s *Spark) Run(args []string) error {
	room := s.Find(args[1])
	if room == "" {
		return fmt.Errorf("Room alias not found")
	}
	client, err := spark.New(s.Token, room, spark.Html)
	if err != nil {
		return err
	}
	switch {
	case len(args) == 2:
		// By default we get the 5 most recent messages from the room requested
		message, err := client.Get(5)
		if err != nil {
			return err
		}
		for _, v := range message {
			fmt.Printf(v)
		}
	case len(args) == 3:
		// If the 3 argument is a number flag (i.e. -10) Run prints out n messages from
		// the room requested.
		arg, _ := strconv.Atoi(args[2])
		if arg < 0 {
			message, err := client.Get((-1 * arg))
			if err != nil {
				return err
			}
			for _, v := range message {
				fmt.Printf(v)
			}
			return nil
		}
		// If the 3rd argument is contained in quotes (i.e, "This is a sample message")
		// or if the 3rd argument is a single word (including -0); a message will be sent
		// using the 3rd argument as the message to be sent to the room requested.
		fmt.Fprintf(client, args[2])
	case len(args) > 3:
		// If args is longer than 3 a message will be sent with every argument from args[2]
		// on as a string to the specified room
		message := ""
		for i := 2; i < len(args); i++ {
			message += args[i] + " "
		}
		fmt.Fprintf(client, message)
	}

	return nil
}

// Find checks if a room alias exists, if it doesn't it will return an empty string
func (s *Spark) Find(name string) string {
	return s.Rooms[name]
}

// Help informs how to use the tool
func Help() {
	help := `
  With the Spark Command Line tool you can get and send messages
  To use this tool you must have a config file named .spark in your home directory.
  NOTE:
    *The tokens are a secret so make sure that the file is only readable and writeable by you.
    The tokens are found at http://developer.cisco.com and you click on your icon in the top right
    of the screen and your Access Token will be there

  The format of the file is as follows:
  __________
  | .spark |
  =====================================================================================
      {
        "token": "PUT YOUR SPARK ACCESS TOKEN HERE"
        "rooms": {
          "RoomAlias1": "RoomId which can be found by opening spark in browser",
          "RoomAlias2": "RoomId which can be found by opening spark in browser"
        }
      }
  =====================================================================================

  TOOL USAGE
    The options of how to use the tool are as follows:
  ----------------------------------------------------
  > GET MESSAGES
      5 messages:
        spark roomAlias

      n messages where n is the number of messages you want back:
        spark roomAlias -n

  > SEND MESSAGES
      A message as one argument:
        spark roomAlias "This is the text that gets sent to roomAlias"

      A message over several arguments:
        spark roomAlias This is the text that gets sent to roomAlias
`
	fmt.Println(help)
}
