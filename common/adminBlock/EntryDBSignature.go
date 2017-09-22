package adminBlock

import (
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DB Signature Entry -------------------------
type DBSignatureEntry struct {
	IdentityAdminChainID interfaces.IHash     `json:"identityadminchainid"`
	PrevDBSig            primitives.Signature `json:"prevdbsig"`
}

var _ interfaces.IABEntry = (*DBSignatureEntry)(nil)
var _ interfaces.BinaryMarshallable = (*DBSignatureEntry)(nil)

func (e *DBSignatureEntry) Init() {
	if e.IdentityAdminChainID == nil {
		e.IdentityAdminChainID = primitives.NewZeroHash()
	}
	e.PrevDBSig.Init()
}

func (c *DBSignatureEntry) UpdateState(state interfaces.IState) error {
	return fmt.Errorf("Should not be called alone!")
	//return nil
}

// Create a new DB Signature Entry
func NewDBSignatureEntry(identityAdminChainID interfaces.IHash, sig interfaces.IFullSignature) (*DBSignatureEntry, error) {
	if identityAdminChainID == nil {
		return nil, fmt.Errorf("No identityAdminChainID provided")
	}
	if sig == nil {
		return nil, fmt.Errorf("No sig provided")
	}
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

func (e *DBSignatureEntry) MarshalBinary() ([]byte, error) {
	e.Init()
	var buf primitives.Buffer

	err := buf.PushByte(e.Type())
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(e.IdentityAdminChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&e.PrevDBSig)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *DBSignatureEntry) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)
	b, err := buf.PopByte()
	if b != e.Type() {
		return nil, fmt.Errorf("Invalid Entry type")
	}
	e.IdentityAdminChainID = new(primitives.Hash)
	err = buf.PopBinaryMarshallable(e.IdentityAdminChainID)
	if err != nil {
		return nil, err
	}
	err = buf.PopBinaryMarshallable(&e.PrevDBSig)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
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
	e.Init()
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
