// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package simplecoin

import (
	"bytes"
	"encoding/binary"
	"fmt"
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
	AddRCD(rcd IRCD)
    GetRCD(i int) IRCD
	Validate(factoshisPerEC uint64) bool
}

type Transaction struct {
	ITransaction
	// Binary Format has these additional fields
	// uint16 number of inputs
	// uint16 number of outputs
	// uint16 number of outECs (Number of EntryCredits)
	lockTime       uint64
	inputs         []IInAddress
	outputs        []IOutAddress
	outECs         []IOutECAddress
	rcds           []IRCD
}

var _ ITransaction = (*Transaction)(nil)

// Only validates that the transaction is well formed.  This means that 
// the inputs cover the value of the outputs.  Can't validate addresses,
// as they are hashes.  
//
// Validates the transaction fee, given the exchange rate to EC.
//
func (t Transaction)Validate(factoshisPerEC uint64) bool {
    var inSum, outSum uint64
    
    for _,input := range t.inputs {
        inSum += input.GetAmount()
    }
     
    for _,output := range t.outputs {
        outSum += output.GetAmount()
    } 
        
    return inSum >= outSum
}

// Tests if the transaction is equal in all of its structures, and
// in order of the structures.  Largely used to test and debug, but
// generally useful.
func (t1 Transaction) IsEqual(trans IBlock) bool {

	t2, ok := trans.(ITransaction)

	if !ok || // Not the right kind of IBlock
		len(t1.inputs) != len(t2.GetInputs()) || // Size of arrays has to match
		len(t1.outputs) != len(t2.GetOutputs()) || // Size of arrays has to match
		len(t1.outECs) != len(t2.GetOutECs()) { // Size of arrays has to match

		return false
	}

	for i, input := range t1.GetInputs() {
		if !input.IsEqual(t2.GetInput(i)) {
			return false
		}
	}
	for i, output := range t1.GetOutputs() {
		if !output.IsEqual(t2.GetOutput(i)) {
			return false
		}
	}
	for i, outEC := range t1.GetOutECs() {
		if !outEC.IsEqual(t2.GetOutEC(i)) {
			return false
		}
	}
	for i, a := range t1.rcds {
        if !a.IsEqual(t2.GetRCD(i)) {
            return false
        }
    }
	return true
}

func (t Transaction) GetInputs() []IInAddress    { return t.inputs }
func (t Transaction) GetOutputs() []IOutAddress  { return t.outputs }
func (t Transaction) GetOutECs() []IOutECAddress { return t.outECs }
func (t Transaction) GetRCDs()   []IRCD          { return t.rcds }


func (t *Transaction) GetInput(i int) IInAddress {
	if i > len(t.inputs) { return nil }
	return t.inputs[i]
}

func (t *Transaction) GetOutput(i int) IOutAddress {
	if i > len(t.outputs) { return nil }
	return t.outputs[i]
}

func (t *Transaction) GetOutEC(i int) IOutECAddress {
	if i > len(t.outECs) { return nil }
	return t.outECs[i]
}

func (t *Transaction) GetRCD(i int) IRCD {
    if i > len(t.rcds) { return nil }
    return t.rcds[i]
}

// UnmarshalBinary assumes that the Binary is all good.  We do error
// out if there isn't enough data, or the transaction is too large.
func (t *Transaction) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	if len(data) < 72 {
		return nil, fmt.Errorf("Transaction data too small: %d bytes", len(data))
	}
	if len(data) > MAX_TRANSACTION_SIZE {
		return nil, fmt.Errorf("Transaction data too large: %d bytes", len(data))
	}
    
    {   // limit the scope of d
        var d [8]byte
        copy(d[3:],data[0:5])
        t.lockTime, data = binary .BigEndian.Uint64(d[:]), data[5:]
    }
	
	numInputs, data := binary.BigEndian.Uint16(data[0:2]), data[2:]
	numOutputs, data := binary.BigEndian.Uint16(data[0:2]), data[2:]
	numOutECs, data := binary.BigEndian.Uint16(data[0:2]), data[2:]

	t.inputs = make([]IInAddress, numInputs, numInputs)
	t.outputs = make([]IOutAddress, numOutputs, numOutputs)
	t.outECs = make([]IOutECAddress, numOutECs, numOutECs)

	for i, _ := range t.inputs {
		t.inputs[i] = new(InAddress)
		data, err = t.inputs[i].UnmarshalBinaryData(data)
		if err != nil || t.inputs[i] == nil {
			return nil, err
		}
	}
	for i, _ := range t.outputs {
		t.outputs[i] = new(OutAddress)
		data, err = t.outputs[i].UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}
	for i, _ := range t.outECs {
		t.outECs[i] = new(OutECAddress)
		data, err = t.outECs[i].UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}
	
	if t.rcds == nil {
        t.rcds = make([]IRCD, len(t.inputs))
    }
    for i := 0; i < len(t.inputs); i++ {
        t.rcds[i] = CreateRCD(data)
        data, err = t.rcds[i].UnmarshalBinaryData(data)
        if err != nil {
            return nil, err
        }
    }
	return data, nil
}

func (t *Transaction) UnmarshalBinary(data []byte) (err error) {
	data, err = t.UnmarshalBinaryData(data)
	return err
}

// This is a workaround helper function so that Signed Transactions don't
// have to duplicate this code.
//
func (t Transaction) MarshalBinary() ([]byte, error) {
	var out bytes.Buffer
    
	{  // limit the scope of tmp
       var tmp bytes.Buffer
       binary.Write(&tmp, binary.BigEndian, uint64(t.lockTime))
	   out.Write(tmp.Bytes()[3:])
    }
    
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

	for _, rcd := range t.rcds {
        data, err := rcd.MarshalBinary()
        if err != nil {
            return nil, err
        }
        out.Write(data)
    }
    
	return out.Bytes(), nil
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
	for _, rcd := range t.rcds {
        text, err = rcd.MarshalText()
        if err != nil {
            return nil, err
        }
        out.Write(text)
    }
	return out.Bytes(), nil
}

// Helper Function.  This simply adds an Authorization to a
// transaction.  DOES NO VALIDATION.  Not the job of construction.
// That's why we have a validation call.
func (t *Transaction) AddAuthorization(auth IRCD) {
    if t.rcds == nil {
        t.rcds = make([]IRCD, 0, 5)
    }
    t.rcds = append(t.rcds, auth)
}
