package main

import (
	"fmt"

	"github.com/limetext/qml-go"
)

// TreeListItem must include the TreeListItemImpl as an embedded struct
//     type MyItem struct {
//         TreeListItemImpl
//     }
// TreeListItemImpl includes do-nothing implementations for all functions except
// Text() and Click() to make implementations easier
//
type TreeListItem interface {
	// Text returns the text to display
	Text() string

	// Type for rendering e.g. "file" or "header"
	Type() string

	Icon() string

	// HasChildren should return if an expando arrow should appear
	HasChildren() bool

	// Children is only called when expanding the item
	Children() []TreeListItem

	// InsertChild is to be called when the children of this item changes, externally
	// to the tree list model. It will not do anything if the node is not expanded.
	InsertChild(item TreeListItem, index int)

	// RemoveChild is to be called when the children of this item changes, externally
	// to the tree list model. It will not do anything if the node is not expanded.
	RemoveChild(index int)

	// Click is called when the user selects the item
	Click()

	// SetVisible is called when the item is shown or hidden.
	SetVisible(bool)

	// SetExpanded is called when an item is expanded and its children are shown.
	// It can be used to add or remove watchers for its children.
	SetExpanded(bool)

	impl() *TreeListItemImpl
}

type TreeListItemImpl struct {
	node *treeListNode
}

func (impl *TreeListItemImpl) impl() *TreeListItemImpl {
	return impl
}

// Type for rendering e.g. "file" or "header"
func (i *TreeListItemImpl) Type() string {
	return "item"
}

func (i *TreeListItemImpl) Icon() string {
	return ""
}

// HasChildren should return if an expando arrow should appear
func (i *TreeListItemImpl) HasChildren() bool {
	return false
}

// Children is only called when expanding the item
func (i *TreeListItemImpl) Children() []TreeListItem {
	return nil
}

// InsertChild is to be called when the children of this item changes, externally
// to the tree list model. It will not do anything if the node is not expanded.
func (i *TreeListItemImpl) InsertChild(item TreeListItem, index int) {
	i.node.model.insertChild(i.node, item, index)
}

// RemoveChild is to be called when the children of this item changes, externally
// to the tree list model. It will not do anything if the node is not expanded.
func (i *TreeListItemImpl) RemoveChild(index int) {
	i.node.model.removeChild(i.node, index)
}

// SetVisible is called when the item is shown or hidden.
func (i *TreeListItemImpl) SetVisible(bool) {

}

// SetExpanded is called when an item is expanded and its children are shown.
// It can be used to add or remove watchers for its children.
func (i *TreeListItemImpl) SetExpanded(bool) {

}

type TreeListModel struct {
	roots []TreeListItem
	nodes []*treeListNode

	Model qml.ItemModel
	qml.ItemModelDefaultImpl
	internal qml.ItemModelInternal
}

var _ qml.ItemModelImpl = &TreeListModel{}

type treeListNode struct {
	Indent      int
	Expanded    bool
	Item        TreeListItem
	parent      *treeListNode
	myChildren  int
	allChildren int
	model       *TreeListModel
}

func NewTreeListModel(engine *qml.Engine, parent qml.Object, roots []TreeListItem) *TreeListModel {
	t := &TreeListModel{
		roots: roots,
	}

	t.nodes = t.mkNodes(roots, 0)

	t.Model, t.internal = qml.NewItemModel(engine, parent, t)

	for _, node := range t.nodes {
		node.Item.SetVisible(true)
	}

	return t
}

func (t *TreeListModel) mkNodes(items []TreeListItem, indent int) []*treeListNode {
	nodes := make([]*treeListNode, len(items))

	store := make([]treeListNode, len(items))
	for i, r := range items {
		store[i].Item = r
		store[i].Indent = indent
		nodes[i] = &store[i]
	}

	return nodes
}

func (t *TreeListModel) insertNodes(nodes []*treeListNode, index int, parent *treeListNode) {
	numNodes := len(nodes)
	parent.myChildren += numNodes

	p := parent
	for p != nil {
		p.allChildren += numNodes
		p = p.parent
	}

	// the inner append here likely creates a new slice
	// a faster way would be to extend the t.nodes if necessary
	// copy nodes after index to the end and copy nodes into the right place
	qml.RunMain(func() {
		t.internal.BeginInsertRows(nil, index, index+numNodes-1)
		t.nodes = append(t.nodes[:index], append(nodes, t.nodes[index:]...)...)
		t.internal.EndInsertRows()
	})

	for _, node := range nodes {
		node.Item.impl().node = node
		node.Item.SetVisible(true)
	}
}

