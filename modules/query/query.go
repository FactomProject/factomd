package query

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

type Queryable struct {
	Election interfaces.IElections
}

var nodes = make(map[string]*Queryable)

func Module(node string) *Queryable {
	q := nodes[node]
	if q == nil {
		q = new(Queryable)
		nodes[node] = q
	}
	return q
}
