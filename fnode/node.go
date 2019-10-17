package fnode

import (
	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
)

type root struct{
	common.Name
}

// root of object hierarchy
var Root = &root{}

func init() {
	Root.Init(Root, "")
}

type FactomNode struct {
	common.Name
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
	node.Init(Root, "svc") // root of service
	node.State.Init(node, node.State.FactomNodeName)

	// FIXME relocate to state ? - this causes issues w/ multiple peers
	//node.State.Hold.Init(node.State, "HoldingList")
	fnodes = append(fnodes, node)
}

func Get(i int) *FactomNode {
	return fnodes[i]
}

func Len() int {
	return len(fnodes)
}
