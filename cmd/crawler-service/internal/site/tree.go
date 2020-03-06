package site

// Tree is a representation of a site tree.
type Tree struct {
	URL      string
	Children []Tree
}

func (t *Tree) Add(path string) {
	// Insert the path to the tree

}
