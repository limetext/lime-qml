// Copyright 2016 The lime Authors.
// Use of this source code is governed by a 2-clause
// BSD-style license that can be found in the LICENSE file.

package main

import "github.com/limetext/qml-go"

type linesList struct {
	qml.ItemModel
	qml.ItemModelDefaultImpl
	lines    []*lineStruct
	internal qml.ItemModelInternal
}

var _ qml.ItemModelImpl = &linesList{}

func NewLinesList(engine *qml.Engine, parent qml.Object) *linesList {
	ll := &linesList{
		lines: make([]*lineStruct, 1, 10),
	}
	ll.lines[0] = &lineStruct{} // always have to have at least 1 line

	ll.ItemModel, ll.internal = qml.NewItemModel(engine, parent, ll)
	return ll
}

func (l *linesList) RowCount(parent qml.ModelIndex) int {
	return len(l.lines)
}

func (l *linesList) Data(index qml.ModelIndex, role qml.Role) interface{} {
	if index.IsValid() {
		return l.lines[index.Row()]
	}
	return nil
}

func (l *linesList) Index(row int, column int, parent qml.ModelIndex) qml.ModelIndex {
	if !parent.IsValid() && column == 0 && row >= 0 && row < len(l.lines) {
		return l.internal.CreateIndex(row, column, 0)
	}
	return nil
}

func (l *linesList) get(i int) *lineStruct {
	return l.lines[i]
}

func (l *linesList) len() int {
	return len(l.lines)
}

func (l *linesList) insertLines(pos int, newLines []*lineStruct) {
	add := len(newLines)
	nn := make([]*lineStruct, len(l.lines)+add)
	copy(nn, l.lines[:pos])
	copy(nn[pos+add:], l.lines[pos:])
	copy(nn[pos:pos+add], newLines)

	qml.RunMain(func() {
		l.internal.BeginInsertRows(nil, pos, pos+add-1)
		l.lines = nn
		l.internal.EndInsertRows()
	})
}

func (l *linesList) deleteLines(pos int, count int) {

	qml.RunMain(func() {
		l.internal.BeginRemoveRows(nil, pos, pos+count-1)
		copy(l.lines[pos:], l.lines[pos+count:])
		l.lines = l.lines[:len(l.lines)-count]
		l.internal.EndRemoveRows()
	})
}

// This allows us to trigger a qml.Changed on a specific line in the view so
// that only it is re-rendered by qml
type lineStruct struct {
	Text     string
	Chunks   []lineChunk
	Width    int
	Measured bool
}

func (l *lineStruct) ChunksLen() int {
	if l == nil {
		return 0
	}
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
