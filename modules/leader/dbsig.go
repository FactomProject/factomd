package leader

import (
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

func (l *Leader) createDBSig() (interfaces.IMsg, interfaces.IMsg) {

	dbs := new(messages.DirectoryBlockSignature)
	dbs.DirectoryBlockHeader = l.Directory.DirectoryBlockHeader
	dbs.ServerIdentityChainID = l.Config.IdentityChainID
	dbs.DBHeight = l.DBHT.DBHeight
	dbs.Timestamp = l.getTimestamp()
	dbs.SetVMHash(nil)
	dbs.SetVMIndex(l.VMIndex)
	dbs.SetLocal(true)
	dbs.Sign(l)
	err := dbs.Sign(l)
	if err != nil {
		panic(err)
	}
	ack := l.NewAck(dbs, l.Balance.BalanceHash).(*messages.Ack)

	return dbs, ack
}

func (l *Leader) sendDBSig() {
	l.Ack = nil // reset last known Ack
	dbs, ack := l.createDBSig()
	l.sendOut(dbs)
	l.sendOut(ack)
}
