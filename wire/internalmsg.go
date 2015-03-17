package wire

import (
)


// Commands used in bitcoin message headers which describe the type of message.
const (

	// Factom internal messages:
	CmdInt_AddPLI     = "int_addpli"
)


// FtmInternalMsg is an interface that describes an internal factom message. 
// The message is used for communication between two modules 
type FtmInternalMsg interface {
	Command() string
}


// response to a getdata message (MsgGetData) for a given block hash.
type MsgInt_PLI struct {
	Transactions []*MsgTx
}

// AddTransaction adds a transaction to the message.
func (msg *MsgInt_PLI) Command() string {
	return CmdInt_AddPLI
} 