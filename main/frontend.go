// Copyright 2013 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image/color"
	"runtime"
	"strings"
	"sync"
	"time"

	"gopkg.in/fsnotify.v1"
	"gopkg.in/qml.v1"

	"github.com/limetext/backend"
	"github.com/limetext/backend/keys"
	"github.com/limetext/backend/log"
	"github.com/limetext/backend/render"
	. "github.com/limetext/text"
	"github.com/limetext/util"
)

const (
	batching_enabled = true
	qmlWindowFile    = "qml/Window.qml"
	qmlViewFile      = "qml/View.qml"

	// http://qt-project.org/doc/qt-5.1/qtcore/qt.html#KeyboardModifier-enum
	shift_mod  = 0x02000000
	ctrl_mod   = 0x04000000
	alt_mod    = 0x08000000
	meta_mod   = 0x10000000
	keypad_mod = 0x20000000
)

// keeping track of frontend state
type qmlfrontend struct {
	status_message string
	lock           sync.Mutex
	windows        map[*backend.Window]*frontendWindow
	Console        *frontendView
	qmlDispatch    chan qmlDispatch
}

// Used for batching qml.Changed calls
type qmlDispatch struct{ value, field interface{} }

func (t *qmlfrontend) Window(w *backend.Window) *frontendWindow {
	return t.windows[w]
}

func (t *qmlfrontend) Show(v *backend.View, r Region) {
	// TODO
}

func (t *qmlfrontend) VisibleRegion(v *backend.View) Region {
	// TODO
	return Region{0, v.Size()}
}

func (t *qmlfrontend) StatusMessage(msg string) {
	t.lock.Lock()
	defer t.lock.Unlock()
	t.status_message = msg
}

func (t *qmlfrontend) ErrorMessage(msg string) {
	log.Error(msg)
	var q qmlDialog
	q.Show(msg, "StandardIcon.Critical")
}

func (t *qmlfrontend) MessageDialog(msg string) {
	var q qmlDialog
	q.Show(msg, "StandardIcon.Information")
}

func (t *qmlfrontend) OkCancelDialog(msg, ok string) bool {
	var q qmlDialog
	return q.Show(msg, "StandardIcon.Question") == 1
}

func (t *qmlfrontend) scroll(b Buffer) {
	t.Show(backend.GetEditor().Console(), Region{b.Size(), b.Size()})
}

func (t *qmlfrontend) Erased(changed_buffer Buffer, region_removed Region, data_removed []rune) {
	t.scroll(changed_buffer)
}

func (t *qmlfrontend) Inserted(changed_buffer Buffer, region_inserted Region, data_inserted []rune) {
	t.scroll(changed_buffer)
}

// Apparently calling qml.Changed also triggers a re-draw, meaning that typed text is at the
// mercy of how quick Qt happens to be rendering.
// Try setting batching_enabled = false to see the effects of non-batching
func (t *qmlfrontend) qmlBatchLoop() {
	queue := make(map[qmlDispatch]bool)
	t.qmlDispatch = make(chan qmlDispatch, 1000)
	for {
		if len(queue) > 0 {
			select {
			case <-time.After(time.Millisecond * 20):
				// Nothing happened for 20 milliseconds, so dispatch all queued changes
				for k := range queue {
					qml.Changed(k.value, k.field)
				}
				queue = make(map[qmlDispatch]bool)
			case d := <-t.qmlDispatch:
				queue[d] = true
			}
		} else {
			queue[<-t.qmlDispatch] = true
		}
	}
}

func (t *qmlfrontend) qmlChanged(value, field interface{}) {
	if !batching_enabled {
		qml.Changed(value, field)
	} else {
		t.qmlDispatch <- qmlDispatch{value, field}
	}
}

func (t *qmlfrontend) DefaultBg() color.RGBA {
	c := t.ColorScheme().Spice(&render.ViewRegions{})
	c.Background.A = 0xff
	return color.RGBA(c.Background)
}

func (t *qmlfrontend) DefaultFg() color.RGBA {
	c := t.ColorScheme().Spice(&render.ViewRegions{})
	c.Foreground.A = 0xff
	return color.RGBA(c.Foreground)
}

// Called when a new view is opened
func (t *qmlfrontend) onNew(v *backend.View) {
	fv := &frontendView{bv: v}
	v.AddObserver(fv)
	v.Settings().AddOnChange("blah", fv.onChange)

	fv.Title.Text = v.FileName()
	if len(fv.Title.Text) == 0 {
		fv.Title.Text = "untitled"
	}

	w2 := t.windows[v.Window()]
	w2.views = append(w2.views, fv)

	if w2.window == nil {
		return
	}

	w2.window.Call("addTab", "", fv)
}

