package leader

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

func (l *Leader) CreateDBSig() (interfaces.IMsg, interfaces.IMsg) {

	dbs := new(messages.DirectoryBlockSignature)
	dbs.DirectoryBlockHeader = l.Directory.DirectoryBlockHeader
	dbs.ServerIdentityChainID = l.Config.IdentityChainID
	dbs.DBHeight = l.Directory.DBHeight
	dbs.Timestamp = l.GetTimestamp()
	dbs.SetVMHash(nil)
	dbs.SetVMIndex(l.VMIndex)
	dbs.SetLocal(true)
	dbs.Sign(l)
	err := dbs.Sign(l)
	if err != nil {
		panic(err)
	}
	ack := l.NewAck(dbs, l.Balance.BalanceHash).(*messages.Ack)

	//l.LogMessage("dbstateprocess", "CreateDBSig", dbs)
	return dbs, ack
}

func (l *Leader) SendDBSig() {
	if l.Ack != nil {
		l.Ack.Height = 0 // reset only pl height
	}
	dbs, ack := l.CreateDBSig()
	l.SendOut(dbs)
	l.SendOut(ack)
}
