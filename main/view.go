// Copyright 2013 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strconv"

	"github.com/limetext/backend"
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
	Title          string
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
	bv.Settings().AddOnChange("qml.view.syntax", v.onChange)
	bv.Settings().AddOnChange("qml.view.syntaxfile", func(name string) {
		if name != "syntax" {
			return
		}
		syn := bv.Settings().String("syntax", "Plain Text")
		syntax := backend.GetEditor().GetSyntax(syn)
		w := fe.windows[bv.Window()]
		w.qw.Call("setSyntaxStatus", syntax.Name())
	})
	bv.Settings().AddOnChange("qml.view.tabSize", func(name string) {
		if name != "tab_size" {
			return
		}
		ts := bv.Settings().Int("tab_size", 4)
		w := fe.windows[bv.Window()]
		w.qw.Call("setIndentStatus", strconv.Itoa(ts))
	})
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
	v.qv = obj
	obj.On("destroyed", func() {
		if v.qv == obj {
			v.qv = nil
		}
	})

	qml.RunMain(func() {
		v.FormattedLines = NewLinesList(obj.Common().Engine(), nil)

	})

	r := Region{A: 0, B: v.bv.Size()}
	v.Inserted(nil, r, v.bv.SubstrR(r))
}

// SetActive is called from QML when the active tab is set to this view
func (v *view) SetActive() {
	v.bv.Window().SetActiveView(v.bv)
}

func (v *view) Erased(changed_buffer Buffer, region_removed Region, data_removed []rune) {
	if v.qv == nil {
		return
	}

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

func (v *view) onChange(name string) {
	if name != "lime.syntax.updated" {
		return
	}
	// force redraw, as the syntax regions might have changed...
	for i := 0; i < v.FormattedLines.len(); i++ {
		v.formatLine(i, v.FormattedLines.get(i))
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
