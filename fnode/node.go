package fnode

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
)

type FactomNode struct {
	Index    int
	State    *state.State
	Peers    []interfaces.IPeer
	P2PIndex int
}

var fnodes []*FactomNode

func GetFnodes() []*FactomNode {
	return fnodes
}

func AddFnode(node *FactomNode) {
	fnodes = append(fnodes, node)
}

func Get(i int) *FactomNode {
	return fnodes[i]
}

func Len() int {
	return len(fnodes)
}
