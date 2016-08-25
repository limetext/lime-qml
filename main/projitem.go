package main

import (
	"path/filepath"
	"sort"

	"github.com/limetext/backend"
)

type ProjectItem struct {
	TreeListItemImpl

	Project *backend.Project
}

var _ TreeListItem = &ProjectItem{}

func (f *ProjectItem) Text() string {
	return ""
}

func (f *ProjectItem) Type() string {
	return "project"
}

func (f *ProjectItem) Icon() string {
	return ""
}

func (f *ProjectItem) HasChildren() bool {
	return true
}

func (f *ProjectItem) Children() []TreeListItem {
	folderPaths := f.Project.Folders()
	folders := make([]*backend.Folder, len(folderPaths))
	for i, p := range folderPaths {
		folders[i] = f.Project.Folder(p)
	}

	items := make([]TreeListItem, len(folders))
	store := make([]ProjFolderItem, len(folders))
	for i, fx := range folders {
		name := fx.Name
		if name == "" {
			name = filepath.Base(fx.Path)
		}
		store[i] = ProjFolderItem{
			path:       fx.Path,
			name:       name,
			projFolder: fx,
		}
		items[i] = &store[i]
	}

	sort.Sort(ByNameFoldersFirst2(items)) // items pointers won't change

	return items
}

func (f *ProjectItem) Click() {
	// fmt.Println("Click: ", f.name)
}

func (f *ProjectItem) SetVisible(vis bool) {
	// fmt.Println("Visible: ", f.name, vis)
}

func (f *ProjectItem) SetExpanded(vis bool) {
	// fmt.Println("Expanded: ", f.name, vis)
}
