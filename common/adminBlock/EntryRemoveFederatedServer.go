package adminBlock

import (
	"bytes"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// DB Signature Entry -------------------------
type RemoveFederatedServer struct {
	IdentityChainID interfaces.IHash
    DBHeight uint32
}

var _ interfaces.IABEntry = (*RemoveFederatedServer)(nil)
var _ interfaces.BinaryMarshallable = (*RemoveFederatedServer)(nil)

func (c *RemoveFederatedServer) UpdateState(state interfaces.IState) {
	if len(state.GetFedServers(c.DBHeight)) == 0 {
		state.AddFedServer(c.DBHeight, c.IdentityChainID)
	}
	state.Println(fmt.Sprintf("Removed Federated Server: %x", c.IdentityChainID.Bytes()[:3]))
}

// Create a new DB Signature Entry
func NewRemoveFederatedServer(dbheight uint32, identityChainID interfaces.IHash) (e *RemoveFederatedServer) {
	e = new(RemoveFederatedServer)
	e.IdentityChainID = primitives.NewHash(identityChainID.Bytes())
    e.DBHeight = dbheight
	return
}

func (e *RemoveFederatedServer) Type() byte {
	return constants.TYPE_REMOVE_FED_SERVER
}

func (e *RemoveFederatedServer) MarshalBinary() (data []byte, err error) {
	var buf bytes.Buffer

	data, err = e.IdentityChainID.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.Bytes(), nil
}

func (e *RemoveFederatedServer) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	newData = data
	newData = newData[1:]

	e.IdentityChainID = new(primitives.Hash)
	newData, err = e.IdentityChainID.UnmarshalBinaryData(newData)
	if err != nil {
		return
	}
	return
}

func (e *RemoveFederatedServer) UnmarshalBinary(data []byte) (err error) {
	_, err = e.UnmarshalBinaryData(data)
	return
}

func (e *RemoveFederatedServer) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *RemoveFederatedServer) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *RemoveFederatedServer) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *RemoveFederatedServer) String() string {
	str := fmt.Sprintf("Add Server with Identity Chain ID = %x", e.IdentityChainID.Bytes()[:5])
	return str
}

func (e *RemoveFederatedServer) IsInterpretable() bool {
	return false
}

func (e *RemoveFederatedServer) Interpret() string {
	return ""
}

func (e *RemoveFederatedServer) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}
