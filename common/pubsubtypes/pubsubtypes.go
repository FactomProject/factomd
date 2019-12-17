package pubsubtypes

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

type DBHT struct {
	DBHT   int64
	Minute byte
}

type IMsg interfaces.IMsg
type Hash interfaces.IHash
type Timestamp interfaces.Timestamp

type CommitRequest struct {
	IMsg    IMsg
	Channel chan error
}
