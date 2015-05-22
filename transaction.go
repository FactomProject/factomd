// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package simplecoin

import (
    "fmt"
	"bytes"
	"encoding/binary"
)

type ITransaction interface {
	IBlock
	AddInput(amount uint64, input IAddress)
	AddOutput(amount uint64, output IAddress)
	AddECOutput(amount uint64, ecoutput IAddress)
    GetInput(i int) IInAddress
    GetOutput(i int) IOutAddress
    GetOutEC(i int) IOutECAddress
    GetInputs() []IInAddress
    GetOutputs() []IOutAddress
    GetOutECs() []IOutECAddress
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

var _ ITransaction = (*Transaction)(nil)

// Tests if the transaction is equal in all of its structures, and 
// in order of the structures.  Largely used to test and debug, but
// generally useful.
func (t1 Transaction) IsEqual2(trans IBlock) bool {
    
    t2, ok := trans.(ITransaction)
    
    if 
        !ok ||                                              // Not the right kind of IBlock
        len(t1.inputs) != len(t2.GetInputs()) ||             // Size of arrays has to match
        len(t1.outputs) != len(t2.GetOutputs()) ||           // Size of arrays has to match
        len(t1.outECs) != len(t2.GetOutECs()) {              // Size of arrays has to match
        
        return false
    }
    
    for i,input := range t1.GetInputs() {
        if !input.IsEqual(t2.GetInput(i)){
            return false
        }
    }
    for i,output := range t1.GetOutputs() {
        if !output.IsEqual(t2.GetOutput(i)){
            return false
        }
    }
    for i,outEC := range t1.GetOutECs() {
        if !outEC.IsEqual(t2.GetOutEC(i)){
            return false
        }
    }
    
    return true
}

func (t Transaction) GetInputs() []IInAddress { return t.inputs }
func (t Transaction) GetOutputs() []IOutAddress { return t.outputs }
func (t Transaction) GetOutECs() []IOutECAddress { return t.outECs }

func (t1 Transaction) IsEqual(trans IBlock) bool {
    return t1.IsEqual2(trans)
}

func (t *Transaction) GetInput(i int) IInAddress {
    if i > len(t.inputs) {
        return nil
    }
    return t.inputs[i]
}

func (t *Transaction) GetOutput(i int) IOutAddress{
    if i > len(t.outputs) {
        return nil
    }
    return t.outputs[i]
}

func (t *Transaction) GetOutEC(i int) IOutECAddress{
    if i > len(t.outECs) {
        return nil
    }
    return t.outECs[i]
}


// UnmarshalBinary assumes that the Binary is all good.  We do error
// out if there isn't enough data, or the transaction is too large.
func (t *Transaction) UnmarshalBinaryData2(data []byte) (newData []byte, err error) {
   
    if len(data)<72 {
        return nil, fmt.Errorf("Transaction data too small: %d bytes",len(data))
    }
    if len(data)>MAX_TRANSACTION_SIZE {
        return nil, fmt.Errorf("Transaction data too large: %d bytes",len(data))
    }
    
    numInputs, data := binary.BigEndian.Uint16(data[0:2]), data[2:]
    numOutputs, data := binary.BigEndian.Uint16(data[0:2]), data[2:]
    numOutECs, data := binary.BigEndian.Uint16(data[0:2]), data[2:]
    
    t.inputs = make([]IInAddress,numInputs,numInputs)
    t.outputs = make([]IOutAddress,numOutputs,numOutputs)
    t.outECs = make([]IOutECAddress,numOutECs,numOutECs)
    
    for i,_ := range t.inputs {
        t.inputs[i]= new (InAddress)
        data, err = t.inputs[i].UnmarshalBinaryData(data)
        if(err != nil || t.inputs[i] == nil) { return nil, err }
    }
    for i,_ := range t.outputs {
        t.outputs[i]= new (OutAddress)
        data, err = t.outputs[i].UnmarshalBinaryData(data)
        if(err != nil) { return nil, err }
    }
    for i,_ := range t.outECs {
        t.outECs[i]= new (OutECAddress)
        data, err = t.outECs[i].UnmarshalBinaryData(data)
        if(err != nil) { return nil, err }
        
    }
    
    return data, nil
}

func (t *Transaction) UnmarshalBinary(data []byte) (err error) {
    data, err = t.UnmarshalBinaryData2(data)
    return err
}


// This is a workaround helper function so that Signed Transactions don't
// have to duplicate this code.
//
func (t Transaction) MarshalBinary2() ([]byte, error) {
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

// We can't call our "parent's" implementation of a function, (or I 
// don't know how to in go) so I have a helper function that leaves
// my parent's implementation callable.
func (t Transaction) MarshalBinary() ([]byte, error) {
    return t.MarshalBinary2()
}

func (cb Transaction) NewBlock() IBlock {
	blk := new(Transaction)
	return blk
}

// Helper function for building transactions.  Add an input to 
// the transaction.  I'm guessing 5 inputs is about all anyone
// will need, so I'll default to 5.  Of course, go will grow
// past that if needed.
func (t *Transaction) AddInput(amount uint64, input IAddress) {
	if t.inputs == nil {
		t.inputs = make([]IInAddress, 0, 5)
	}
	out := NewInAddress(amount, input)
	t.inputs = append(t.inputs, out)
}

// Helper function for building transactions.  Add an output to
// the transaction.  I'm guessing 5 outputs is about all anyone
// will need, so I'll default to 5.  Of course, go will grow
// past that if needed.
func (t *Transaction) AddOutput(amount uint64, output IAddress) {
	if t.outputs == nil {
		t.outputs = make([]IOutAddress, 0, 5)
	}
	out := NewOutAddress(amount, output)
	t.outputs = append(t.outputs, out)

}

// Add a EntryCredit output.  Validating this is going to require
// access to the exchange rate.  This is literally how many entry
// credits are being added to the specified Entry Credit address.
func (t *Transaction) AddECOutput(amount uint64, ecoutput IAddress) {
	if t.outECs == nil {
		t.outECs = make([]IOutECAddress, 0, 5)
	}
	out := NewOutECAddress(amount, ecoutput)
	t.outECs = append(t.outECs, out)

}

// Helper function because I don't know how to call a parent struct's
// implementation in Go.  Of course, instead of inheritence, we could
// just keep a pointer to the transaction in SignedTransaction.  That
// likely makes more sense, but we can change that later easy enough.
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

// MarshalText.  Eough said. Mostly for debugging.
func (t Transaction) MarshalText() (text []byte, err error) {
	return t.MarshalText2()
}

/************************
 * SignedTransaction
 ************************/

// These transactions are exactly Transactions plus authentication 
// blocks.

type ISignedTransaction interface {
	ITransaction
	AddAuthorization(auth IRCD)
}

type SignedTransaction struct {
	Transaction
	authorizations []IRCD
}

var _ ISignedTransaction = (*SignedTransaction)(nil)

// Tests if the transaction is equal in all of its structures, and 
// in order of the structures.  Largely used to test and debug, but
// generally useful.
func (t1 SignedTransaction) IsEqual(trans IBlock) bool {
    t2, ok := trans.(*SignedTransaction)
    if  !ok || !t1.IsEqual2(t2) {              // It is the right type, and 
        return false                           //  the regular transaction stuff has to match 
    }
    
    for i,a := range t1.authorizations {
        if !a.IsEqual(t2.authorizations[i]){
            return false
        }
    }
    
    return true
}


func (t *SignedTransaction) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
    data, err = t.UnmarshalBinaryData2(data)
    
    if t.authorizations == nil {
        t.authorizations = make([]IRCD, len(t.inputs))
    }
    
    for i:=0; i<len(t.inputs); i++ {
        t.authorizations[i], data, err = UnmarshalBinaryAuth(data) 
        if err != nil {
            return nil, err
        }
    }
    
    return data,err
}

func (t *SignedTransaction) UnmarshalBinary(data []byte) (err error) {
    data, err = t.UnmarshalBinaryData(data)
    return err
}

// MarshalBinary.  Enough said.
func (st SignedTransaction) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer

	data, err := st.MarshalBinary2()
	if err != nil {
		return nil, err
	}
	out.Write(data)
   
	for _, authorization := range st.authorizations {
        data, err = authorization.MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}
	
	return out.Bytes(), nil
}

// Create one of me!
func (cb SignedTransaction) NewBlock() IBlock {
	blk := new(SignedTransaction)
	return blk
}

// Helper Function.  This simply adds an Authorization to a 
// transaction.  DOES NO VALIDATION.  Not the job of construction.
// That's why we have a validation call.
func (st *SignedTransaction) AddAuthorization(auth IRCD) {
	if st.authorizations == nil {
		st.authorizations = make([]IRCD, 0, 5)
	}
	st.authorizations = append(st.authorizations, auth)
}

// Marshal Text.  Mostly a debugging and testing thing.
func (st SignedTransaction) MarshalText() ([]byte, error) {
	var out bytes.Buffer

	text, err := st.MarshalText2()
	if err != nil {
		return nil, err
	}
	out.Write(text)

	for _, authorization := range st.authorizations {
		text, err = authorization.MarshalText()
		if err != nil {
			return nil, err
		}
		out.Write(text)
	}

	return out.Bytes(), nil
}
