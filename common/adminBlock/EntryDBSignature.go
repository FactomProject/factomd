package adminBlock

import (
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

func (c *DBSignatureEntry) UpdateState(state interfaces.IState) error {
	return fmt.Errorf("Should not be called alone!")
	//return nil
}

// Create a new DB Signature Entry
func NewDBSignatureEntry(identityAdminChainID interfaces.IHash, sig interfaces.IFullSignature) (*DBSignatureEntry, error) {
	e := new(DBSignatureEntry)
	e.IdentityAdminChainID = identityAdminChainID
	bytes, err := sig.MarshalBinary()
	if err != nil {
		return nil, err
	}
	prevDBSig := new(primitives.Signature)
	prevDBSig.SetPub(bytes[:32])
	err = prevDBSig.SetSignature(bytes[32:])
	if err != nil {
		return nil, err
	}
	e.PrevDBSig = *prevDBSig
	return e, nil
}

func (e *DBSignatureEntry) Type() byte {
	return constants.TYPE_DB_SIGNATURE
}

func (e *DBSignatureEntry) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

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

	return buf.DeepCopyBytes(), nil
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

func (e *DBSignatureEntry) String() string {
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %8x %12s %8s %12s %8x",
		"DB Signature",
		"IdentityChainID", e.IdentityAdminChainID.Bytes()[3:5],
		"PubKey", e.PrevDBSig.Pub.String()[:8],
		"Signature", e.PrevDBSig.Sig.String()[:8]))
	return (string)(out.DeepCopyBytes())
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
