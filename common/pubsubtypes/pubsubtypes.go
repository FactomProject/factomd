package pubsubtypes

import "github.com/FactomProject/factomd/common/interfaces"

type DBHT struct {
	DBHT   int64
	Minute byte
}

type IMsg struct{ interfaces.IMsg }
