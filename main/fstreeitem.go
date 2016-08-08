package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
)

// Filesystem tree item

type FSTreeItem struct {
	TreeListItemImpl
	path  string
	name  string
	isDir bool
}

type ByNameFoldersFirst []FSTreeItem

func (a ByNameFoldersFirst) Len() int      { return len(a) }
func (a ByNameFoldersFirst) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByNameFoldersFirst) Less(i, j int) bool {
	if a[i].isDir != a[j].isDir {
		return a[i].isDir
	}
	return a[i].name < a[j].name
}

var _ TreeListItem = &FSTreeItem{}

func (f *FSTreeItem) Text() string {
	return f.name
}

func (f *FSTreeItem) Type() string {
	return "file"
}

func (f *FSTreeItem) Icon() string {
	return ""
}

func (f *FSTreeItem) HasChildren() bool {
	return f.isDir
}

func (f *FSTreeItem) Children() []TreeListItem {
	file, err := os.Open(f.path)
	if err != nil {
		log.Printf("Error opening dir %#v: %s", f.path, err)
		return nil
	}
	defer file.Close()

	files, err := file.Readdir(-1)
	if err != nil {
		log.Printf("Error retreiving children of %#v: %s", f.path, err)
	}

	items := make([]TreeListItem, len(files))
	store := make([]FSTreeItem, len(files))
	for i, fx := range files {
		store[i] = FSTreeItem{
			path:  filepath.Join(f.path, fx.Name()),
			name:  fx.Name(),
			isDir: fx.IsDir(),
		}
		items[i] = &store[i]
	}

	sort.Sort(ByNameFoldersFirst(store)) // items pointers won't change

	return items
}

func (f *FSTreeItem) Click() {
	fmt.Println("Click: ", f.name)
}

func (f *FSTreeItem) SetVisible(vis bool) {
	fmt.Println("Visible: ", f.name, vis)
}
