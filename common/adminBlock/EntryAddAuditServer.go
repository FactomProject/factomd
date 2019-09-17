package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// AddAuditServer is an admin block entry which instructs factomd to add a new audit server at a specified directory block height
type AddAuditServer struct {
	AdminIDType     uint32           `json:"adminidtype"`     // the type of action in this admin block entry: uint32(TYPE_ADD_AUDIT_SERVER)
	IdentityChainID interfaces.IHash `json:"identitychainid"` // the server identity of the new audit server to be added
	DBHeight        uint32           `json:"dbheight"`        // the directory block height when the new audit server should be added
}

var _ interfaces.IABEntry = (*AddAuditServer)(nil)
var _ interfaces.BinaryMarshallable = (*AddAuditServer)(nil)

// Init sets all nil hashs to the zero hash and sets the hardcoded AdminIDType
func (e *AddAuditServer) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

// String returns the AddAuditServer string
func (e *AddAuditServer) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %8x %12s %8d",
		"AddAuditServer",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"DBHeight", e.DBHeight))
	return (string)(out.DeepCopyBytes())
}

// UpdateState updates the factomd state to include the new audit server information
func (e *AddAuditServer) UpdateState(state interfaces.IState) error {
	e.Init()
	state.AddAuditServer(e.DBHeight, e.IdentityChainID)
	authorityDeltaString := fmt.Sprintf("AdminBlock (AddAudMsg DBHt: %d) \n v %s", e.DBHeight, e.IdentityChainID.String()[5:10])
	state.AddStatus(authorityDeltaString)
	state.AddAuthorityDelta(authorityDeltaString)
	state.UpdateAuthorityFromABEntry(e)
	return nil
}

// NewAddAuditServer creates a new AddAuditServer object with the specified inputs
func NewAddAuditServer(identityChainID interfaces.IHash, dbheight uint32) (e *AddAuditServer) {
	if identityChainID == nil {
		return nil
	}
	e = new(AddAuditServer)
	e.DBHeight = dbheight
	e.IdentityChainID = primitives.NewHash(identityChainID.Bytes())
	return
}

// Type returns the hardcoded TYPE_ADD_AUDIT_SERVER
func (e *AddAuditServer) Type() byte {
	return constants.TYPE_ADD_AUDIT_SERVER
}

// MarshalBinary marshals the AddAuditServer
func (e *AddAuditServer) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AddAuditServer.MarshalBinary err:%v", *pe)
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

// UnmarshalBinaryData unmarshals the input data into this AddAuditServer
func (e *AddAuditServer) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	b, err := buf.PopByte()
	if err != nil {
		return nil, err
	}

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

// UnmarshalBinary unmarshals the input data into this AddAuditServer
func (e *AddAuditServer) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *AddAuditServer) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the json encoded string
func (e *AddAuditServer) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// IsInterpretable always returns false
func (e *AddAuditServer) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *AddAuditServer) Interpret() string {
	return ""
}

// Hash marshals the AddAuditServer and computes its hash
func (e *AddAuditServer) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("AddAuditServer.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
