package site

// Tree is a representation of a site tree.
type Tree struct {
	URL      string
	Children []Tree
}
