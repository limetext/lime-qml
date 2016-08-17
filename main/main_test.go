// Copyright 2014 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package main

import (
	"testing"
	"time"

	"github.com/limetext/qml-go"
)

func init() { qml.SetupTesting() }

func repeatTest(t func() bool, times int, sleep time.Duration) bool {
	for i := 0; i < times; i++ {
		if t() {
			return true
		}
		time.Sleep(sleep)
	}
	return false
}

func TestFrontend(t *testing.T) {
	didQuit := make(chan struct{})
	go func() {
		initFrontend()
		didQuit <- struct{}{}
	}()

	if !repeatTest(func() bool { return fe != nil && len(fe.windows) == 1 }, 25, 200*time.Millisecond) {
		t.Errorf("Expected exactly 1 window open, have %d", len(fe.windows))
	}

	time.Sleep(500 * time.Millisecond)

	// f := fe // save a copy

	err := fe.Quit()
	if err != nil {
		t.Errorf("Expected to Quit with no errors, received: %s", err)
	}

	select {
	case <-didQuit:
		// good
	case <-time.After(4 * time.Second):
		t.Errorf("Expected initFrontend to return after Quit")
	}

	// fe.Quit() just kills the QML connections, it doesn't actually quit everything
	// if len(f.windows) != 0 {
	// 	t.Errorf("Expected 0 windows open after Quit, have %d", len(f.windows))
	// }
}
