// Copyright 2013 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"sync"

	"github.com/limetext/backend"
	"github.com/limetext/backend/log"
	"github.com/limetext/backend/render"
	"github.com/limetext/qml-go"
	. "github.com/limetext/text"
	"github.com/limetext/util"
)

// A helper glue structure connecting the backend View with the qml code that
// then ends up rendering it.
type view struct {
	id             int
	bv             *backend.View
	qv             qml.Object
	FormattedLines *linesList
	linesLock      sync.Mutex
	Title          string

	watchedSettings map[string]watchedSetting

	// these setting: tags are merely for ease of reading, they aren't actually used
	TabSize    int    `setting:"tab_size"`
	SyntaxName string `setting:"syntax"`

	FontSize int    `setting:"font_size"`
	FontFace string `setting:"font_face"`
}

func newView(bv *backend.View) *view {
	v := &view{
		id:    int(bv.Id()),
		bv:    bv,
		Title: bv.FileName(),
	}
	if len(v.Title) == 0 {
		v.Title = "untitled"
	}
	bv.AddObserver(v)
	bv.Settings().AddOnChange("qml.view", v.onSettingChange)

	watcher := newSettingsWatcher(v, bv.Settings())

	watcher.watchInt("tab_size", &v.TabSize, 4)
	watcher.watchString("syntax", &v.SyntaxName, "Plain Text", func(syn string) string {
		if syntax := backend.GetEditor().GetSyntax(syn); syntax != nil {
			return syntax.Name()
		}
		return syn
	})
	watcher.watchInt("font_size", &v.FontSize, 10)
	watcher.watchString("font_face", &v.FontFace, "Monospace")

	return v
}

// htmlcol returns the hex color value for the given Colour object
func htmlcol(c render.Colour) string {
	return fmt.Sprintf("%02X%02X%02X", c.R, c.G, c.B)
}

func (v *view) Region(a int, b int) Region {
	return Region{a, b}
}

func (v *view) RegionLines() int {
	var count int = 0
	regs := v.bv.Sel().Regions()
	if v.bv != nil {
		for _, r := range regs {
			count += len(v.bv.Lines(r))
		}
	}
	return count
}

func (v *view) Setting(name string) interface{} {
	return v.Back().Settings().Get(name, nil)
}

func (v *view) Back() *backend.View {
	return v.bv
}

func (v *view) Fix(obj qml.Object) {
	if v.qv == obj {
		// already Fixed?
		return
	}
	if v.qv != nil {
		log.Warn("view already has a QML view set! (%v, %v)", v.qv, obj)
	}
	v.qv = obj
	obj.On("destroyed", func() {
		if v.qv == obj {
			v.qv = nil
		}
	})

	v.linesLock.Lock()
	defer v.linesLock.Unlock()
	qml.RunMain(func() {
		v.FormattedLines = NewLinesList(obj.Common().Engine(), nil)

		// initialize the v.FormattedLines now
		r := Region{A: 0, B: v.bv.Size()}
		v.insertedNoLock(nil, r, v.bv.SubstrR(r))

		fe.qmlChanged(v, &v.FormattedLines)
	})

}

// SetActive is called from QML when the active tab is set to this view
func (v *view) SetActive() {
	v.bv.Window().SetActiveView(v.bv)
}

func (v *view) Erased(changed_buffer Buffer, region_removed Region, data_removed []rune) {
	if v.qv == nil {
		return
	}

	v.linesLock.Lock()
	defer v.linesLock.Unlock()

	prof := util.Prof.Enter("view.Erased")
	defer prof.Exit()

	row1, col1 := v.bv.RowCol(region_removed.A)

	newlines := 0
	for _, r := range data_removed {
		if r == '\n' {
			newlines += 1
		}
	}

	delLines := newlines //row2 - row1

	// first line
	if col1 > 0 { // line already exists, inserting in the middle of the line
		v.formatLine(row1, v.FormattedLines.get(row1))
		row1 += 1
		col1 = 0
	}

	if delLines > 0 {
		v.FormattedLines.deleteLines(row1, delLines)
	}
}

func (v *view) Inserted(changed_buffer Buffer, region_inserted Region, data_inserted []rune) {
	if v.qv == nil {
		return
	}

	v.linesLock.Lock()
	defer v.linesLock.Unlock()

	v.insertedNoLock(changed_buffer, region_inserted, data_inserted)
}
func (v *view) insertedNoLock(changed_buffer Buffer, region_inserted Region, data_inserted []rune) {
	prof := util.Prof.Enter("view.Inserted")
	defer prof.Exit()

	row1, col1 := v.bv.RowCol(region_inserted.A)
	row2, _ := v.bv.RowCol(region_inserted.B)

	addLines := row2 - row1

	// first line
	if col1 > 0 { // line already exists, inserting in the middle of the line
		v.formatLine(row1, v.FormattedLines.get(row1))
		row1 += 1
		col1 = 0
	}

	if addLines > 0 {
		newLines := make([]*lineStruct, addLines)
		for i := 0; i < addLines; i++ {
			line := &lineStruct{}
			v.formatLine(row1+i, line)
			newLines[i] = line
		}

		v.FormattedLines.insertLines(row1, newLines)
	}
}

func (v *view) onSettingChange(name string) {
	settings := v.bv.Settings()
	fmt.Printf("SettingChanged: %s %v\n", name, settings.Get(name))

	switch name {
	case "lime.syntax.updated":
		// force redraw, as the syntax regions might have changed...
		for i := 0; i < v.FormattedLines.len(); i++ {
			v.formatLine(i, v.FormattedLines.get(i))
		}

	}
}

func (v *view) formatLine(linenum int, line *lineStruct) {
	prof := util.Prof.Enter("view.formatLine")
	defer prof.Exit()

	vr := v.bv.Line(v.bv.TextPoint(linenum, 0))
	if vr.Size() == 0 {
		if line.Text != "" {
			line.Text = ""
			line.Chunks = line.Chunks[0:0]
			fe.qmlChanged(line, line)
		}
		return
	}
	recipie := v.bv.Transform(vr).Transcribe()
	lastEnd := vr.Begin()

	chunks := line.Chunks
	changed := false
	chunkI := 0

	nextChunk := func(lc lineChunk) {
		if chunkI >= len(chunks) {
			chunks = append(chunks, lc)
			changed = true
		} else if chunks[chunkI] != lc {
			chunks[chunkI] = lc
			changed = true
		}
		chunkI += 1
	}

	for _, reg := range recipie {
		if lastEnd != reg.Region.Begin() {
			lc := lineChunk{Text: v.bv.Substr(Region{lastEnd, reg.Region.Begin()})}
			nextChunk(lc)
		}
		lc := lineChunk{Text: v.bv.Substr(reg.Region), Foreground: htmlcol(reg.Flavour.Foreground), Background: htmlcol(reg.Flavour.Background)}
		nextChunk(lc)

		lastEnd = reg.Region.End()
	}
	if lastEnd != vr.End() {
		lc := lineChunk{Text: v.bv.Substr(Region{lastEnd, vr.End()})}
		nextChunk(lc)
	}

	if chunkI != len(chunks) {
		chunks = chunks[:chunkI]
		changed = true
	}

	if changed {
		line.Text = v.bv.Substr(vr)
		line.Chunks = chunks
		fe.qmlChanged(line, line)
	}
}
