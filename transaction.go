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
	MarshalBinarySig() ([]byte, error)
	
	AddInput(input IAddress, amount uint64, )
    AddOutput(output IAddress, amount uint64 )
    AddECOutput(ecoutput IAddress, amount uint64 )
    AddRCD(rcd IRCD)
    GetInput(i int) IInAddress
	GetOutput(i int) IOutAddress
	GetOutEC(i int) IOutECAddress
	GetRCD(i int) IRCD
	GetSignatureBlock(i int) ISignatureBlock
	SetSignatureBlock(i int, signatureblk ISignatureBlock)
	GetInputs() []IInAddress
	GetOutputs() []IOutAddress
	GetOutECs() []IOutECAddress
	GetRCDs() []IRCD
	GetSignatureBlocks() []ISignatureBlock
	TotalInputs() uint64
	TotalOutputs() uint64
	TotalECs() uint64
    Validate() bool
    
	CalculateFee(factoshisPerEC uint64) (uint64,error)
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
	sigBlocks      []ISignatureBlock
}

var _ ITransaction = (*Transaction)(nil)

func (t *Transaction)SetSignatureBlock(i int, sig ISignatureBlock)  {
    for len(t.sigBlocks) <= i {
        t.sigBlocks = append(t.sigBlocks,new(SignatureBlock))
    }
    t.sigBlocks[i] = sig
}

func (t *Transaction)GetSignatureBlock(i int) ISignatureBlock {
    for len(t.sigBlocks) <= i {
        t.sigBlocks = append(t.sigBlocks,new(SignatureBlock))
    }
    return t.sigBlocks[i]
}

func (t *Transaction)AddRCD(rcd IRCD) {
    t.rcds = append(t.rcds, rcd)
}
    
func (w1 Transaction)GetDBHash() IHash {
    return Sha([]byte("Transaction"))
}

func (w1 Transaction)GetNewInstance() IBlock {
    return new(Transaction)
}

// Fee structure can be found:
// https://github.com/FactomProject/FactomDocs/blob/master/factomDataStructureDetails.md#sighash-type
//
//Transaction data size. -- Factoid transactions are charged the same 
//    amount as Entry Credits (EC). The size fees are 1 EC per KiB with a 
//    maximum transaction size of 10 KiB.
//Number of outputs created -- These are data points which potentially 
//    need to be tracked far into the future. They are more expensive 
//    to handle, and require a larger sacrifice. Outputs cost 10 EC per 
//    output. A purchase of Entry Credits also requires the 10 EC sized 
//    fee to be valid.
//Number of signatures checked -- These cause expensive computation on 
//    all full nodes. A fee of 1 EC equivalent must be paid for each 
//    signature included.
func (t Transaction)CalculateFee(factoshisPerEC uint64) (uint64, error) {
    
    // First look at the size of the transaction, and make sure
    // everything is inbounds.
    data,err := t.MarshalBinary()
    if err != nil {                  
        return 0, fmt.Errorf("Can't Marshal the Transaction")   
    }
    if len(data) > MAX_TRANSACTION_SIZE { // Can't be bigger than our limits
        return 0, fmt.Errorf("Transaction is greater than the max transaction size")
    }
    // Okay, we know the transaction is mostly good. Let's calculate 
    // fees.
    var fee uint64
    
    fee = factoshisPerEC * uint64((len(data)+1023)/1024)
    
    fee += factoshisPerEC * 10 * uint64( len(t.outputs) + len(t.outECs) )
    
    for _,rcd := range t.rcds {
        fee += factoshisPerEC * uint64(rcd.NumberOfSignatures())
    }
    
    return fee, nil
}

func (t Transaction)TotalInputs() uint64 {
    var sum uint64
    for _,input := range t.inputs {
        sum += input.GetAmount()
    }
    return sum
}

func (t Transaction)TotalOutputs() uint64 {
    var sum uint64
    for _,output := range t.outputs {
        sum += output.GetAmount()
    } 
    return sum
}

