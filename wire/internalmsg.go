package wire

import ()

// Commands used in bitcoin message headers which describe the type of message.
const (

	// Factom internal messages:
	CmdInt_FactoidObj = "int_factoidobj"
	CmdInt_EOM        = "int_eom"
)

// FtmInternalMsg is an interface that describes an internal factom message.
// The message is used for communication between two modules
type FtmInternalMsg interface {
	Command() string
}

// Factoid Obj to carry factoid transation data to constuct the Process lit item.
type MsgInt_FactoidObj struct {
	FactoidTx    *MsgTx
	TxSha        *ShaHash
	EntryCredits map[ShaHash]uint64 // TODO: this should really be a single-hash per Brian (?)
}

// Factoid Obj to carry factoid transation data to constuct the Process lit item.
// func (msg *MsgInt_FactoidObj) Command() string {
func (msg MsgInt_FactoidObj) Command() string {
	return CmdInt_FactoidObj
}

// End-of-Minute internal message for time commnunications between Goroutines
type MsgInt_EOM struct {
	EOM_Type         byte
	NextDBlockHeight uint64
}

// End-of-Minute internal message for time commnunications between Goroutines
func (msg *MsgInt_EOM) Command() string {
	return CmdInt_EOM
}
