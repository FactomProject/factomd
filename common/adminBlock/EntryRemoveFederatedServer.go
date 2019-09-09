package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// RemoveFederatedServer is an admin block entry which instructs factomd to remove a federated server at an upcoming block height
type RemoveFederatedServer struct {
	AdminIDType     uint32           `json:"adminidtype"`     //  the type of action in this admin block entry: uint32(TYPE_REMOVE_FED_SERVER)
	IdentityChainID interfaces.IHash `json:"identitychainid"` // The identity of the federated server to be removed
	DBHeight        uint32           `json:"dbheight"`        // The directory block height when the system should remove the federated server
}

var _ interfaces.IABEntry = (*RemoveFederatedServer)(nil)
var _ interfaces.BinaryMarshallable = (*RemoveFederatedServer)(nil)

// Init initializes any nil hashes in the RemoveFederatedServer to the zero hash and sets AdminIDType to is hardcoded value
func (e *RemoveFederatedServer) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

// String returns the RemoveFederatedServer string
func (e *RemoveFederatedServer) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %35s -- %17s %8x %12s %8d",
		"Remove Federated Server",
		"IdentityChainID",
		e.IdentityChainID.Bytes()[3:6],
		"DBHeight",
		e.DBHeight))
	return (string)(out.DeepCopyBytes())
}

// UpdateState updates the factomd state to removed the federated server at the specific directory block height
func (e *RemoveFederatedServer) UpdateState(state interfaces.IState) error {
	e.Init()
	if len(state.GetFedServers(e.DBHeight)) != 0 {
		state.RemoveFedServer(e.DBHeight, e.IdentityChainID)
	}
	if state.GetOut() {
		state.Println(fmt.Sprintf("Removed Federated Server: %x", e.IdentityChainID.Bytes()[:4]))
	}
	authorityDeltaString := fmt.Sprintf("AdminBlock (RemoveFedMsg DBHt: %d) \n v %s", e.DBHeight, e.IdentityChainID.String()[5:10])
	state.AddStatus(authorityDeltaString)
	state.AddAuthorityDelta(authorityDeltaString)
	state.UpdateAuthorityFromABEntry(e)
	return nil
}

// NewRemoveFederatedServer creates a new RemoveFederatedServer object with the inputs
func NewRemoveFederatedServer(identityChainID interfaces.IHash, dbheight uint32) (e *RemoveFederatedServer) {
	if identityChainID == nil {
		return nil
	}
	e = new(RemoveFederatedServer)
	e.IdentityChainID = primitives.NewHash(identityChainID.Bytes())
	e.DBHeight = dbheight
	return
}

// Type returns the hardcoded TYPE_REMOVE_FED_SERVER
func (e *RemoveFederatedServer) Type() byte {
	return constants.TYPE_REMOVE_FED_SERVER
}

// MarshalBinary marshals this RemoveFederatedServer object
func (e *RemoveFederatedServer) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "RemoveFederatedServer.MarshalBinary err:%v", *pe)
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

// UnmarshalBinaryData unmarshals the input data into this RemoveFederatedServer
func (e *RemoveFederatedServer) UnmarshalBinaryData(data []byte) ([]byte, error) {
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

// UnmarshalBinary unmarshals the input data into this RemoveFederatedServer
func (e *RemoveFederatedServer) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *RemoveFederatedServer) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *RemoveFederatedServer) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *RemoveFederatedServer) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *RemoveFederatedServer) Interpret() string {
	return ""
}

// Hash marshals this RemoveFederatedServer and computes its hash
func (e *RemoveFederatedServer) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("RemoveFederatedServer.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