func (t Transaction)TotalECs() uint64 {
    var sum uint64
    for _,ec := range t.outECs {
        sum += ec.GetAmount()
    } 
    return sum
}

// Only validates that the transaction is well formed.  This means that 
// the inputs cover the value of the outputs.  Can't validate addresses,
// as they are hashes.  Can't validate the fee, because it might change 
// in the next period.
//
// If this validation returns false, the transaction can safely be 
// discarded.
//
// Note that the coinbase transaction for any block is never technically
// valid.  That validation must be done separately.
//
// Also note that we DO allow for transactions that do not have any outputs.
// This provides for a provable "burn" of factoids, since all inputs would
// go as "transaction fees" and those fees do not go to anyone.
func (t Transaction)Validate() bool {
         
    // Inputs must cover outputs
    if t.TotalInputs() < t.TotalOutputs()+t.TotalECs() {return false} 
    // Cannot have zero inputs.  This means you cannot use this function
    // to validate coinbase transactions, because they cannot have any
    // inputs.
    if   len(t.inputs) == 0 {return false}
    // Because of our fee structure, we may not enforce a minimum spend.
    // However, we do check the constant anyway.
    if t.TotalInputs() < MINIMUM_AMOUNT {return false}
    // Every input must have an RCD block
    if len(t.inputs) != len(t.rcds) { return false }
    // Every input must match the address of an RCD (which is the hash
    // of the RCD
    for i,rcd := range t.rcds {
        // Get the address specified by the RCD.
        address, err := rcd.GetAddress()
        // If there is anything wrong with the RCD, then the transaction isn't
        // valid.
        if err != nil {return false}
        // If the Address (which is really a hash) isn't equal to the hash of
        // the RCD, this transaction is bogus.
        if !t.inputs[i].GetAddress().IsEqual(address) {return false}
    }
    return true
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
    for i, s := range t1.sigBlocks {
        if !s.IsEqual(t2.GetSignatureBlock(i)) {
            return false
        }
    }
    
	return true
}

func (t Transaction) GetInputs()          []IInAddress      { return t.inputs }
func (t Transaction) GetOutputs()         []IOutAddress     { return t.outputs }
func (t Transaction) GetOutECs()          []IOutECAddress   { return t.outECs }
func (t Transaction) GetRCDs()            []IRCD            { return t.rcds }

func (t *Transaction) GetSignatureBlocks() []ISignatureBlock { 
    if len(t.sigBlocks) > len(t.inputs) {                    // If too long, nil out
        for i := len(t.inputs); i < len(t.sigBlocks); i++ {  // the extra entries, and
            t.sigBlocks[i] = nil                            // cut it to length.
        }
        t.sigBlocks = t.sigBlocks[:len(t.inputs)]
        return t.sigBlocks
    }
    for i := len(t.sigBlocks); i <len(t.inputs); i++ {          // If too short, then
        t.sigBlocks = append(t.sigBlocks, new(SignatureBlock)) // pad it with 
    }                                                           // signature blocks.
    return t.sigBlocks 
}

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
    
