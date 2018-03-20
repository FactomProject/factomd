package electionMsgs

import (
	"github.com/FactomProject/factomd/common/interfaces"
)

type ElectionsFactory struct{}

var _ interfaces.IElectionsFactory = (*ElectionsFactory)(nil)

//func (e *ElectionsFactory) NewElectionAdapter(el interfaces.IElections) interfaces.IElectionAdapter {
//	return NewElectionAdapter(el.(*elections.Elections))
//}

func (e *ElectionsFactory) NewAddLeaderInternal(name string, dbheight uint32, serverID interfaces.IHash) interfaces.IMsg {
	msg := new(AddLeaderInternal)
	msg.NName = name
	msg.DBHeight = dbheight
	msg.ServerID = serverID
	return msg
}
func (e *ElectionsFactory) NewAddAuditInternal(name string, dbheight uint32, serverID interfaces.IHash) interfaces.IMsg {
	msg := new(AddAuditInternal)
	msg.NName = name
	msg.DBHeight = dbheight
	msg.ServerID = serverID
	return msg
}
func (e *ElectionsFactory) NewRemoveLeaderInternal(name string, dbheight uint32, serverID interfaces.IHash) interfaces.IMsg {
	msg := new(RemoveLeaderInternal)
	msg.NName = name
	msg.DBHeight = dbheight
	msg.ServerID = serverID
	return msg
}
func (e *ElectionsFactory) NewRemoveAuditInternal(name string, dbheight uint32, serverID interfaces.IHash) interfaces.IMsg {
	msg := new(RemoveAuditInternal)
	msg.NName = name
	msg.DBHeight = dbheight
	msg.ServerID = serverID
	return msg
}
func (e *ElectionsFactory) NewEomSigInternal(name string, dbheight uint32, minute uint32, vmIndex int, height uint32, serverID interfaces.IHash) interfaces.IMsg {
	msg := new(EomSigInternal)
	msg.SigType = true
	msg.NName = name
	msg.DBHeight = dbheight
	msg.Minute = byte(minute)
	msg.VMIndex = vmIndex
	msg.ServerID = serverID
	return msg
}
func (e *ElectionsFactory) NewDBSigSigInternal(name string, dbheight uint32, minute uint32, vmIndex int, height uint32, serverID interfaces.IHash) interfaces.IMsg {
	msg := new(EomSigInternal)
	msg.SigType = false
	msg.NName = name
	msg.DBHeight = dbheight
	msg.Minute = byte(minute)
	msg.VMIndex = vmIndex
	msg.ServerID = serverID
	return msg
}
