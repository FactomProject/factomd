package adminBlock

import (
	"bytes"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DB Signature Entry -------------------------
type DBSignatureEntry struct {
	IdentityAdminChainID interfaces.IHash
	PrevDBSig            primitives.Signature
}

var _ interfaces.IABEntry = (*DBSignatureEntry)(nil)
var _ interfaces.BinaryMarshallable = (*DBSignatureEntry)(nil)

func (c *DBSignatureEntry) UpdateState(state interfaces.IState) {

}

// Create a new DB Signature Entry
func NewDBSignatureEntry(identityAdminChainID interfaces.IHash, sig primitives.Signature) (e *DBSignatureEntry) {
	e = new(DBSignatureEntry)
	e.IdentityAdminChainID = identityAdminChainID
	copy(e.PrevDBSig.Pub[:], sig.Pub[:])
	copy(e.PrevDBSig.Sig[:], sig.Sig[:])
	return
}

func (e *DBSignatureEntry) Type() byte {
	return constants.TYPE_DB_SIGNATURE
}

func (e *DBSignatureEntry) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	buf.Write([]byte{e.Type()})

	data, err = e.IdentityAdminChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	_, err = buf.Write(e.PrevDBSig.Pub[:])
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(e.PrevDBSig.Sig[:])
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (e *DBSignatureEntry) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshallig DBSignature Entry: %v", r)
		}
	}()
	newData = data
	if newData[0] != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}
	newData = newData[1:]

	e.IdentityAdminChainID = new(primitives.Hash)
	newData, err = e.IdentityAdminChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}

	newData, err = e.PrevDBSig.UnmarshalBinaryData(newData)

	return
}

func (e *DBSignatureEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *DBSignatureEntry) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *DBSignatureEntry) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *DBSignatureEntry) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *DBSignatureEntry) String() string {
	str, _ := e.JSONString()
	return str
}

func (e *DBSignatureEntry) IsInterpretable() bool {
	return false
}

func (e *DBSignatureEntry) Interpret() string {
	return ""
}

func (e *DBSignatureEntry) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
