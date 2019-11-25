package event

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"time"
)

type pubSubPaths struct {
	EOM          string
	Seq          string
	Directory    string
	Bank         string
	LeaderConfig string
}

var Path = pubSubPaths{
	EOM:          "EOM",
	Seq:          "seq",
	Directory:    "directory",
	Bank:         "bank",
	LeaderConfig: "leader-config",
}

type Balance struct {
	DBHeight    uint32
	BalanceHash interfaces.IHash
}

type Directory struct {
	DBHeight             uint32
	VMIndex              int
	DirectoryBlockHeader interfaces.IDirectoryBlockHeader
	Timestamp            interfaces.Timestamp
}

type DBHT struct {
	DBHeight uint32
	Min      int
}

// event created when Ack is actually sent out
type Ack struct {
	Height      uint32
	MessageHash interfaces.IHash
}

type LeaderConfig struct {
	IdentityChainID interfaces.IHash
	Salt            interfaces.IHash // only change on boot
	ServerPrivKey   *primitives.PrivateKey
	FactomSecond    time.Duration // only change in simulator
}

type EOM struct {
	Timestamp     interfaces.Timestamp
	LLeaderHeight uint32
	SysHeight     uint32
	VMIndex       int
	Minute        byte
}
