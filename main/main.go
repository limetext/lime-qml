// Copyright 2013 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package main

import (
	"flag"
	"runtime"

	"github.com/limetext/backend/log"
	_ "github.com/limetext/commands"
	"github.com/limetext/gopy/lib"
	_ "github.com/limetext/sublime"
)

var rotateLog = flag.Bool("rotateLog", false, "Rotate debug log")

func main() {
	flag.Parse()

	// Need to lock the OS thread as OSX GUI requires GUI stuff to run in the main thread
	runtime.LockOSThread()

	log.AddFilter("file", log.FINEST, log.NewFileLogWriter("debug.log", *rotateLog))
	defer func() {
		py.NewLock()
		py.Finalize()
	}()

	initFrontend()
}
