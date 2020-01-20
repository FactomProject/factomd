package leader

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

// REVIEW: there may be another method needed to create eom
// when reporting a dropped message that triggers an election
func (l *Leader) createEOM() (eom *messages.EOM, ack interfaces.IMsg) {
	eom = new(messages.EOM)

	eom.Timestamp = l.getTimestamp()
	eom.ChainID = l.Config.IdentityChainID
	eom.Sign(l)
	eom.SetLocal(false)

	// Put the System Height and Serial Hash into the EOM
	eom.SysHeight = 0

	eom.DBHeight = l.DBHT.DBHeight
	eom.VMIndex = l.VMIndex
	// EOM.Minute is zerobased, while LeaderMinute is 1 based.  So
	// a simple assignment works.
	eom.Minute = byte(l.DBHT.Minute)

	// REVIEW: do we need to record this?
	//vm.EomMinuteIssued = l.DBHT.Minute + 1

	eom.Sign(l)
	eom.MsgHash = nil
	ack = l.NewAck(eom, nil).(*messages.Ack)
	eom.MsgHash = nil
	eom.RepeatHash = nil
	return eom, ack
}

func (l *Leader) sendEOM() bool {
	ack, eom := l.createEOM()
	l.sendOut(ack)
	l.sendOut(eom)
	return true
}
