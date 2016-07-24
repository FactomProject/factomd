package adminBlock

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DB Signature Entry -------------------------
type AddAuditServer struct {
	IdentityChainID interfaces.IHash
	DBHeight        uint32
}

var _ interfaces.IABEntry = (*AddAuditServer)(nil)
var _ interfaces.BinaryMarshallable = (*AddAuditServer)(nil)

func (e *AddAuditServer) String() string {
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %8x %12s %8d\n", "AddAuditServer", "IdentityChainID", e.IdentityChainID.Bytes()[:4], "DBHeight", e.DBHeight))
	return (string)(out.DeepCopyBytes())
}

func (c *AddAuditServer) UpdateState(state interfaces.IState) {
	state.AddAuditServer(c.DBHeight, c.IdentityChainID)
	state.UpdateAuthorityFromABEntry(c)
}

// Create a new DB Signature Entry
func NewAddAuditServer(identityChainID interfaces.IHash, dbheight uint32) (e *AddAuditServer) {
	e = new(AddAuditServer)
	e.DBHeight = dbheight
	e.IdentityChainID = primitives.NewHash(identityChainID.Bytes())
	return
}

func (e *AddAuditServer) Type() byte {
	return constants.TYPE_ADD_AUDIT_SERVER
}

func (e *AddAuditServer) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	buf.Write([]byte{e.Type()})

	data, err = e.IdentityChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	binary.Write(&buf, binary.BigEndian, e.DBHeight)

	return buf.DeepCopyBytes(), nil
}

func (e *AddAuditServer) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling Add Federated Server Entry: %v", r)
		}
	}()

	newData = data
	newData = newData[1:]

	e.IdentityChainID = new(primitives.Hash)
	newData, err = e.IdentityChainID.UnmarshalBinaryData(newData)
	if err != nil {
		panic(err.Error())
	}

	e.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	return
}

func (e *AddAuditServer) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *AddAuditServer) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *AddAuditServer) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *AddAuditServer) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *AddAuditServer) IsInterpretable() bool {
	return false
}

func (e *AddAuditServer) Interpret() string {
	return ""
}

func (e *AddAuditServer) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
