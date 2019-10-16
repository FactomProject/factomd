package common

// Objects that embed a Name and call Init() will get hierarchical names

type NamedObject interface {
	GetName() string
	Init(p NamedObject, n string)
	String() string
}

type Name struct {
	parent NamedObject
	name   string
}

func (n *Name) GetName() string {
	if n == nil {
		return ""
	}
	return n.name
}

func (n *Name) String() string {
	return n.GetName()
}

func (t *Name) Init(p NamedObject, n string) {
	if t.parent != nil {
		panic("Already Inited")
	}
	t.parent = p
	t.name = t.parent.GetName() + "/" + n
}

var _ NamedObject = (*Name)(nil) // Check that the interface is met
