// Copyright 2013 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package main

import (
	"sync"

	"github.com/limetext/backend"
	"github.com/limetext/qml-go"
)

// A helper glue structure connecting the backend Window with the qml.Window
type window struct {
	bw    *backend.Window
	views []*view
	qw    *qml.Window
}

// Instantiates a new window, and launches a new goroutine waiting for it
// to be closed. The WaitGroup is increased at function entry and decreased
// once the window closes.
func (w *window) launch(wg *sync.WaitGroup, component qml.Object) {
	wg.Add(1)
	w.qw = component.CreateWindow(nil)
	w.qw.Show()
	w.qw.Set("myWindow", w)

	go func() {
		w.qw.Wait()
		wg.Done()
	}()
}

func (w *window) View(idx int) *view {
	return w.views[idx]
}

func (w *window) ActiveViewIndex() int {
	for i, v := range w.views {
		if v.bv == w.bw.ActiveView() {
			return i
		}
	}
	return 0
}

func (w *window) Back() *backend.Window {
	return w.bw
}

func (w *window) findView(bv *backend.View) (*view, int) {
	for i, v := range w.views {
		if v.bv == bv {
			return v, i
		}
	}
	return nil, -1
}
