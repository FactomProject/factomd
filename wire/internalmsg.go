package wire

import ()

// Commands used in bitcoin message headers which describe the type of message.
const (

	// Factom internal messages:
	CmdInt_FactoidObj    = "int_factoidobj"
	CmdInt_FactoidTxHash = "int_txhash"
	CmdInt_EOM           = "int_eom"
)

// FtmInternalMsg is an interface that describes an internal factom message.
// The message is used for communication between two modules
type FtmInternalMsg interface {
	Command() string
}

type EntryCreditMap struct {
	pubkey ShaHash
	value  uint64
}

// Factoid Obj to carry factoid transation data to constuct the Process lit item.
type MsgInt_FactoidObj struct {
	FactoidTx MsgTx // Jack: get the TX hash this way: FactoidTx.TxSha()
	EC_map    []EntryCreditMap
}

// Factoid Obj to carry factoid transation data to constuct the Process lit item.
// func (msg *MsgInt_FactoidObj) Command() string {
func (msg MsgInt_FactoidObj) Command() string {
	return CmdInt_FactoidObj
}

/* SEE ABOVE: FactoidTx.TxSha()
// Factoid transaction hash for commnunications between Goroutines
type MsgInt_TxHash struct {
	Hash *ShaHash
}

// Factoid transaction hash for commnunications between Goroutines
func (msg *MsgInt_TxHash) Command() string {
	return CmdInt_FactoidTxHash
}
*/

// End-of-Minute internal message for time commnunications between Goroutines
type MsgInt_EOM struct {
	EOM_Type byte
}

// End-of-Minute internal message for time commnunications between Goroutines
func (msg *MsgInt_EOM) Command() string {
	return CmdInt_EOM
}
