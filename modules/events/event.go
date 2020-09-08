package events

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"path"
)

type pubSubPaths struct {
	EOM               string
	Seq               string
	Directory         string
	Bank              string
	LeaderConfig      string
	LeaderMsgIn       string
	LeaderMsgOut      string
	CommitChain       string
	CommitEntry       string
	RevealEntry       string
	CommitDBState     string
	DBAnchored        string
	NodeMessage       string
	AuthoritySet      string
	ConnectionMetrics string
	ConnectionAdded   string
	ConnectionRemoved string
	ProcessListInfo   string
	StateUpdate       string
}

var Path = pubSubPaths{
	EOM:               "EOM",
	Seq:               "seq",
	Directory:         "directory",
	Bank:              "bank",
	LeaderConfig:      "leader-config",
	LeaderMsgIn:       "leader-msg-in",
	LeaderMsgOut:      "leader-msg-out",
	CommitChain:       "commit-chain",
	CommitEntry:       "commit-entry",
	RevealEntry:       "reveal-entry",
	CommitDBState:     "commit-dbstate",
	NodeMessage:       "node-message",
	ConnectionMetrics: path.Join("connection", "metrics"), // REVIEW: pubsub has it's own path maybe should use that
	ConnectionAdded:   "connection-added",
	ConnectionRemoved: "connection-removed",
	ProcessListInfo:   "process-list",
	StateUpdate:       "state-update",
	AuthoritySet:      "authority-set",
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
	Minute   int
}

// event created when Ack is actually sent out
type Ack struct {
	Height      uint32
	MessageHash interfaces.IHash
}

type LeaderConfig struct {
	NodeName           string
	IdentityChainID    interfaces.IHash
	Salt               interfaces.IHash // only change on boot
	ServerPrivKey      *primitives.PrivateKey
	BlocktimeInSeconds int
}

type EOM struct {
	Timestamp interfaces.Timestamp
}

type ProcessListInfo struct {
	ProcessTime interfaces.Timestamp
	Dump        string
	PrintMap    string
}

type AuthoritySet struct {
	LeaderHeight uint32
	FedServers   []interfaces.IServer
	AuditServers []interfaces.IServer
}

type ConnectionChanged struct {
	IP       string
	Status   string
	Duration string
	Send     string
	Received string
	IsOnline bool
	State    string
}

type ConnectionAdded struct {
	ConnectionChanged
}

type ConnectionRemoved struct {
	ConnectionChanged
}

type StateUpdate struct {
	NodeTime           interfaces.Timestamp
	LeaderHeight       uint32
	Summary            string
	IdentitiesDetails  string
	AuthoritiesDetails string
}

type CommitChain struct {
	RequestState RequestState
	DBHeight     uint32
	CommitChain  ICommitChain
}

type CommitEntry struct {
	RequestState RequestState
	DBHeight     uint32
	CommitEntry  ICommitEntry
}

type RevealEntry struct {
	RequestState RequestState
	DBHeight     uint32
	RevealEntry  IRevealEntry
	MsgTimestamp interfaces.Timestamp
}

type DBStateCommit struct {
	DBHeight uint32
	DBState  IDBState
}

type DBAnchored struct {
	DBHeight     uint32
	DirBlockInfo interfaces.IDirBlockInfo
}
