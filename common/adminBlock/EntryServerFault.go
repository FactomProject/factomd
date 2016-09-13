package adminBlock

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type ServerFault struct {
	Timestamp interfaces.Timestamp
	// The following 4 fields represent the "Core" of the message
	// This should match the Core of ServerFault messages
	ServerID      interfaces.IHash
	AuditServerID interfaces.IHash
	VMIndex       byte
	DBHeight      uint32
	Height        uint32

	SignatureList SigList
}

type SigList struct {
	Length uint32
	List   []interfaces.IFullSignature
}

var _ interfaces.IABEntry = (*ServerFault)(nil)
var _ interfaces.BinaryMarshallable = (*ServerFault)(nil)

func (sl *SigList) MarshalBinary() (data []byte, err error) {
	var buf primitives.Buffer

	binary.Write(&buf, binary.BigEndian, uint32(sl.Length))

	for _, individualSig := range sl.List {
		if d, err := individualSig.MarshalBinary(); err != nil {
			return nil, err
		} else {
			buf.Write(d)
		}
	}

	return buf.DeepCopyBytes(), nil
}

func (sl *SigList) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling SigList in Full Server Fault: %v", r)
		}
	}()
	newData = data
	sl.Length, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	for i := sl.Length; i > 0; i-- {
		tempSig := new(primitives.Signature)
		newData, err = tempSig.UnmarshalBinaryData(newData)
		if err != nil {
			return nil, err
		}
		sl.List = append(sl.List, tempSig)
	}
	return newData, nil
}

func (c *ServerFault) UpdateState(state interfaces.IState) {
	//TODO: do
	/*
		state.AddFedServer(c.DBHeight, c.IdentityChainID)
		state.UpdateAuthorityFromABEntry(c)
	*/
}
func (m *ServerFault) MarshalCore() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error marshalling Server Fault Core: %v", r)
		}
	}()

	var buf primitives.Buffer

	if d, err := m.ServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}
	if d, err := m.AuditServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	buf.WriteByte(m.VMIndex)
	binary.Write(&buf, binary.BigEndian, uint32(m.DBHeight))
	binary.Write(&buf, binary.BigEndian, uint32(m.Height))

	return buf.DeepCopyBytes(), nil
}

func (m *ServerFault) MarshalBinary() (data []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error marshalling Invalid Server Fault: %v", r)
		}
	}()

	var buf primitives.Buffer

	if d, err := m.Timestamp.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	if d, err := m.ServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}
	if d, err := m.AuditServerID.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	buf.WriteByte(m.VMIndex)
	binary.Write(&buf, binary.BigEndian, uint32(m.DBHeight))
	binary.Write(&buf, binary.BigEndian, uint32(m.Height))

	if d, err := m.SignatureList.MarshalBinary(); err != nil {
		return nil, err
	} else {
		buf.Write(d)
	}

	return buf.DeepCopyBytes(), nil
}

func (m *ServerFault) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling With Signatures Invalid Server Fault: %v", r)
		}
	}()
	newData = data

	m.Timestamp = new(primitives.Timestamp)
	newData, err = m.Timestamp.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	if m.ServerID == nil {
		m.ServerID = primitives.NewZeroHash()
	}
	newData, err = m.ServerID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	if m.AuditServerID == nil {
		m.AuditServerID = primitives.NewZeroHash()
	}
	newData, err = m.AuditServerID.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	m.VMIndex, newData = newData[0], newData[1:]
	m.DBHeight, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]
	m.Height, newData = binary.BigEndian.Uint32(newData[0:4]), newData[4:]

	newData, err = m.SignatureList.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	return newData, nil
}

func (m *ServerFault) UnmarshalBinary(data []byte) error {
	_, err := m.UnmarshalBinaryData(data)
	return err
}

func (e *ServerFault) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *ServerFault) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *ServerFault) JSONBuffer(b *bytes.Buffer) error {
	return primitives.EncodeJSONToBuffer(e, b)
}

func (e *ServerFault) IsInterpretable() bool {
	return false
}

func (e *ServerFault) Interpret() string {
	return ""
}

func (e *ServerFault) Hash() interfaces.IHash {
	bin, err := e.MarshalBinary()
	if err != nil {
		panic(err)
	}
	return primitives.Sha(bin)
}

func (e *ServerFault) String() string {
	str, _ := e.JSONString()
	return str
	/*
		var out primitives.Buffer
		out.WriteString(fmt.Sprintf("    E: %35s -- %17s %8x %12s %8d\n", "AddFedServer", "IdentityChainID", e.IdentityChainID.Bytes()[:4], "DBHeight", e.DBHeight))
		return (string)(out.DeepCopyBytes())
	*/
}

func (e *ServerFault) Type() byte {
	return constants.TYPE_SERVER_FAULT
}
