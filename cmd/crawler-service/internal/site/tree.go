package site

import (
	"sort"
	"strings"
)

// Tree is a representation of a site tree.
type Tree struct {
	Root     string
	Children []*Tree
}

// Add inserts the path to the list of children. It will recursively walk down
// the children until it places it in the correct location.
func (t *Tree) Add(path string) {
	if len(path) == 0 {
		return
	}

	// Remove starting and trailing / to prevent empty strings in the tree.
	path = strings.Trim(path, "/")
	parts := strings.Split(path, "/")
	root := parts[0]
	children := parts[1:]

	t.add(root, children...)
}

func (t *Tree) add(root string, descendants ...string) {
	// It already exists so we don't need to add it to the tree.
	if t.Root == root {
		return
	}

	// Find where we should insert the current root.
	i := sort.Search(len(t.Children), func(i int) bool { return t.Children[i].Root >= root })

	// If we didn't find the root in the list of children then just insert all the children here.
	if i == len(t.Children) || t.Children[i].Root != root {
		// Create our own subtree because we haven't seen anything with this root
		subTree := &Tree{Root: root}
		parent := subTree

		// Walk down the list descendants and add them as subtrees to each other.
		for _, d := range descendants {
			parent.Children = []*Tree{{Root: d}}
			parent = parent.Children[0]
		}

		t.Children = insert(t.Children, subTree, i)
		return
	}

	// Our root matches the child at index i
	if len(descendants) > 0 {
		child := t.Children[i]
		child.add(descendants[0], descendants[1:]...)
	}

}

func insert(trees []*Tree, t *Tree, i int) []*Tree {
	// This will add 1 to the length our slice of trees if we need to, and will
	// only allocate an new slice if adding the new element grows the slice larger
	// than the capacity.
	trees = append(trees, nil)
	copy(trees[i+1:], trees[i:])
	trees[i] = t

	return trees
}
