package adminBlock

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DB Signature Entry -------------------------
type AddAuditServer struct {
	AdminIDType     uint32           `json:"adminidtype"`
	IdentityChainID interfaces.IHash `json:"identitychainid"`
	DBHeight        uint32           `json:"dbheight"`
}

var _ interfaces.IABEntry = (*AddAuditServer)(nil)
var _ interfaces.BinaryMarshallable = (*AddAuditServer)(nil)

func (e *AddAuditServer) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

func (e *AddAuditServer) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %8x %12s %8d",
		"AddAuditServer",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"DBHeight", e.DBHeight))
	return (string)(out.DeepCopyBytes())
}

func (c *AddAuditServer) UpdateState(state interfaces.IState) error {
	c.Init()
	state.AddAuditServer(c.DBHeight, c.IdentityChainID)
	authorityDeltaString := fmt.Sprintf("AdminBlock (AddAudMsg DBHt: %d) \n v %s", c.DBHeight, c.IdentityChainID.String()[5:10])
	state.AddStatus(authorityDeltaString)
	state.AddAuthorityDelta(authorityDeltaString)
	state.UpdateAuthorityFromABEntry(c)
	return nil
}

// Create a new DB Signature Entry
func NewAddAuditServer(identityChainID interfaces.IHash, dbheight uint32) (e *AddAuditServer) {
	if identityChainID == nil {
		return nil
	}
	e = new(AddAuditServer)
	e.DBHeight = dbheight
	e.IdentityChainID = primitives.NewHash(identityChainID.Bytes())
	return
}

func (e *AddAuditServer) Type() byte {
	return constants.TYPE_ADD_AUDIT_SERVER
}

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

func (e *AddAuditServer) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *AddAuditServer) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *AddAuditServer) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *AddAuditServer) IsInterpretable() bool {
	return false
}

func (e *AddAuditServer) Interpret() string {
	return ""
}

func (e *AddAuditServer) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AddAuditServer.Hash") }()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
