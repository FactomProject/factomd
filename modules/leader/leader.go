package leader

import (
	"encoding/binary"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/state"
)

var logfile = "fnode0_leader"

type Leader struct {
	Pub
	Sub
	*Events     // events indexed by VM
	VMIndex int // vm this leader is responsible fore
	exit    chan interface{}
	ticker  chan interface{}
}

// initialize the leader event aggregate
func New(s *state.State) *Leader {
	l := new(Leader)
	l.VMIndex = s.LeaderVMIndex
	l.ticker = make(chan interface{})
	l.exit = make(chan interface{})

	l.Events = &Events{
		Config: &event.LeaderConfig{
			Salt:               s.Salt,
			IdentityChainID:    s.IdentityChainID,
			ServerPrivKey:      s.ServerPrivKey,
			BlocktimeInSeconds: s.DirectoryBlockInSeconds,
		},
		DBHT: &event.DBHT{ // moved to new height/min
			DBHeight: s.DBHeightAtBoot,
			Minute:   0,
		},
		Balance:   nil, // last perm balance computed
		Directory: nil, // last dblock created
		Ack:       nil, // last ack
	}

	return l
}

func (l *Leader) Sign(b []byte) interfaces.IFullSignature {
	return l.Config.ServerPrivKey.Sign(b)
}

func (l *Leader) sendOut(msg interfaces.IMsg) {
	log.LogMessage(logfile, "sendout", msg)
	l.Pub.MsgOut.Write(msg)
}

// Returns a millisecond timestamp
func (l *Leader) getTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampNow()
}

func (l *Leader) getSalt(ts interfaces.Timestamp) uint32 {
	var b [32]byte
	copy(b[:], l.Config.Salt.Bytes())
	binary.BigEndian.PutUint64(b[:], uint64(ts.GetTimeMilli()))
	c := primitives.Sha(b[:])
	return binary.BigEndian.Uint32(c.Bytes())
}

func (l *Leader) sendAck(m interfaces.IMsg) {
	ack := l.NewAck(m, l.BalanceHash).(*messages.Ack) // LeaderExecute
	l.sendOut(ack)
}
