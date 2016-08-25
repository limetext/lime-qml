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
	bw          *backend.Window
	qw          *qml.Window
	views       map[*backend.View]*view
	SidebarTree *TreeListModel
}

func newWindow(bw *backend.Window) *window {
	return &window{
		bw:    bw,
		views: make(map[*backend.View]*view),
	}
}

// Instantiates a new window, and launches a new goroutine waiting for it
// to be closed. The WaitGroup is increased at function entry and decreased
// once the window closes.
func (w *window) launch(wg *sync.WaitGroup, component qml.Object) {
	wg.Add(1)

	project := w.Back().OpenProject("testproj.sublime-project")

	// me, _ := filepath.Abs(os.Args[0])

	filesHeader := &HeaderItem{name: "files"}
	// root := &FSTreeItem{path: filepath.Dir(me), isDir: true}
	proj := &ProjectItem{Project: project}

	w.SidebarTree = NewTreeListModel(component.Common().Engine(), nil, []TreeListItem{filesHeader, proj})

	w.qw = component.CreateWindow(nil)
	w.qw.Show()
	w.qw.Set("myWindow", w)

	go func() {
		w.qw.Wait()
		wg.Done()
	}()
}

func (w *window) Back() *backend.Window {
	return w.bw
}
