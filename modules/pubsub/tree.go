package pubsub

import (
	"path/filepath"
	"strings"

	"github.com/DiSiqueira/GoTree"
)

// Just an example implementation of using filepath syntax
// to store and explore publishers.

func (r *Registry) PrintTree() string {
	return r.tree.Print()
}

func (r *Registry) AddPath(path string) {
	r.addFromNode(path, r.tree)
}

func (r *Registry) addFromNode(path string, node gotree.Tree) {
	list := strings.Split(path, string(filepath.Separator))
	if len(list) == 1 {
		node.Add(list[0])
		return
	}

	dir := list[0]
	for _, item := range node.Items() {
		if item.Text() == dir {
			r.addFromNode(filepath.Join(list[1:]...), item)
			return
		}
	}

	item := node.Add(dir)
	r.addFromNode(filepath.Join(list[1:]...), item)
}
