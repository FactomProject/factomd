package adminBlock

import (
	"fmt"
	"os"
	"reflect"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DBSignatureEntry contains information for a directory block Signature Entry -------------------------
type DBSignatureEntry struct {
	AdminIDType          uint32               `json:"adminidtype"`          //  the type of action in this admin block entry: uint32(TYPE_DB_SIGNATURE)
	IdentityAdminChainID interfaces.IHash     `json:"identityadminchainid"` // Server 32 byte identity chain id
	PrevDBSig            primitives.Signature `json:"prevdbsig"`            // Contains a 32 byte Ed25519 public key and a 64 byte signature of the previous directory blocks header
}

var _ interfaces.IABEntry = (*DBSignatureEntry)(nil)
var _ interfaces.BinaryMarshallable = (*DBSignatureEntry)(nil)

// Init initializes the IdentityAdminChainID to the zero hash if nil
func (e *DBSignatureEntry) Init() {
	if e.IdentityAdminChainID == nil {
		e.IdentityAdminChainID = primitives.NewZeroHash()
	}
	e.PrevDBSig.Init()
	e.AdminIDType = uint32(e.Type())
}

// UpdateState always returns an error, this function should not be called
func (e *DBSignatureEntry) UpdateState(state interfaces.IState) error {
	return fmt.Errorf("Should not be called alone!")
	//return nil
}

// NewDBSignatureEntry creates a new directory block signature entry
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

// Type always returns TYPE_DB_SIGNATURE
func (e *DBSignatureEntry) Type() byte {
	return constants.TYPE_DB_SIGNATURE
}

// MarshalBinary marshals the DBSignatureEntry
func (e *DBSignatureEntry) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "DBSignatureEntry.MarshalBinary err:%v", *pe)
		}
	}(&err)
	e.Init()
	var buf primitives.Buffer

	err = buf.PushByte(e.Type())
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

// UnmarshalBinaryData unmarshals the input data into this DBSignatureEntry
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

// UnmarshalBinary unmarshals the input data into this DBSignatureEntry
func (e *DBSignatureEntry) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

// JSONByte returns the json encoded byte array
func (e *DBSignatureEntry) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

// JSONString returns the sjon encoded string
func (e *DBSignatureEntry) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

// String returns the string of the DBSignatureEntry
func (e *DBSignatureEntry) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %20s -- %17s %8x %12s %8s %12s %8x",
		"DB Signature",
		"IdentityChainID", e.IdentityAdminChainID.Bytes()[3:6],
		"PubKey", e.PrevDBSig.Pub.String()[:8],
		"Signature", e.PrevDBSig.Sig.String()[:8]))
	return (string)(out.DeepCopyBytes())
}

// IsInterpretable always returns false
func (e *DBSignatureEntry) IsInterpretable() bool {
	return false
}

// Interpret always returns the empty string ""
func (e *DBSignatureEntry) Interpret() string {
	return ""
}

// Hash marshals the DBSignatureEntry and calculates its hashS
func (e *DBSignatureEntry) Hash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("DBSignatureEntry.Hash() saw an interface that was nil")
		}
	}()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
