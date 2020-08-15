package adminBlock

import (
	"fmt"
	"os"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

// DB Signature Entry -------------------------
type AddFederatedServerSigningKey struct {
	AdminIDType     uint32               `json:"adminidtype"`
	IdentityChainID interfaces.IHash     `json:"identitychainid"`
	KeyPriority     byte                 `json:"keypriority"`
	PublicKey       primitives.PublicKey `json:"publickey"`
	DBHeight        uint32               `json:"dbheight"`
}

var _ interfaces.IABEntry = (*AddFederatedServerSigningKey)(nil)
var _ interfaces.BinaryMarshallable = (*AddFederatedServerSigningKey)(nil)

func (e *AddFederatedServerSigningKey) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

func (c *AddFederatedServerSigningKey) UpdateState(state interfaces.IState) error {
	c.Init()
	state.UpdateAuthorityFromABEntry(c)
	return nil
}

func (e *AddFederatedServerSigningKey) SortedIdentity() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AddFederatedServerSigningKey.SortedIdentity") }()

	return e.IdentityChainID
}

func (e *AddFederatedServerSigningKey) String() string {
	e.Init()
	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("    E: %35s -- %17s %8x %12s %8x %12s %8s %12s %d",
		"AddFederatedServerSigningKey",
		"IdentityChainID", e.IdentityChainID.Bytes()[3:6],
		"KeyPriority", e.KeyPriority,
		"PublicKey", e.PublicKey.String()[:8],
		"DBHeight", e.DBHeight))
	return (string)(out.DeepCopyBytes())
}

// Create a new DB Signature Entry
func NewAddFederatedServerSigningKey(identityChainID interfaces.IHash, keyPriority byte, publicKey primitives.PublicKey, height uint32) (e *AddFederatedServerSigningKey) {
	e = new(AddFederatedServerSigningKey)
	e.IdentityChainID = identityChainID
	e.KeyPriority = keyPriority
	e.PublicKey = publicKey
	e.DBHeight = height
	return
}

func (e *AddFederatedServerSigningKey) Type() byte {
	return constants.TYPE_ADD_FED_SERVER_KEY
}

func (e *AddFederatedServerSigningKey) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "AddFederatedServerSigningKey.MarshalBinary err:%v", *pe)
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
	err = buf.PushByte(e.KeyPriority)
	if err != nil {
		return nil, err
	}
	err = buf.PushBinaryMarshallable(&e.PublicKey)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt32(e.DBHeight)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *AddFederatedServerSigningKey) UnmarshalBinaryData(data []byte) ([]byte, error) {
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
	e.KeyPriority, err = buf.PopByte()
	err = buf.PopBinaryMarshallable(&e.PublicKey)
	if err != nil {
		return nil, err
	}
	e.DBHeight, err = buf.PopUInt32()
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (e *AddFederatedServerSigningKey) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *AddFederatedServerSigningKey) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *AddFederatedServerSigningKey) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *AddFederatedServerSigningKey) IsInterpretable() bool {
	return false
}

func (e *AddFederatedServerSigningKey) Interpret() string {
	return ""
}

func (e *AddFederatedServerSigningKey) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AddFederatedServerSigningKey.Hash") }()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
