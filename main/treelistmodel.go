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
	roots []*treeListRoot

	Model qml.ItemModel
	qml.ItemModelDefaultImpl
	internal qml.ItemModelInternal
}

type treeListRoot struct {
	rootItem TreeListItem
	nodes    []*treeListNode
}

var _ qml.ItemModelImpl = &TreeListModel{}

type treeListNode struct {
	Indent      int
	Expanded    bool
	Item        TreeListItem
	parent      *treeListNode
	root        *treeListRoot
	myChildren  int
	allChildren int
	model       *TreeListModel
}

func NewTreeListModel(engine *qml.Engine, parent qml.Object, rootItems []TreeListItem) *TreeListModel {
	t := &TreeListModel{}

	roots := make([]*treeListRoot, len(rootItems))
	for i, item := range rootItems {
		roots[i] = t.mkRoot(item)
	}

	t.roots = roots

	t.Model, t.internal = qml.NewItemModel(engine, parent, t)

	for _, r := range t.roots {
		for _, node := range r.nodes {
			node.Item.SetVisible(true)
		}
	}

	return t
}

func (t *TreeListModel) mkRoot(item TreeListItem) *treeListRoot {
	root := &treeListRoot{
		rootItem: item,
	}
	var items []TreeListItem
	if !item.HasChildren() {
		items = []TreeListItem{item}
	} else {
		item.SetExpanded(true)
		items = item.Children()
		// fmt.Println("rootChildren:", items)
	}
	root.nodes = t.mkNodes(items, 0, nil, root)

	return root
}

func (t *TreeListModel) mkNodes(items []TreeListItem, indent int, parent *treeListNode, root *treeListRoot) []*treeListNode {
	nodes := make([]*treeListNode, len(items))

	store := make([]treeListNode, len(items))
	for i, r := range items {
		store[i].Item = r
		store[i].Indent = indent
		store[i].parent = parent
		store[i].root = root
		nodes[i] = &store[i]
	}

	return nodes
}

func (t *TreeListModel) indexOfRoot(root *treeListRoot) int {
	indexSoFar := 0
	for _, r := range t.roots {
		if r == root {
			return indexSoFar
		}
		indexSoFar += len(r.nodes)
	}
	return -1
}

// rootForIndex takes a global index, and returns a root plus the index relative to that root
func (t *TreeListModel) rootAndIndex(gindex int) (root *treeListRoot, relIndex int) {
	indexSoFar := 0
	for _, r := range t.roots {
		if gindex < indexSoFar+len(r.nodes) {
			return r, gindex - indexSoFar
		}
		indexSoFar += len(r.nodes)
	}
	return nil, -1
}

func (t *TreeListModel) insertNodes(nodes []*treeListNode, index int, parent *treeListNode) {
	numNodes := len(nodes)
	parent.myChildren += numNodes

	p := parent
	for p != nil {
		p.allChildren += numNodes
		p = p.parent
	}

	root := parent.root
	rootIndex := t.indexOfRoot(root)

	// the inner append here likely creates a new slice
	// a faster way would be to extend the t.nodes if necessary
	// copy nodes after index to the end and copy nodes into the right place
	qml.RunMain(func() {
		t.internal.BeginInsertRows(nil, rootIndex+index, rootIndex+index+numNodes-1)
		root.nodes = append(root.nodes[:index], append(nodes, root.nodes[index:]...)...)
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

	root := parent.root
	rootIndex := t.indexOfRoot(root)

	for _, node := range root.nodes[index : index+length] {
		node.Item.impl().node = nil
		node.Item.SetVisible(false)
	}

	qml.RunMain(func() {
		t.internal.BeginRemoveRows(nil, rootIndex+index, rootIndex+index+length-1)
		nodesLen := len(root.nodes)
		copy(root.nodes[index:], root.nodes[index+length:])
		for k := nodesLen - length; k < nodesLen; k++ {
			root.nodes[k] = nil // clear pointers at the end of the array for GC
		}
		root.nodes = root.nodes[:nodesLen-length]
		t.internal.EndRemoveRows()
	})
}

func (t *TreeListModel) insertChild(parent *treeListNode, item TreeListItem, index int) {
	if !parent.Expanded {
		return
	}

	root := parent.root

	idx := t.findIndex(parent)
	idx += 1 // children
	for i := 0; i < index; i++ {
		idx += root.nodes[idx].allChildren
	}
	// idx is now the index where the new node should go

	nodes := t.mkNodes([]TreeListItem{item}, parent.Indent+1, parent, root)
	t.insertNodes(nodes, idx, parent)
}

func (t *TreeListModel) removeChild(parent *treeListNode, index int) {
	if !parent.Expanded {
		return
	}

	if index >= parent.myChildren {
		panic(fmt.Sprintf("child index (%v) > parent's children (%v)", index, parent.myChildren))
	}

	root := parent.root

	idx := t.findIndex(parent)
	idx += 1 // children
	for i := 0; i < index; i++ {
		idx += root.nodes[idx].allChildren
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
	root := node.root
	n := node
	p := node.parent
	if p != nil {
		idx = t.findIndex(n.parent)
		if root.nodes[idx].Expanded == false {
			panic("Parent node is not expanded?")
		}
		if idx == -1 {
			return -1
		}
	} else {
		// probably one of the roots. If not, it is not in the tree!
	}

	for {
		if idx >= len(root.nodes) {
			return -1
		}
		item := root.nodes[idx]
		if item.parent != node.parent {
			panic("item.parent != node.parent!")
		}
		if node == item {
			return idx
		}
		idx += item.allChildren
	}
}

func (t *TreeListModel) ExpandNode(gindex int) {
	root, index := t.rootAndIndex(gindex)

	node := root.nodes[index]
	if !node.Item.HasChildren() {
		return
	}
	children := node.Item.Children()
	childNodes := t.mkNodes(children, node.Indent+1, node, root)
	node.Expanded = true
	node.Item.SetExpanded(true)
	qml.Changed(node, &node.Expanded)
	t.insertNodes(childNodes, index+1, node)
}

func (t *TreeListModel) CollapseNode(gindex int) {
	root, index := t.rootAndIndex(gindex)

	node := root.nodes[index]
	node.Expanded = false
	node.Item.SetExpanded(false)
	qml.Changed(node, &node.Expanded)
	t.removeNodes(index+1, node.allChildren, node)
}

//ItemModelImpl

func (t *TreeListModel) RowCount(parent qml.ModelIndex) int {
	count := 0
	for _, r := range t.roots {
		count += len(r.nodes)
	}
	// fmt.Println("rowcount: ", count)
	return count
}

func (t *TreeListModel) Data(mindex qml.ModelIndex, role qml.Role) interface{} {
	row := mindex.Row()
	root, index := t.rootAndIndex(row)
	// fmt.Println("data: ", row, root, index, root.nodes[index])
	return root.nodes[index]
}

func (t *TreeListModel) Index(row int, column int, parent qml.ModelIndex) qml.ModelIndex {
	if !parent.IsValid() && column == 0 && row >= 0 {
		root, _ := t.rootAndIndex(row)
		if root == nil {
			return nil
		}
		return t.internal.CreateIndex(row, column, 0)
	}
	return nil
}
