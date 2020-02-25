package leader

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/modules/event"
)

// Create a new Acknowledgement.  Must be called by a leader.  This
// call assumes all the pieces are in place to create a new acknowledgement
func (l *Leader) NewAck(msg interfaces.IMsg, balanceHash interfaces.IHash) interfaces.IMsg {
	// these don't affect the msg hash, just for local use...
	msg.SetLeaderChainID(l.Config.IdentityChainID)
	ack := new(messages.Ack)
	ack.DBHeight = l.DBHT.DBHeight
	ack.VMIndex = l.VMIndex
	ack.Minute = byte(l.DBHT.Minute)
	ack.Timestamp = l.getTimestamp()
	ack.SaltNumber = l.getSalt(ack.Timestamp)
	copy(ack.Salt[:8], l.Config.Salt.Bytes()[:8])
	ack.MessageHash = msg.GetMsgHash()
	ack.LeaderChainID = l.Config.IdentityChainID
	ack.BalanceHash = balanceHash

	if l.Ack != nil {
		ack.Height = l.Ack.Height + 1
		ack.SerialHash, _ = primitives.CreateHash(l.Ack.MessageHash, ack.MessageHash)
	} else {
		ack.Height = 0
		ack.SerialHash = ack.MessageHash
	}

	// keep height and hash from latest ack
	l.Ack = &events.Ack{
		Height:      ack.Height,
		MessageHash: ack.MessageHash,
	}

	ack.Sign(l)
	return ack
}
