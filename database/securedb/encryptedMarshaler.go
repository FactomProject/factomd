package securedb

import (
	"fmt"

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
	if e.Original == nil {
		return nil, fmt.Errorf("No object given")
	}

	plainData, err := Decrypt(cipherData, e.EncryptionKey)
	if err != nil {
		return nil, err
	}

	newData = plainData
	newData, err = e.Original.UnmarshalBinaryData(newData)
	if err != nil {
		return nil, err
	}

	if len(newData) > 0 {
		cipherNewData, err := Encrypt(newData, e.EncryptionKey)
		if err != nil {
			return nil, err
		}
		return cipherNewData, nil
	}
	return newData, nil
}

func (e *EncryptedMarshaler) MarshalBinary() ([]byte, error) {
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

	return cipherData, nil
}
