package fnode

import (
	"fmt"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/worker"
)

type root struct {
	common.Name
}

// factory method to spawn new nodes
var Factory func(w *worker.Thread)

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

func New(s *state.State) *FactomNode {
	n := new(FactomNode)
	n.State = s
	n.Init(Root, "svc") // root of service
	fnodes = append(fnodes, n)
	n.addFnodeName()
	n.State.Init(n, n.State.FactomNodeName)
	return n
}

var fnodes []*FactomNode

func GetFnodes() []*FactomNode {
	return fnodes
}

func AddFnode(node *FactomNode) {
	node.Init(Root, "svc") // root of service
	node.State.Init(node, node.State.FactomNodeName)
	node.State.Hold.Init(node.State, "HoldingList")
	fnodes = append(fnodes, node)
}

func Get(i int) *FactomNode {
	return fnodes[i]
}

func Len() int {
	return len(fnodes)
}

// useful for logging
func (node *FactomNode) addFnodeName() {
	// full name
	name := node.State.FactomNodeName
	globals.FnodeNames[node.State.IdentityChainID.String()] = name

	// common short set
	globals.FnodeNames[fmt.Sprintf("%x", node.State.IdentityChainID.Bytes()[3:6])] = name
	globals.FnodeNames[fmt.Sprintf("%x", node.State.IdentityChainID.Bytes()[:5])] = name
	globals.FnodeNames[fmt.Sprintf("%x", node.State.IdentityChainID.Bytes()[:])] = name
	globals.FnodeNames[fmt.Sprintf("%x", node.State.IdentityChainID.Bytes()[:8])] = name
}
