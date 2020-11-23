package securedb

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/primitives"
	llog "github.com/FactomProject/factomd/log"
)

type SecureDBMetaData struct {
	Salt      primitives.ByteSlice
	Challenge primitives.ByteSlice
}

func NewSecureDBMetaData() *SecureDBMetaData {
	s := new(SecureDBMetaData)
	return s
}

func (m *SecureDBMetaData) IsSameAs(b *SecureDBMetaData) bool {
	if !m.Salt.IsSameAs(&b.Salt) {
		return false
	}

	if !m.Challenge.IsSameAs(&b.Challenge) {
		return false
	}

	return true
}

func (m *SecureDBMetaData) UnmarshalBinary(data []byte) (err error) {
	_, err = m.UnmarshalBinaryData(data)
	return
}

func (m *SecureDBMetaData) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
			llog.LogPrintf("recovery", "Error unmarshalling: %v", r)
		}
	}()

	newData = data

	slen, err := bytesToUint32(newData[:4])
	if err != nil {
		return nil, err
	}
	m.Salt.Bytes = make([]byte, slen)
	copy(m.Salt.Bytes, newData[4:slen+4])
	newData = newData[slen+4:]

	clen, err := bytesToUint32(newData[:4])
	if err != nil {
		return nil, err
	}
	m.Challenge.Bytes = make([]byte, clen)
	copy(m.Challenge.Bytes, newData[4:clen+4])
	newData = newData[clen+4:]

	return
}

func (m *SecureDBMetaData) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "SecureDBMetaData.MarshalBinary err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)

	buf.Write(intToBytes(len(m.Salt.Bytes)))
	data, err := m.Salt.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	buf.Write(intToBytes(len(m.Challenge.Bytes)))
	data, err = m.Challenge.MarshalBinary()
	if err != nil {
		return nil, err
	}
	buf.Write(data)

	return buf.DeepCopyBytes(), nil
}

func bytesToUint32(data []byte) (ret uint32, err error) {
	buf := bytes.NewBuffer(data)
	err = binary.Read(buf, binary.BigEndian, &ret)
	return
}

func intToBytes(val int) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, uint32(val))

	return b
}
