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

package log

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/cisco/arc/pkg/env"
	"github.com/cisco/arc/pkg/route"
)

type logger struct {
	lock          *sync.Mutex
	file          *os.File
	writer        *bufio.Writer
	enableVerbose bool
	verbose       *log.Logger
	enableDebug   bool
	debug         *log.Logger
	enableInfo    bool
	info          *log.Logger
	enableWarn    bool
	warn          *log.Logger
	enableError   bool
	err           *log.Logger
}

var l *logger

func Init(appname string) error {
	if l != nil {
		return nil
	}
	fileName := fmt.Sprintf(env.Lookup(strings.ToUpper(appname)) + "/" + appname + ".log")
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	if err := os.Chmod(fileName, 0644); err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	l = &logger{
		lock:          &sync.Mutex{},
		file:          file,
		writer:        writer,
		enableVerbose: true,
		verbose:       log.New(writer, "Verbose| ", log.Ldate|log.Ltime),
		enableDebug:   true,
		debug:         log.New(writer, "Debug  | ", log.Ldate|log.Ltime),
		enableInfo:    true,
		info:          log.New(writer, "Info   | ", log.Ldate|log.Ltime),
		enableWarn:    true,
		warn:          log.New(writer, "Warn   | ", log.Ldate|log.Ltime),
		enableError:   true,
		err:           log.New(writer, "Error  | ", log.Ldate|log.Ltime),
	}
	if os.Getenv("debug") == "no" {
		l.enableDebug = false
	}
	if os.Getenv("verbose") == "no" {
		l.enableVerbose = false
	}
	Info("%s %s", appname, env.Lookup("VERSION"))
	return nil
}

func Fini() {
	if l == nil {
		return
	}
	l.file.Close()
	l = nil
}

func Verbose(format string, a ...interface{}) {
	if l.enableVerbose == true {
		logMsg(l.verbose, format, a...)
	}
}

func Debug(format string, a ...interface{}) {
	if l.enableDebug == true {
		logMsg(l.debug, format, a...)
	}
}

func Info(format string, a ...interface{}) {
	if l.enableInfo == true {
		logMsg(l.info, format, a...)
	}
}

func Warn(format string, a ...interface{}) {
	if l.enableWarn == true {
		logMsg(l.warn, format, a...)
	}
}

func Error(format string, a ...interface{}) {
	if l.enableError == true {
		logMsg(l.err, format, a...)
	}
}

func logMsg(lg *log.Logger, format string, a ...interface{}) {
	l.lock.Lock()
	defer l.lock.Unlock()

	// Add caller's module short name and line number to log message
	_, fn, line, _ := runtime.Caller(2)
	fileName := path.Base(fn)
	if fileName == "msg.go" {
		_, fn, line, _ = runtime.Caller(3)
		fileName = path.Base(fn)
	}
	args := []interface{}{fileName, line}
	args = append(args, a...)
	lg.Println(fmt.Sprintf("| %22s +%-3d | "+format, args...))
	l.writer.Flush()
}

func Route(req *route.Request, format string, a ...interface{}) {
	name := fmt.Sprintf(format, a...)
	if l.enableDebug == true {
		logMsg(l.debug, "%s routing request: %q", name, req)
	}
}
