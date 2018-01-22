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

package env

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"
)

func usage() {
	fmt.Printf("usage: %s name command\n", path.Base(os.Args[0]))
}

var env = map[string]string{}

func Init(appname, version string) error {
	u, err := user.Current()
	if err != nil {
		return err
	}
	searchPath := []string{os.Getenv(strings.ToUpper(appname)), os.Getenv("HOME") + "/." + appname}
	appDir := ""
	for _, d := range searchPath {
		info, err := os.Stat(d)
		if err == nil && info.Mode().IsDir() {
			appDir = d
			break
		}
	}
	if appDir == "" {
		appDir = os.Getenv("HOME" + "/." + appname)
		if err := os.MkdirAll(appDir, 0755); err != nil {
			return err
		}
	}
	latest := appDir + "/latest"

	// Using a modified RFC3339 format to having colons in the path
	t := time.Now()
	appDir += "/" + t.Format("2006-01-02_150405.000")

	if err := os.MkdirAll(appDir, 0755); err != nil {
		return err
	}
	if err := os.Chmod(appDir, 0755); err != nil {
		return err
	}
	if _, err := os.Stat(latest); err == nil {
		os.Remove(latest)
	}
	if err := os.Symlink(appDir, latest); err != nil {
		return err
	}

	err = cleanup(filepath.Dir(appDir))
	if err != nil {
		fmt.Printf("\nWarning: %s\n\n", err.Error())
	}

	root := os.Getenv(strings.ToUpper(appname) + "_ROOT")
	if root == "" {
		info, err := os.Stat("./cmd/" + appname + "/main.go")
		if err == nil && info.Mode().IsRegular() {
			root, err = os.Getwd()
			if err != nil {
				return err
			}
		}
	}
	err = os.Setenv(strings.ToUpper(appname), appDir)
	if err != nil {
		return err
	}

	env[strings.ToUpper(appname)] = appDir
	env["ROOT"] = root
	env["VERSION"] = version
	env["SPARK_TOKEN"] = initSpark()
	env["USER"] = u.Username
	env["SSH_USER"] = u.Username
	if sshUser := os.Getenv("SSH_USER"); sshUser != "" {
		env["SSH_USER"] = sshUser
	}
	return nil
}

func Lookup(k string) string {
	return env[k]
}

func Set(k, v string) {
	env[k] = v
}

func cleanup(dir string) error {
	// Get list of all directories in the arc directory
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	// Get list using only directories staring with a digit (the ones arc produces)
	l := []string{}
	for _, file := range files {
		mode := file.Mode()
		if mode.IsDir() && unicode.IsDigit(rune(file.Name()[0])) {
			l = append(l, dir+"/"+file.Name())
		}
	}

	// Keep 20 latest directories and remove the rest
	sort.Sort(sort.Reverse(sort.StringSlice(l)))
	for i, d := range l {
		if i >= 20 {
			err := os.RemoveAll(d)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func initSpark() string {
	if token := os.Getenv("SPARK_TOKEN"); token != "" {
		return token
	}
	spark := env["ROOT"] + "/etc/arc/spark.gpg"
	cmd := env["ROOT"] + "/usr/local/bin/decrypt_file"
	token, err := exec.Command(cmd, spark).CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(token))
}