func (t *TreeListModel) removeNodes(index, length int, parent *treeListNode) {
	parent.myChildren -= length

	p := parent
	for p != nil {
		p.allChildren -= length
		p = p.parent
	}

	for _, node := range t.nodes[index : index+length] {
		node.Item.impl().node = nil
		node.Item.SetVisible(false)
	}

	qml.RunMain(func() {
		t.internal.BeginRemoveRows(nil, index, index+length-1)
		nodesLen := len(t.nodes)
		copy(t.nodes[index:], t.nodes[index+length:])
		for k, n := nodesLen-length, nodesLen; k < n; k++ {
			t.nodes[k] = nil // clear pointers at the end of the array for GC
		}
		t.nodes = t.nodes[:nodesLen-length]
		t.internal.EndRemoveRows()
	})
}

func (t *TreeListModel) insertChild(parent *treeListNode, item TreeListItem, index int) {
	if !parent.Expanded {
		return
	}

	idx := t.findIndex(parent)
	idx += 1 // children
	for i := 0; i < index; i++ {
		idx += t.nodes[idx].allChildren
	}
	// idx is now the index where the new node should go

	nodes := t.mkNodes([]TreeListItem{item}, parent.Indent+1)
	t.insertNodes(nodes, idx, parent)
}

func (t *TreeListModel) removeChild(parent *treeListNode, index int) {
	if !parent.Expanded {
		return
	}

	if index >= parent.myChildren {
		panic(fmt.Sprintf("child index (%v) > parent's children (%v)", index, parent.myChildren))
	}

	idx := t.findIndex(parent)
	idx += 1 // children
	for i := 0; i < index; i++ {
		idx += t.nodes[idx].allChildren
	}
	// idx is now the index of the child to remove

	t.removeNodes(idx, 1, parent)
}

// FindIndex returns the absolute index of the item in the currently visible list,
// or -1 if it is not currently visible.
func (t *TreeListModel) FindIndex(item TreeListItem) int {
	impl := item.impl()
	if impl == nil {
		return -1
	}
	return t.findIndex(impl.node)
}

func (t *TreeListModel) findIndex(node *treeListNode) int {
	idx := 0
	n := node
	p := node.parent
	if p != nil {
		idx = t.findIndex(n.parent)
		if t.nodes[idx].Expanded == false {
			panic("Parent node is not expanded?")
		}
		if idx == -1 {
			return -1
		}
	} else {
		// probably one of the roots. If not, it is not in the tree!
	}

	for {
		if idx >= len(t.nodes) {
			return -1
		}
		item := t.nodes[idx]
		if item.parent != node.parent {
			panic("item.parent != node.parent!")
		}
		if node == item {
			return idx
		}
		idx += item.allChildren
	}
}

func (t *TreeListModel) ExpandNode(index int) {
	node := t.nodes[index]
	if !node.Item.HasChildren() {
		return
	}
	children := node.Item.Children()
	childNodes := t.mkNodes(children, node.Indent+1)
	node.Expanded = true
	qml.Changed(node, &node.Expanded)
	t.insertNodes(childNodes, index+1, node)
}

func (t *TreeListModel) CollapseNode(index int) {
	node := t.nodes[index]
	node.Expanded = false
	qml.Changed(node, &node.Expanded)
	t.removeNodes(index+1, node.allChildren, node)
}

//ItemModelImpl

func (t *TreeListModel) RowCount(parent qml.ModelIndex) int {
	return len(t.nodes)
}

func (t *TreeListModel) Data(index qml.ModelIndex, role qml.Role) interface{} {
	row := index.Row()
	return t.nodes[row]
}

func (t *TreeListModel) Index(row int, column int, parent qml.ModelIndex) qml.ModelIndex {
	if !parent.IsValid() && column == 0 && row >= 0 && row < len(t.nodes) {
		return t.internal.CreateIndex(row, column, 0)
	}
	return nil
}
