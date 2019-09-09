package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// AddFederatedServer is a directory block entry which instructs the system to add a federated server at a specified directory block height
type AddFederatedServer struct {
	AdminIDType     uint32           `json:"adminidtype"`     // the type of action in this admin block entry: uint32(TYPE_ADD_FED_SERVER)
	IdentityChainID interfaces.IHash `json:"identitychainid"` // The federated server id to be added
	DBHeight        uint32           `json:"dbheight"`        // The directory block height this action is to be executed
}

var _ interfaces.IABEntry = (*AddFederatedServer)(nil)
var _ interfaces.BinaryMarshallable = (*AddFederatedServer)(nil)

// Init initializes the identity chain id to the zero hash if it is currently nil
func (e *AddFederatedServer) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

// String writes this AddFederatedServer to a string
func (e *AddFederatedServer) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %35s -- %17s %8x %12s %8d",
		"AddFedServer",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"DBHeight",
		e.DBHeight))
	return (string)(out.DeepCopyBytes())
}

// UpdateState updates factomd's state with instructions on adding this federated server
func (e *AddFederatedServer) UpdateState(state interfaces.IState) error {
	e.Init()
	if e.DBHeight == 1 {
		//use the bootstrap identity for the process list following the genesis block
		id := state.GetNetworkBootStrapIdentity()
		state.AddFedServer(e.DBHeight, id)
	} else {
		state.AddFedServer(e.DBHeight, e.IdentityChainID)
	}
	authorityDeltaString := fmt.Sprintf("AdminBlock (AddFedMsg DBHt: %d) \n ^ %s", e.DBHeight, e.IdentityChainID.String()[5:10])
	state.AddStatus(authorityDeltaString)
	state.AddAuthorityDelta(authorityDeltaString)
	state.UpdateAuthorityFromABEntry(e)
	return nil
}

// NewAddFederatedServer creates a new AddFederatedServer object with the inputs
func NewAddFederatedServer(identityChainID interfaces.IHash, dbheight uint32) (e *AddFederatedServer) {
	if identityChainID == nil {
		return nil
	}
	e = new(AddFederatedServer)
	e.DBHeight = dbheight
	e.IdentityChainID = primitives.NewHash(identityChainID.Bytes())
	return
}

// Type always returns TYPE_ADD_FED_SERVER
func (e *AddFederatedServer) Type() byte {
	return constants.TYPE_ADD_FED_SERVER
}

// MarshalBinary marshals the AddFederatedServer to a byte array
func (e *AddFederatedServer) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AddFederatedServer.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	var buf primitives.Buffer

	err = buf.PushByte(e.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.IdentityChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.DBHeight)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinaryData unmarshals the input data into this AddFederatedServer
func (e *AddFederatedServer) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	b, err := buf.PopByte()
	if b != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}

	e.IdentityChainID = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(e.IdentityChainID)
	if err != nil {
		return nil, err
	}
	e.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

// UnmarshalBinary unmarshals the input data into this AddFederatedServer
func (e *AddFederatedServer) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *AddFederatedServer) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *AddFederatedServer) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *AddFederatedServer) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *AddFederatedServer) Interpret() string {
	return ""
}

// Hash marshals the AddFederatedServer and returns its hash
func (e *AddFederatedServer) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddFederatedServer.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
