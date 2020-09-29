package common

import (
	"fmt"

	"github.com/FactomProject/factomd/util/atomic"
)

// Objects that embed a Name and call Init() will get hierarchical names

type NamedObject interface {
	GetName() string
	GetPath() string
	GetParentName() string
	NameInit(p NamedObject, n string, t string)
	String() string
	// Debug
	GetType() string
	AddChild(NamedObject)
	GetChildren() []NamedObject
}

type Name struct {
	parent NamedObject
	path   string
	name   string
	t      string // type
	// debug
	by       string // who init'ed me (file:line) for debug
	children []NamedObject
}

func (n *Name) GetName() string {
	if n == nil {
		return ""
	}
	return n.name
}
func (n *Name) GetType() string {
	if n == nil {
		return "unknown"
	}
	return n.t
}

func (n *Name) GetPath() string {
	if n == nil {
		return ""
	}
	return n.path
}

func (n *Name) GetParentName() string {
	if n == nil {
		return ""
	}
	return n.parent.GetName()
}

func (n *Name) String() string {
	return n.GetPath()
}

var root Name = Name{NamedObject(nil), "/", "/", "root", "namedobjects.go:56", nil}

func (n *Name) GetChildren() []NamedObject {
	return n.children
}

func (n *Name) AddChild(kid NamedObject) {
	if n != nil {
		n.children = append(n.children, kid)
	} else {
		root.AddChild(kid)
	}
}

func (n *Name) NameInit(p NamedObject, name string, t string) {
	if n.parent != nil {
		panic("Already Inited by " + n.by)

	}
	n.parent = p
	n.path = n.parent.GetName() + "/" + name
	n.name = name
	// debug
	n.t = t
	n.by = atomic.WhereAmIString(1)
	n.parent.AddChild(n)

}

var NilName *Name                // This is a nil of name type Do NOT write it!
var _ NamedObject = (*Name)(nil) // Check that the interface is met

func PrintNames(i int, n NamedObject) {
	fmt.Printf("%*s %s-%s\n", 3*i, "", n.GetPath(), n.GetType())
	for _, kid := range n.GetChildren() {
		PrintNames(i+1, kid)
	}
}

func PrintAllNames() {
	PrintNames(0, &root)
}