//     {   // limit the scope of d
//         var d [8]byte
//         copy(d[3:],data[0:5])
//         t.lockTime, data = binary .BigEndian.Uint64(d[:]), data[5:]
//     }
    t.lockTime, data = binary .BigEndian.Uint64(data[:]), data[8:]
    
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
    if t.sigBlocks == nil {
        t.sigBlocks = make([]ISignatureBlock, len(t.inputs))
    }
    for i := 0; i < len(t.inputs); i++ {
        t.rcds[i] = CreateRCD(data)
        data, err = t.rcds[i].UnmarshalBinaryData(data)
        if err != nil {
            return nil, err
        }
    
        t.sigBlocks[i] = new(SignatureBlock)
        data, err = t.sigBlocks[i].UnmarshalBinaryData(data)
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

// This is what Gets Signed.  Yet signature blocks are part of the transaction.
// We don't include them here, and tack them on later.
func (t *Transaction) MarshalBinarySig() ([]byte, error) {
	var out bytes.Buffer
    
// 	{  // limit the scope of tmp
//        var tmp bytes.Buffer
//        binary.Write(&tmp, binary.BigEndian, uint64(t.lockTime))
// 	   out.Write(tmp.Bytes()[3:])
//     }
    binary.Write(&out, binary.BigEndian, uint64(t.lockTime)) 
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

// This just Marshals what gets signed, i.e. MarshalBinarySig(), then
// Marshals the signatures and the RCDs for this transaction.
func (t Transaction) MarshalBinary() ([]byte, error) {
    var out bytes.Buffer
    
    data, err := t.MarshalBinarySig()
    if err != nil {
        return nil, err
    }
    out.Write(data)
    
    for i, rcd := range t.rcds {
        
        // Write the RCD
        data, err := rcd.MarshalBinary()
        if err != nil {
            return nil, err
        }
        out.Write(data)
        
        // Then write its signature blocks.  This needs to be
        // reworked so we use the information from the RCD block
        // to control the writing of the signatures.  After all,
        // we don't want to restrict what might be required to
        // sign an input.
        if len(t.sigBlocks) <= i {
            t.sigBlocks = append(t.sigBlocks, new(SignatureBlock))
        }
        data, err = t.sigBlocks[i].MarshalBinary()
        if err != nil {
            return nil, err
        }
        out.Write(data)
    }
    
    for i := 0; i < len(t.inputs); i++ {
    }
    
    return out.Bytes(), nil
}


// Helper function for building transactions.  Add an input to
// the transaction.  I'm guessing 5 inputs is about all anyone
// will need, so I'll default to 5.  Of course, go will grow
// past that if needed.
func (t *Transaction) AddInput( input IAddress,amount uint64) {
	if t.inputs == nil {
		t.inputs = make([]IInAddress, 0, 5)
	}
	out := NewInAddress(input, amount)
	t.inputs = append(t.inputs, out)
}

// Helper function for building transactions.  Add an output to
// the transaction.  I'm guessing 5 outputs is about all anyone
// will need, so I'll default to 5.  Of course, go will grow
// past that if needed.
func (t *Transaction) AddOutput(output IAddress,amount uint64) {
	if t.outputs == nil {
		t.outputs = make([]IOutAddress, 0, 5)
	}
	out := NewOutAddress(output, amount)
	t.outputs = append(t.outputs, out)

}

// Add a EntryCredit output.  Validating this is going to require
// access to the exchange rate.  This is literally how many entry
// credits are being added to the specified Entry Credit address.
func (t *Transaction) AddECOutput( ecoutput IAddress, amount uint64) {
	if t.outECs == nil {
		t.outECs = make([]IOutECAddress, 0, 5)
	}
	out := NewOutECAddress(ecoutput, amount)
	t.outECs = append(t.outECs, out)

}

// Marshal to text.  Largely a debugging thing.
func (t Transaction) MarshalText() (text []byte, err error) {
	var out bytes.Buffer

	out.WriteString("Transaction:\n locktime")
    WriteNumber64(&out, uint64(t.lockTime))
	out.WriteString("\n in:  ")
	WriteNumber16(&out, uint16(len(t.inputs)))
	out.WriteString("\n out: ")
	WriteNumber16(&out, uint16(len(t.outputs)))
	out.WriteString("\n ec:  ")
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
	for i, rcd := range t.rcds {
        text, err = rcd.MarshalText()
        if err != nil {
            return nil, err
        }
        out.Write(text)
    
        if len(t.sigBlocks) < i {
            t.sigBlocks = append(t.sigBlocks, new(SignatureBlock))
        }
        text, err := t.sigBlocks[i].MarshalText()
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
