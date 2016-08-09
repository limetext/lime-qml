package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/limetext/backend"
	"github.com/ryanuber/go-glob"
)

// Filesystem tree item

type ProjFolderItem struct {
	TreeListItemImpl
	path string
	name string

	projFolder *backend.Folder
}

type ByNameFoldersFirst2 []TreeListItem

func (a ByNameFoldersFirst2) Len() int      { return len(a) }
func (a ByNameFoldersFirst2) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByNameFoldersFirst2) Less(i, j int) bool {
	if a[i].HasChildren() != a[j].HasChildren() {
		return a[i].HasChildren()
	}
	return a[i].Text() < a[j].Text()
}

var _ TreeListItem = &ProjFolderItem{}

func (f *ProjFolderItem) Text() string {
	return f.name
}

func (f *ProjFolderItem) Type() string {
	return "file"
}

func (f *ProjFolderItem) Icon() string {
	return ""
}

func (f *ProjFolderItem) HasChildren() bool {
	return true
}

func (f *ProjFolderItem) Children() []TreeListItem {
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

	items := make([]TreeListItem, 0, len(files)/2)
	for _, fx := range files {
		path := filepath.Join(f.path, fx.Name())
		if fx.IsDir() {
			if !f.isIncluded(path+"/", f.projFolder.IncludePatterns, f.projFolder.ExcludePatterns) {
				continue
			}
			items = append(items, &ProjFolderItem{
				path:       path,
				name:       fx.Name(),
				projFolder: f.projFolder,
			})
		} else {
			if !f.isIncluded(fx.Name(), f.projFolder.FileIncludePatterns, f.projFolder.FileExcludePatterns) {
				continue
			}
			items = append(items, &FSTreeItem{
				path:  path,
				name:  fx.Name(),
				isDir: false,
			})
		}
	}

	sort.Sort(ByNameFoldersFirst2(items))

	return items
}

func (f *ProjFolderItem) isIncluded(path string, includePatterns []string, excludePatterns []string) bool {
	if len(includePatterns) > 0 {
		included := false
		for _, incPat := range includePatterns {
			if glob.Glob(incPat, path) {
				included = true
				break
			}
		}
		if !included {
			return false
		}
	}

	for _, excPat := range excludePatterns {
		if glob.Glob(excPat, path) {
			return false
		}
	}

	return true
}

func (f *ProjFolderItem) Click() {
	fmt.Println("Click: ", f.name)
}

func (f *ProjFolderItem) SetVisible(vis bool) {
	// fmt.Println("Visible: ", f.name, vis)
}

func (f *ProjFolderItem) SetExpanded(vis bool) {
	fmt.Println("Expanded: ", f.name, vis)
}
