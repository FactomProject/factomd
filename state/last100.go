package state

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

type Last100 struct {
	lookup map[[32]byte]interfaces.IMsg
	n      int
	list   [100]interfaces.IMsg
}

func (list *Last100) Add(msg interfaces.IMsg) {
	if list.lookup == nil {
		list.lookup = make(map[[32]byte]interfaces.IMsg)
	}
	old := list.list[list.n]
	if old != nil {
		delete(list.lookup, old.GetHash().Fixed())
	}
	list.list[list.n] = msg
	list.lookup[msg.GetHash().Fixed()] = msg
	list.n = (list.n + 1) % len(list.list)
}

func (list *Last100) Get(h [32]byte) interfaces.IMsg {
	return list.lookup[h]
}
