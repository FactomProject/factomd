package electionMsgs

import "github.com/FactomProject/factomd/common/interfaces"

type ElectionsFactory struct{}

var _ interfaces.IElectionsFactory = (*ElectionsFactory)(nil)

func (e *ElectionsFactory) NewAddLeaderInternal(name string, dbheight uint32, serverID interfaces.IHash) interfaces.IMsg {
	msg := new(RemoveAuditInternal)
	msg.NName = name
	msg.DBHeight = dbheight
	msg.ServerID = serverID
	return msg
}
func (e *ElectionsFactory) NewAddAuditInternal(name string, dbheight uint32, serverID interfaces.IHash) interfaces.IMsg {
	msg := new(RemoveAuditInternal)
	msg.NName = name
	msg.DBHeight = dbheight
	msg.ServerID = serverID
	return msg
}
func (e *ElectionsFactory) NewRemoveLeaderInternal(name string, dbheight uint32, serverID interfaces.IHash) interfaces.IMsg {
	msg := new(RemoveAuditInternal)
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
func (e *ElectionsFactory) NewEomSigInternal(name string, dbheight uint32, minute uint32, height uint32, serverID interfaces.IHash) interfaces.IMsg {
	msg := new(EomSigInternal)
	msg.NName = name
	msg.DBHeight = dbheight
	msg.Minute = minute
	msg.ServerID = serverID
	return msg
}
