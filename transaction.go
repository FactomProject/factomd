// Copyright (c) 2013-2015 Conformal Systems LLC.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package simplecoin

import (
	// "fmt"
	"bytes"
	"encoding/binary"
)

type ITransaction interface {
	IBlock
	AddInput(amount uint64, input IAddress)
	AddOutput(amount uint64, output IAddress)
	AddECOutput(amount uint64, ecoutput IAddress)
}

type Transaction struct {
	ITransaction
	// Binary Format has these additional fields
	// uint16 number of inputs
	// uint16 number of outputs
	// uint16 number of outECs (Number of EntryCredits)
	inputs  []IInAddress
	outputs []IOutAddress
	outECs  []IOutECAddress
}

func (t Transaction) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	binary.Write(&out, binary.BigEndian, uint16(len(t.inputs)))
	binary.Write(&out, binary.BigEndian, uint16(len(t.outputs)))
	binary.Write(&out, binary.BigEndian, uint16(len(t.outECs)))

	for _, input := range t.inputs {
		data, err := input.MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}

	for _, output := range t.outputs {
		data, err := output.MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
    }

	for _, outEC := range t.outECs {
		data, err := outEC.MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
    }
	
	return out.Bytes(), nil
}

var _ ITransaction = (*Transaction)(nil)

func (cb Transaction) NewBlock() IBlock {
	blk := new(Transaction)
	return blk
}

func (t *Transaction) AddInput(amount uint64, input IAddress) {
	if t.inputs == nil {
		t.inputs = make([]IInAddress, 0, 5)
	}
	out := NewInAddress(amount, input)
	t.inputs = append(t.inputs, out)
}

func (t *Transaction) AddOutput(amount uint64, output IAddress) {
	if t.outputs == nil {
		t.outputs = make([]IOutAddress, 0, 5)
	}
	out := NewOutAddress(amount, output)
	t.outputs = append(t.outputs, out)

}

func (t *Transaction) AddECOutput(amount uint64, ecoutput IAddress) {
	if t.outECs == nil {
		t.outECs = make([]IOutECAddress, 0, 5)
	}
	out := NewOutECAddress(amount, ecoutput)
	t.outECs = append(t.outECs, out)

}

func (t Transaction) MarshalText2() (text []byte, err error) {
	var out bytes.Buffer

	out.WriteString("in  ")
	WriteNumber16(&out, uint16(len(t.inputs)))
	out.WriteString("\nout ")
	WriteNumber16(&out, uint16(len(t.outputs)))
	out.WriteString("\nec  ")
	WriteNumber16(&out, uint16(len(t.outECs)))
	out.WriteString("\n")

	for _, address := range t.inputs {
		text, _ := address.MarshalText()
		out.Write(text)
	}
	for _, address := range t.outputs {
		text, _ := address.MarshalText()
		out.Write(text)
	}
	for _, ecaddress := range t.outECs {
		text, _ := ecaddress.MarshalText()
		out.Write(text)
	}

	return out.Bytes(), nil
}

func (t Transaction) MarshalText() (text []byte, err error) {
    return t.MarshalText2()
}

/************************
 * SignedTransaction
 ************************/

type ISignedTransaction interface {
	ITransaction
	AddAuthorization(auth IAuthorization)
}

type SignedTransaction struct {
    Transaction
	authorizations []IAuthorization
}

func (cb SignedTransaction) NewBlock() IBlock {
    blk := new(SignedTransaction)
    return blk
}

func (st *SignedTransaction) AddAuthorization(auth IAuthorization) {
	if st.authorizations == nil {
		st.authorizations = make([]IAuthorization, 0, 5)
	}
	st.authorizations = append(st.authorizations, auth)
}

func (st SignedTransaction) MarshalText() ( []byte,  error) {
    var out bytes.Buffer
    
    text, err := st.MarshalText2()
    out.Write(text)
    if err != nil {
        return nil,err
    }
    
    for _, authorization := range st.authorizations {
        text, err = authorization.MarshalText()
        out.Write(text)
        if err != nil {
            return nil,err
        }
    }
    
    return out.Bytes(), nil
}