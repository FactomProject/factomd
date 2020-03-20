package adminBlock

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DB Signature Entry -------------------------
type RemoveFederatedServer struct {
	AdminIDType     uint32           `json:"adminidtype"`
	IdentityChainID interfaces.IHash `json:"identitychainid"`
	DBHeight        uint32           `json:"dbheight"`
}

var _ interfaces.IABEntry = (*RemoveFederatedServer)(nil)
var _ interfaces.BinaryMarshallable = (*RemoveFederatedServer)(nil)

func (e *RemoveFederatedServer) Init() {
	if e.IdentityChainID == nil {
		e.IdentityChainID = primitives.NewZeroHash()
	}
	e.AdminIDType = uint32(e.Type())
}

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

func (c *RemoveFederatedServer) UpdateState(state interfaces.IState) error {
	c.Init()
	if len(state.GetFedServers(c.DBHeight)) != 0 {
		state.RemoveFedServer(c.DBHeight, c.IdentityChainID)
	}
	if state.GetOut() {
		state.Println(fmt.Sprintf("Removed Federated Server: %x", c.IdentityChainID.Bytes()[:4]))
	}
	authorityDeltaString := fmt.Sprintf("AdminBlock (RemoveFedMsg DBHt: %d) \n v %s", c.DBHeight, c.IdentityChainID.String()[5:10])
	state.AddStatus(authorityDeltaString)
	state.AddAuthorityDelta(authorityDeltaString)
	state.UpdateAuthorityFromABEntry(c)
	return nil
}

// Create a new DB Signature Entry
func NewRemoveFederatedServer(identityChainID interfaces.IHash, dbheight uint32) (e *RemoveFederatedServer) {
	if identityChainID == nil {
		return nil
	}
	e = new(RemoveFederatedServer)
	e.IdentityChainID = primitives.NewHash(identityChainID.Bytes())
	e.DBHeight = dbheight
	return
}

func (e *RemoveFederatedServer) Type() byte {
	return constants.TYPE_REMOVE_FED_SERVER
}

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

func (e *RemoveFederatedServer) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *RemoveFederatedServer) JSONByte() ([]byte, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSON(e)
}

func (e *RemoveFederatedServer) JSONString() (string, error) {
	e.AdminIDType = uint32(e.Type())
	return primitives.EncodeJSONString(e)
}

func (e *RemoveFederatedServer) IsInterpretable() bool {
	return false
}

func (e *RemoveFederatedServer) Interpret() string {
	return ""
}

func (e *RemoveFederatedServer) Hash() (rval interfaces.IHash) {
	defer func() { rval = primitives.CheckNil(rval, "RemoveFederatedServer.Hash") }()

	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
