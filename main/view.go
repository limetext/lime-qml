// Copyright 2013 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

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

func (v *view) Lines() int {
	return len(v.FormattedLine)
}

func (v *view) Line(index int) *lineStruct {
	if index < 0 || index >= len(v.FormattedLine) {
		log.Error("Error? Line index out of bounds: %v %v\n", index, len(v.FormattedLine))
		return nil
	}
	return v.FormattedLine[index]
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

	if len(v.FormattedLine) == 0 && v.bv.Size() > 0 {
		v.bufferChanged(0, v.bv.Size())
		return
	}

	log.Info("Fix: %v  %v  %v", obj, len(v.FormattedLine), v.bv.Size())

	for i := range v.FormattedLine {
		_ = i
		obj.Call("addLine")
	}
}

func (v *view) bufferChanged(pos, delta int) {
	prof := util.Prof.Enter("view.bufferChanged")
	defer prof.Exit()

	defer func() {
		if r := recover(); r != nil {
			log.Error("Recovered from error in bufferChanged", r)
			v.qv = nil
		}
	}()

	row1, _ := v.bv.RowCol(pos)
	row2, _ := v.bv.RowCol(pos + delta)
	if row1 > row2 {
		row1, row2 = row2, row1
	}

	if delta > 0 && v.qv != nil {
		r1 := row1
		if add := strings.Count(v.bv.Substr(Region{pos, pos + delta}), "\n"); add > 0 {
			nn := make([]*lineStruct, len(v.FormattedLine)+add)
			copy(nn, v.FormattedLine[:r1])
			copy(nn[r1+add:], v.FormattedLine[r1:])
			for i := 0; i < add; i++ {
				nn[r1+i] = &lineStruct{}
			}
			v.FormattedLine = nn
			for i := 0; i < add; i++ {
				v.qv.Call("insertLine", r1+i)
			}
		}
	}

	for i := row1; i <= row2; i++ {
		v.formatLine(i)
	}
}

func (v *view) Erased(changed_buffer Buffer, region_removed Region, data_removed []rune) {
	v.bufferChanged(region_removed.B, region_removed.A-region_removed.B)
}

func (v *view) Inserted(changed_buffer Buffer, region_inserted Region, data_inserted []rune) {
	v.bufferChanged(region_inserted.A, region_inserted.B-region_inserted.A)
}

func (v *view) onChange(name string) {
	if name != "syntax" {
		return
	}
	// force redraw, as the syntax regions might have changed...
	for i := range v.FormattedLine {
		v.formatLine(i)
	}
}

func (v *view) formatLine(line int) {
	prof := util.Prof.Enter("view.formatLine")
	defer prof.Exit()
	buf := bytes.NewBuffer(nil)
	vr := v.bv.Line(v.bv.TextPoint(line, 0))
	for line >= len(v.FormattedLine) {
		v.FormattedLine = append(v.FormattedLine, &lineStruct{})
		if v.qv != nil {
			v.qv.Call("addLine")
		}
	}
	if vr.Size() == 0 {
		if v.FormattedLine[line].Text != "" {
			v.FormattedLine[line].Text = ""
			v.FormattedLine[line].Chunks = v.FormattedLine[line].Chunks[0:0]
			fe.qmlChanged(v.FormattedLine[line], v.FormattedLine[line])
		}
		return
	}
	recipie := v.bv.Transform(vr).Transcribe()
	highlight_line := false
	if b, ok := v.bv.Settings().Get("highlight_line", highlight_line).(bool); ok {
		highlight_line = b
	}
	lastEnd := vr.Begin()

	chunks := make([]lineChunk, 0, len(recipie))

	for _, reg := range recipie {
		if lastEnd != reg.Region.Begin() {
			fmt.Fprintf(buf, "<span>%s</span>", v.bv.Substr(Region{lastEnd, reg.Region.Begin()}))
			chunks = append(chunks, lineChunk{Text: v.bv.Substr(Region{lastEnd, reg.Region.Begin()})})
		}
		fmt.Fprintf(buf, "<span style=\"white-space:pre; color:#%s; background:#%s\">%s</span>", htmlcol(reg.Flavour.Foreground), htmlcol(reg.Flavour.Background), v.bv.Substr(reg.Region))
		chunks = append(chunks, lineChunk{Text: v.bv.Substr(reg.Region), Foreground: htmlcol(reg.Flavour.Foreground), Background: htmlcol(reg.Flavour.Background)})
		lastEnd = reg.Region.End()
	}
	if lastEnd != vr.End() {
		io.WriteString(buf, v.bv.Substr(Region{lastEnd, vr.End()}))
		chunks = append(chunks, lineChunk{Text: v.bv.Substr(Region{lastEnd, vr.End()})})
	}

	str := buf.String()

	if v.FormattedLine[line].Text != str {
		v.FormattedLine[line].RawText = v.bv.Substr(vr)
		v.FormattedLine[line].Text = str
		v.FormattedLine[line].Chunks = chunks
		fe.qmlChanged(v.FormattedLine[line], v.FormattedLine[line])
	}
}