// called when a view is closed
func (t *qmlfrontend) onClose(v *backend.View) {
	w2 := t.windows[v.Window()]
	for i := range w2.views {
		if w2.views[i].bv == v {
			w2.window.Call("removeTab", i)
			copy(w2.views[i:], w2.views[i+1:])
			w2.views = w2.views[:len(w2.views)-1]
			return
		}
	}
	log.Error("Couldn't find closed view...")
}

// called when a view has loaded
func (t *qmlfrontend) onLoad(v *backend.View) {
	w2 := t.windows[v.Window()]
	i := 0
	for i = range w2.views {
		if w2.views[i].bv == v {
			break
		}
	}
	v2 := w2.views[i]
	v2.Title.Text = v.FileName()
	if w2.window != nil {
		w2.window.Call("activateTab", w2.ActiveViewIndex())
		w2.window.Call("setTabTitle", i, v2.Title.Text)
	}
}

func (t *qmlfrontend) onSelectionModified(v *backend.View) {
	w2 := t.windows[v.Window()]
	i := 0
	for i = range w2.views {
		if w2.views[i].bv == v {
			break
		}
	}
	v2 := w2.views[i]
	if v2.qv == nil {
		return
	}
	v2.qv.Call("onSelectionModified")
}

func (t *qmlfrontend) onStatusChanged(v *backend.View) {
	w2 := t.windows[v.Window()]
	i := 0
	for i = range w2.views {
		if w2.views[i].bv == v {
			break
		}
	}
	v2 := w2.views[i]
	if v2.qv == nil {
		return
	}
	v2.qv.Call("onStatusChanged")
}

// Launches the provided command in a new goroutine
// (to avoid locking up the GUI)
func (t *qmlfrontend) RunCommand(command string) {
	t.RunCommandWithArgs(command, make(backend.Args))
}

func (t *qmlfrontend) RunCommandWithArgs(command string, args backend.Args) {
	ed := backend.GetEditor()
	go ed.RunCommand(command, args)
}

func (t *qmlfrontend) HandleInput(text string, keycode int, modifiers int) bool {
	log.Debug("qmlfrontend.HandleInput: text=%v, key=%x, modifiers=%x", text, keycode, modifiers)
	shift := false
	alt := false
	ctrl := false
	super := false

	if key, ok := lut[keycode]; ok {
		ed := backend.GetEditor()

		if (modifiers & shift_mod) != 0 {
			shift = true
		}
		if (modifiers & alt_mod) != 0 {
			alt = true
		}
		if (modifiers & ctrl_mod) != 0 {
			if runtime.GOOS == "darwin" {
				super = true
			} else {
				ctrl = true
			}
		}
		if (modifiers & meta_mod) != 0 {
			if runtime.GOOS == "darwin" {
				ctrl = true
			} else {
				super = true
			}
		}

		ed.HandleInput(keys.KeyPress{Text: text, Key: key, Shift: shift, Alt: alt, Ctrl: ctrl, Super: super})
		return true
	}
	return false
}

func (t *qmlfrontend) ColorScheme() backend.ColorScheme {
	ed := backend.GetEditor()
	return ed.GetColorScheme(ed.Settings().Get("color_scheme", "").(string))
}

// Quit closes all open windows to de-reference all qml objects
func (t *qmlfrontend) Quit() (err error) {
	// todo: handle changed files that aren't saved.
	for _, v := range t.windows {
		if v.window != nil {
			v.window.Hide()
			v.window.Destroy()
			v.window = nil
		}
	}
	return
}

