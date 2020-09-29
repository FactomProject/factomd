package pubsubtypes

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type DBHT struct {
	DBHT   int64
	Minute byte
}

type IMsg struct{ interfaces.IMsg }
type Hash primitives.Hash
