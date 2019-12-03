package leader

import (
	"encoding/binary"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	llog "github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/state"
	"sync"
)

var log = llog.PackageLogger

type Leader struct {
	Pub
	Sub
	*Events          // events indexed by VM
	VMIndex      int // vm this leader is responsible fore
	EOMSyncEnd   int64
	EOMIssueTime int64
	loaded sync.WaitGroup
}

// initialize the leader event aggregate
func New(s *state.State) *Leader {
	// TODO: track Db Height so we can decide whether to send out dbsigs
	l := new(Leader)
	l.loaded.Add(1)
	l.VMIndex = s.LeaderVMIndex

	l.Events = &Events{
		Config: &event.LeaderConfig{
			Salt:            s.Salt,
			IdentityChainID: s.IdentityChainID,
			ServerPrivKey:   s.ServerPrivKey,
			FactomSecond:    s.FactomSecond(),
		},
		DBHT: &event.DBHT{
			DBHeight: s.DBHeightAtBoot,
			VMIndex:  s.LeaderVMIndex,
			Minute:   0,
		},
		Balance:   nil, // last perm balance computed
		Directory: nil, // last dblock created
		Ack:       nil, // last ack
	}

	return l
}

func (l *Leader) SendOut(msg interfaces.IMsg) {
	log.LogMessage("leader.txt", "sendout", msg)
	l.Pub.MsgOut.Write(msg)
}

func (l *Leader) Sign(b []byte) interfaces.IFullSignature {
	return l.Config.ServerPrivKey.Sign(b)
}

// Returns a millisecond timestamp
func (l *Leader) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (l *Leader) GetSalt(ts interfaces.Timestamp) uint32 {
	var b [32]byte
	copy(b[:], l.Config.Salt.Bytes())
	binary.BigEndian.PutUint64(b[:], uint64(ts.GetTimeMilli()))
	c := primitives.Sha(b[:])
	return binary.BigEndian.Uint32(c.Bytes())
}

func (l *Leader) sendAck(m interfaces.IMsg) {
	ack := l.NewAck(m, l.BalanceHash).(*messages.Ack) // LeaderExecute
	l.SendOut(ack)
}
