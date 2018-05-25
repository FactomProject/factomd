package securedb

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/interfaces"
)

type EncryptedMarshaler struct {
	EncryptionKey []byte

	Original interfaces.BinaryMarshallable
}

func NewEncryptedMarshaler(key []byte, o interfaces.BinaryMarshallable) *EncryptedMarshaler {
	e := new(EncryptedMarshaler)
	e.EncryptionKey = key
	e.Original = o

	return e
}

func (e *EncryptedMarshaler) New() interfaces.BinaryMarshallableAndCopyable {
	e2 := NewEncryptedMarshaler(e.EncryptionKey, nil)
	c, ok := e.Original.(interfaces.BinaryMarshallableAndCopyable)
	if !ok {
		return e2
	}
	e2.Original = c.New()

	return e2
}

func (e *EncryptedMarshaler) UnmarshalBinary(cipherData []byte) (err error) {
	_, err = e.UnmarshalBinaryData(cipherData)
	return
}

func (e *EncryptedMarshaler) UnmarshalBinaryData(cipherData []byte) (newData []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling: %v", r)
		}
	}()

	if e.Original == nil {
		return nil, fmt.Errorf("No object given")
	}

	l, err := bytesToUint32(cipherData[:4])
	if err != nil {
		return nil, err
	}

	newData = cipherData[l+4:]

	plainData, err := Decrypt(cipherData[4:l+4], e.EncryptionKey)
	if err != nil {
		return nil, err
	}

	_, err = e.Original.UnmarshalBinaryData(plainData)
	if err != nil {
		return nil, err
	}

	return newData, nil
}

func (e *EncryptedMarshaler) MarshalBinary() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "EncryptedMarshaler.MarshalBinary err:%v", *pe)
		}
	}(&err)
	if e.Original == nil {
		return nil, fmt.Errorf("No object given")
	}

	plainData, err := e.Original.MarshalBinary()
	if err != nil {
		return nil, err
	}

	cipherData, err := Encrypt(plainData, e.EncryptionKey)
	if err != nil {
		return nil, err
	}
	l := intToBytes(len(cipherData))

	return append(l, cipherData...), nil
}
