package common

// Objects that embed a Name and call Init() will get hierarchical names

type NamedObject interface {
	GetName() string
	GetPath() string
	GetParentName() string
	Init(p NamedObject, n string)
	String() string
}

type Name struct {
	parent NamedObject
	path   string
	name   string
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

func (t *Name) Init(p NamedObject, n string) {
	if t.parent != nil {
		panic("Already Inited")
	}
	t.parent = p
	t.path = t.parent.GetName() + "/" + n
	t.name = n
}

var NilName *Name                // This is a nil of name type Do NOT write it!
var _ NamedObject = (*Name)(nil) // Check that the interface is met
