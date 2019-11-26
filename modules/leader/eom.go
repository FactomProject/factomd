package leader

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

// REVIEW: there may be another method needed to create eom
// when reporting a dropped message that triggers an election
func (l *Leader) CreateEOM() (eom *messages.EOM, ack interfaces.IMsg) {
	eom = new(messages.EOM)

	eom.Timestamp = l.GetTimestamp()
	eom.ChainID = l.Config.IdentityChainID
	eom.Sign(l)
	eom.SetLocal(true)

	// Put the System Height and Serial Hash into the EOM
	eom.SysHeight = uint32(l.SysHeight)

	eom.DBHeight = l.DBHT.DBHeight
	eom.VMIndex = l.VMIndex
	// EOM.Minute is zerobased, while LeaderMinute is 1 based.  So
	// a simple assignment works.
	eom.Minute = byte(l.Minute)

	// REVIEW: do we need to record this?
	//vm.EomMinuteIssued = l.DBHT.Minute + 1

	eom.Sign(l)
	eom.MsgHash = nil
	ack = l.NewAck(eom, nil).(*messages.Ack)
	eom.MsgHash = nil
	eom.RepeatHash = nil
	return eom, ack
}

func (l *Leader) SendEOM() {
	ack, eom := l.CreateEOM()
	l.SendOut(ack)
	l.SendOut(eom)
}
