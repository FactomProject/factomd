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
	Init(p NamedObject, n string)
	String() string
	// Debug
	AddChild(NamedObject)
	GetChildren() []NamedObject
}

type Name struct {
	parent NamedObject
	path   string
	name   string
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

var root Name = Name{NamedObject(nil), "/", "/", "namedobjects.go:56", nil}

func (t *Name) GetChildren() []NamedObject {
	return t.children
}

func (t *Name) AddChild(kid NamedObject) {
	if t != nil {
		t.children = append(t.children, kid)
	} else {
		root.AddChild(kid)
	}
}

func (t *Name) Init(p NamedObject, n string) {
	if t.parent != nil {
		panic("Already Inited by " + t.by)

	}
	t.parent = p
	t.path = t.parent.GetName() + "/" + n
	t.name = n
	// debug
	t.by = atomic.WhereAmIString(1)
	t.parent.AddChild(t)

}

var NilName *Name                // This is a nil of name type Do NOT write it!
var _ NamedObject = (*Name)(nil) // Check that the interface is met

func PrintNames(i int, n NamedObject) {
	fmt.Printf("%*s %s\n", 3*i, "", n.GetPath())
	for _, kid := range n.GetChildren() {
		PrintNames(i+1, kid)
	}
}

func PrintAllNames() {
	PrintNames(0, &root)
}