func (t *qmlfrontend) loop() (err error) {
	ed := backend.GetEditor()
	// TODO: As InitCallback doc says initiation code to be deferred until
	// after the UI is up and running. but because we dont have any
	// scheme we are initing editor before the UI comes up.
	ed.Init()
	ed.SetDefaultPath("../packages/Default")
	ed.SetUserPath("../packages/User")

	// Some packages(e.g Vintageos) need available window and view at start
	// so we need at least one window and view(event empty) before loading
	// packages. Sublime text also has available window view on startup
	w := ed.NewWindow()
	// TODO: we should open empty file but due qml package error we can't
	// open any file from the qml frontend for now
	v := w.OpenFile("main.go", 0)
	ed.AddPackagesPath("../packages")

	ed.SetFrontend(t)
	ed.LogInput(false)
	ed.LogCommands(false)

	c := ed.Console()
	t.Console = &frontendView{bv: c}
	c.AddObserver(t.Console)
	c.AddObserver(t)

	var (
		engine    *qml.Engine
		component qml.Object
		// WaitGroup keeping track of open windows
		wg sync.WaitGroup
	)

	// create and setup a new engine, destroying
	// the old one if one exists.
	//
	// This is needed to re-load qml files to get
	// the new file contents from disc as otherwise
	// the old file would still be what is referenced.
	newEngine := func() (err error) {
		if engine != nil {
			log.Debug("calling destroy")
			// TODO(.): calling this appears to make the editor *very* crash-prone, just let it leak for now
			// engine.Destroy()
			engine = nil
		}
		log.Debug("calling newEngine")
		engine = qml.NewEngine()
		engine.On("quit", t.Quit)
		log.Debug("setvar frontend")
		engine.Context().SetVar("frontend", t)
		log.Debug("setvar editor")
		engine.Context().SetVar("editor", backend.GetEditor())

		log.Debug("loadfile")
		component, err = engine.LoadFile(qmlWindowFile)
		return
	}
	if err := newEngine(); err != nil {
		log.Error(err)
	}

	addWindow := func(w *backend.Window) {
		fw := &frontendWindow{bw: w}
		t.windows[w] = fw
		if component != nil {
			fw.launch(&wg, component)
		}
	}

	backend.OnNew.Add(t.onNew)
	backend.OnClose.Add(t.onClose)
	backend.OnLoad.Add(t.onLoad)
	backend.OnSelectionModified.Add(t.onSelectionModified)
	backend.OnNewWindow.Add(addWindow)
	backend.OnStatusChanged.Add(t.onStatusChanged)

	// we need to add windows and views that are added before we registered
	// actions for OnNewWindow, OnNew, ... events
	for _, w := range ed.Windows() {
		addWindow(w)
		for _, v := range w.Views() {
			t.onNew(v)
			t.onLoad(v)
		}
	}

	// also because we are openning a file now we need to set the syntax
	// manually here
	v.Settings().Set("syntax", "../packages/go-tmbundle/Syntaxes/Go.tmLanguage")

	defer func() {
		fmt.Println(util.Prof)
	}()

	// The rest of code is related to livereloading qml files
	// TODO: this doesnt work currently
	watch, err := fsnotify.NewWatcher()
	if err != nil {
		log.Error("Unable to create file watcher: %s", err)
		return
	}
	defer watch.Close()
	watch.Add("qml")
	defer watch.Remove("qml")

	reloadRequested := false
	waiting := false

	go func() {
		// reloadRequested = true
		// t.Quit()

		lastTime := time.Now()

		for {

			select {
			case ev := <-watch.Events:
				if time.Now().Sub(lastTime) < 1*time.Second {
					// quitting too frequently causes crashes
					lastTime = time.Now()
					continue
				}
				if strings.HasSuffix(ev.Name, ".qml") && ev.Op == fsnotify.Write && ev.Op != fsnotify.Chmod && !reloadRequested && waiting {
					reloadRequested = true
					t.Quit()
					lastTime = time.Now()
				}
			}
		}
	}()

	for {
		// Reset reload status
		reloadRequested = false

		log.Debug("Waiting for all windows to close")
		// wg would be the WaitGroup all windows belong to, so first we wait for
		// all windows to close.
		waiting = true
		wg.Wait()
		waiting = false
		log.Debug("All windows closed. reloadRequest: %v", reloadRequested)
		// then we check if there's a reload request in the pipe
		if !reloadRequested || len(t.windows) == 0 {
			// This would be a genuine exit; all windows closed by the user
			break
		}

		// *We* closed all windows because we want to reload freshly changed qml
		// files.
		for {
			log.Debug("Calling newEngine")
			if err := newEngine(); err != nil {
				// Reset reload status
				reloadRequested = false
				log.Error(err)
				for !reloadRequested {
					// This loop allows us to re-try reloading
					// if there was an error in the file this time,
					// we just loop around again when we receive the next
					// reload request (ie on the next save of the file).
					time.Sleep(time.Second)
				}
				continue
			}
			log.Debug("break")
			break
		}
		log.Debug("re-launching all windows")
		// Succeeded loading the file, re-launch all windows
		for _, v := range t.windows {
			v.launch(&wg, component)

			for i, bv := range v.Back().Views() {
				t.onNew(bv)
				t.onLoad(bv)

				v.View(i)
			}
		}
	}

	return
}
