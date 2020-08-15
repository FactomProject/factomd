package adminBlock

import (
	"fmt"
	"os"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/primitives"
)

// DB Signature Entry -------------------------
type AddFederatedServer struct {
	AdminIDType     uint32           `json:"adminidtype"`
	IdentityChainID interfaces.IHash `json:"identitychainid"`
	DBHeight        uint32           `json:"dbheight"`
}

var _ interfaces.IABEntry = (*AddFederatedServer)(nil)
var _ interfaces.BinaryMarshallable = (*AddFederatedServer)(nil)

func (e *AddFederatedServer) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

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

func (c *AddFederatedServer) UpdateState(state interfaces.IState) error {
	c.Init()
	if c.DBHeight == 1 {
		//use the bootstrap identity for the process list following the genesis block
		id := state.GetNetworkBootStrapIdentity()
		state.AddFedServer(c.DBHeight, id)
	} else {
		state.AddFedServer(c.DBHeight, c.IdentityChainID)
	}
	authorityDeltaString := fmt.Sprintf("AdminBlock (AddFedMsg DBHt: %d) \n ^ %s", c.DBHeight, c.IdentityChainID.String()[5:10])
	state.AddStatus(authorityDeltaString)
	state.AddAuthorityDelta(authorityDeltaString)
	state.UpdateAuthorityFromABEntry(c)
	return nil
}

// Create a new DB Signature Entry
func NewAddFederatedServer(identityChainID interfaces.IHash, dbheight uint32) (e *AddFederatedServer) {
	if identityChainID == nil {
		return nil
	}
	e = new(AddFederatedServer)
	e.DBHeight = dbheight
	e.IdentityChainID = primitives.NewHash(identityChainID.Bytes())
	return
}

func (e *AddFederatedServer) Type() byte {
	return constants.TYPE_ADD_FED_SERVER
}

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

func (e *AddFederatedServer) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *AddFederatedServer) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *AddFederatedServer) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *AddFederatedServer) IsInterpretable() bool {
	return false
}

func (e *AddFederatedServer) Interpret() string {
	return ""
}

func (e *AddFederatedServer) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "AddFederatedServer.Hash") }()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
