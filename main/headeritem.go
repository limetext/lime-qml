package main

type HeaderItem struct {
	TreeListItemImpl
	name    string
	isChild bool
}

func (f *HeaderItem) Text() string {
	return f.name
}

func (f *HeaderItem) Type() string {
	return "header"
}

func (f *HeaderItem) Click() {

}

//
// func (f *FSTreeItem) HasChildren() bool {
// 	return f.isDir
// }
//
// func (f *FSTreeItem) Children() []TreeListItem {
// 	file, err := os.Open(f.path)
// 	if err != nil {
// 		log.Printf("Error opening dir %#v: %s", f.path, err)
// 		return nil
// 	}
// 	defer file.Close()
//
// 	files, err := file.Readdir(-1)
// 	if err != nil {
// 		log.Printf("Error retreiving children of %#v: %s", f.path, err)
// 	}
//
// 	items := make([]TreeListItem, len(files))
// 	store := make([]FSTreeItem, len(files))
// 	for i, fx := range files {
// 		store[i] = FSTreeItem{
// 			path:  filepath.Join(f.path, fx.Name()),
// 			name:  fx.Name(),
// 			isDir: fx.IsDir(),
// 		}
// 		items[i] = &store[i]
// 	}
//
// 	sort.Sort(ByNameFoldersFirst(store)) // items pointers won't change
//
// 	return items
// }
