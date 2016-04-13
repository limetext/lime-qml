// Copyright 2013 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"gopkg.in/qml.v1"

	"github.com/limetext/lime-backend/lib"
	"github.com/limetext/lime-backend/lib/log"
	"github.com/limetext/lime-backend/lib/render"
	"github.com/limetext/lime-backend/lib/util"
	. "github.com/limetext/text"
)

// A helper glue structure connecting the backend View with the qml code that
// then ends up rendering it.
type frontendView struct {
	bv            *backend.View
	qv            qml.Object
	FormattedLine []*lineStruct
	Title         lineStruct
}

// This allows us to trigger a qml.Changed on a specific line in the view so
// that only it is re-rendered by qml
type lineStruct struct {
	Text     string
	RawText  string
	Chunks   []lineChunk
	Width    int
	Measured bool
}

func (l *lineStruct) ChunksLen() int {
	return len(l.Chunks)
}

func (l *lineStruct) Chunk(i int) *lineChunk {
	return &l.Chunks[i]
}

type lineChunk struct {
	Text       string
	Background string
	Foreground string
	SkipWidth  int
	Width      int
	Measured   bool
}

// htmlcol returns the hex color value for the given Colour object
func htmlcol(c render.Colour) string {
	return fmt.Sprintf("%02X%02X%02X", c.R, c.G, c.B)
}

func (fv *frontendView) Lines() int {
	return len(fv.FormattedLine)
}

func (fv *frontendView) Line(index int) *lineStruct {
	if index < 0 || index >= len(fv.FormattedLine) {
		log.Error("Error? Line index out of bounds: %v %v\n", index, len(fv.FormattedLine))
		return nil
	}
	return fv.FormattedLine[index]
}

func (fv *frontendView) Region(a int, b int) Region {
	return Region{a, b}
}

func (fv *frontendView) RegionLines() int {
	var count int = 0
	regs := fv.bv.Sel().Regions()
	if fv.bv != nil {
		for _, r := range regs {
			count += len(fv.bv.Lines(r))
		}
	}
	return count
}

func (fv *frontendView) Setting(name string) interface{} {
	return fv.Back().Settings().Get(name, nil)
}

func (fv *frontendView) Back() *backend.View {
	return fv.bv
}

func (fv *frontendView) Fix(obj qml.Object) {
	fv.qv = obj
	obj.On("destroyed", func() {
		if fv.qv == obj {
			fv.qv = nil
		}
	})

	if len(fv.FormattedLine) == 0 && fv.bv.Size() > 0 {
		fv.bufferChanged(0, fv.bv.Size())
		return
	}

	log.Info("Fix: %v  %v  %v", obj, len(fv.FormattedLine), fv.bv.Size())

	for i := range fv.FormattedLine {
		_ = i
		obj.Call("addLine")
	}
}

func (fv *frontendView) bufferChanged(pos, delta int) {
	prof := util.Prof.Enter("frontendView.bufferChanged")
	defer prof.Exit()

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered from error in bufferChanged", r)
			fv.qv = nil
		}
	}()

	row1, _ := fv.bv.RowCol(pos)
	row2, _ := fv.bv.RowCol(pos + delta)
	if row1 > row2 {
		row1, row2 = row2, row1
	}

	if delta > 0 && fv.qv != nil {
		r1 := row1
		if add := strings.Count(fv.bv.Substr(Region{pos, pos + delta}), "\n"); add > 0 {
			nn := make([]*lineStruct, len(fv.FormattedLine)+add)
			copy(nn, fv.FormattedLine[:r1])
			copy(nn[r1+add:], fv.FormattedLine[r1:])
			for i := 0; i < add; i++ {
				nn[r1+i] = &lineStruct{}
			}
			fv.FormattedLine = nn
			for i := 0; i < add; i++ {
				fv.qv.Call("insertLine", r1+i)
			}
		}
	}

	for i := row1; i <= row2; i++ {
		fv.formatLine(i)
	}
}

func (fv *frontendView) Erased(changed_buffer Buffer, region_removed Region, data_removed []rune) {
	fv.bufferChanged(region_removed.B, region_removed.A-region_removed.B)
}

func (fv *frontendView) Inserted(changed_buffer Buffer, region_inserted Region, data_inserted []rune) {
	fv.bufferChanged(region_inserted.A, region_inserted.B-region_inserted.A)
}

func (fv *frontendView) onChange(name string) {
	if name != "lime.syntax.updated" {
		return
	}
	// force redraw, as the syntax regions might have changed...
	for i := range fv.FormattedLine {
		fv.formatLine(i)
	}
}

func (fv *frontendView) formatLine(line int) {
	prof := util.Prof.Enter("frontendView.formatLine")
	defer prof.Exit()
	buf := bytes.NewBuffer(nil)
	vr := fv.bv.Line(fv.bv.TextPoint(line, 0))
	for line >= len(fv.FormattedLine) {
		fv.FormattedLine = append(fv.FormattedLine, &lineStruct{})
		if fv.qv != nil {
			fv.qv.Call("addLine")
		}
	}
	if vr.Size() == 0 {
		if fv.FormattedLine[line].Text != "" {
			fv.FormattedLine[line].Text = ""
			fv.FormattedLine[line].Chunks = fv.FormattedLine[line].Chunks[0:0]
			t.qmlChanged(fv.FormattedLine[line], fv.FormattedLine[line])
		}
		return
	}
	recipie := fv.bv.Transform(scheme, vr).Transcribe()
	highlight_line := false
	if b, ok := fv.bv.Settings().Get("highlight_line", highlight_line).(bool); ok {
		highlight_line = b
	}
	lastEnd := vr.Begin()

	chunks := make([]lineChunk, 0, len(recipie))

	for _, reg := range recipie {
		if lastEnd != reg.Region.Begin() {
			fmt.Fprintf(buf, "<span>%s</span>", fv.bv.Substr(Region{lastEnd, reg.Region.Begin()}))
			chunks = append(chunks, lineChunk{Text: fv.bv.Substr(Region{lastEnd, reg.Region.Begin()})})
		}
		fmt.Fprintf(buf, "<span style=\"white-space:pre; color:#%s; background:#%s\">%s</span>", htmlcol(reg.Flavour.Foreground), htmlcol(reg.Flavour.Background), fv.bv.Substr(reg.Region))
		chunks = append(chunks, lineChunk{Text: fv.bv.Substr(reg.Region), Foreground: htmlcol(reg.Flavour.Foreground), Background: htmlcol(reg.Flavour.Background)})
		lastEnd = reg.Region.End()
	}
	if lastEnd != vr.End() {
		io.WriteString(buf, fv.bv.Substr(Region{lastEnd, vr.End()}))
		chunks = append(chunks, lineChunk{Text: fv.bv.Substr(Region{lastEnd, vr.End()})})
	}

	str := buf.String()

	if fv.FormattedLine[line].Text != str {
		fv.FormattedLine[line].RawText = fv.bv.Substr(vr)
		fv.FormattedLine[line].Text = str
		fv.FormattedLine[line].Chunks = chunks
		t.qmlChanged(fv.FormattedLine[line], fv.FormattedLine[line])
	}
}
