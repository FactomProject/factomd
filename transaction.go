// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"time"
)

type ITransaction interface {
	IBlock
	// Marshals the parts of the transaction that are signed to
	// validate the transaction.  This includes the transaction header,
	// the locktime, the inputs, outputs, and outputs to EntryCredits.  It
	// does not include the signatures and RCDs.  The inputs are the hashes
	// of the RCDs, so they are included indirectly.  The signatures
	// sign this hash, so they are included indirectly.
	MarshalBinarySig() ([]byte, error)
	// Add an input to the transaction.  No validation.
	AddInput(input IAddress, amount uint64)
	// Add an output to the transaction.  No validation.
	AddOutput(output IAddress, amount uint64)
	// Add an Entry Credit output to the transaction.  Denominated in
	// Factoids, and interpreted by the exchange rate in the server at
	// the time the transaction is added to Factom.
	AddECOutput(ecoutput IAddress, amount uint64)
	// Add an RCD.  Must match the input in the same order.  Inputs and
	// RCDs are generally added at the same time.
	AddRCD(rcd IRCD)

	// Accessors the inputs, outputs, and Entry Credit outputs (ecoutputs)
	// to this transaction.
	GetInput(int) (IInAddress, error)
	GetOutput(int) (IOutAddress, error)
	GetECOutput(int) (IOutECAddress, error)
	GetRCD(int) (IRCD, error)
	GetInputs() []IInAddress
	GetOutputs() []IOutAddress
	GetECOutputs() []IOutECAddress
	GetRCDs() []IRCD

	GetVersion() uint64
	// Locktime serves as a nonce to make every transaction unique. Transactions
	// that are more than 24 hours old are not included nor propagated through
	// the network.
	GetMilliTimestamp() uint64
	SetMilliTimestamp(uint64)
	// Get a signature
	GetSignatureBlock(i int) ISignatureBlock
	SetSignatureBlock(i int, signatureblk ISignatureBlock)
	GetSignatureBlocks() []ISignatureBlock

	// Helper functions for validation.
	TotalInputs() (uint64, error)
	TotalOutputs() (uint64, error)
	TotalECs() (uint64, error)

	// Validate does everything but check the signatures.
	Validate(int) error
	ValidateSignatures() error

	// Calculate the fee for a transaction, given the specified exchange rate.
	CalculateFee(factoshisPerEC uint64) (uint64, error)
}

type Transaction struct {
	// version     uint64         Version of transaction. Hardcoded, naturally.
	MilliTimestamp uint64
	// #inputs     uint8          number of inputs
	// #outputs    uint8          number of outputs
	// #ecoutputs  uint8          number of outECs (Number of EntryCredits)
	Inputs    []IInAddress
	Outputs   []IOutAddress
	OutECs    []IOutECAddress
	RCDs      []IRCD
	SigBlocks []ISignatureBlock

	MarshalSig IHash // cache to avoid unnecessary marshal/unmarshals
}

var _ ITransaction = (*Transaction)(nil)
var _ Printable = (*Transaction)(nil)

// Clears caches if they are no long valid.
func (t *Transaction) clearCaches() {
	return
	t.MarshalSig = nil
}

func (Transaction) GetVersion() uint64 {
	return 2
}

func (t Transaction) GetHash() IHash {
	m, err := t.MarshalBinary()
	if err != nil {
		return nil
	}
	return Sha(m)
}

func (t Transaction) String() string {
	txt, err := t.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

// MilliTimestamp is in milliseconds
func (t *Transaction) GetMilliTimestamp() uint64 {
	return t.MilliTimestamp
}
func (t *Transaction) SetMilliTimestamp(ts uint64) {
	t.MilliTimestamp = ts
}

func (t *Transaction) SetSignatureBlock(i int, sig ISignatureBlock) {
	for len(t.SigBlocks) <= i {
		t.SigBlocks = append(t.SigBlocks, new(SignatureBlock))
	}
	t.SigBlocks[i] = sig
}

func (t *Transaction) GetSignatureBlock(i int) ISignatureBlock {
	for len(t.SigBlocks) <= i {
		t.SigBlocks = append(t.SigBlocks, new(SignatureBlock))
	}
	return t.SigBlocks[i]
}

func (t *Transaction) AddRCD(rcd IRCD) {
	t.RCDs = append(t.RCDs, rcd)
	t.clearCaches()
}

func (Transaction) GetDBHash() IHash {
	return Sha([]byte("Transaction"))
}

func (w1 Transaction) GetNewInstance() IBlock {
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
//    all full nodes. A fee of 10 EC equivalent must be paid for each
//    signature included.
func (t Transaction) CalculateFee(factoshisPerEC uint64) (uint64, error) {

	// First look at the size of the transaction, and make sure
	// everything is inbounds.
	data, err := t.MarshalBinary()
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

	fee += factoshisPerEC * 10 * uint64(len(t.Outputs)+len(t.OutECs))

	for _, rcd := range t.RCDs {
		fee += factoshisPerEC * uint64(rcd.NumberOfSignatures())
	}

	return fee, nil
}

// Checks that the sum of the given amounts do not cross
// a signed boundry.  Returns false if invalid, and the
// sum if valid.  Returns 0 and true if nothing is passed in.
func ValidateAmounts(amts ...uint64) (uint64, error) {
	var sum int64
	for _, amt := range amts {
		if int64(amt) < 0 {
			return 0, fmt.Errorf("Negative amounts are not allowed")
		}
		sum += int64(amt)
		if int64(sum) < 0 {
			return 0, fmt.Errorf("The amounts specified are too large")
		}
	}
	return uint64(sum), nil
}

func (t Transaction) TotalInputs() (sum uint64, err error) {
	if len(t.Inputs) > 255 {
		return 0, fmt.Errorf("The number of inputs must be less than 255")
	}
	for _, input := range t.Inputs {
		sum, err = ValidateAmounts(sum, input.GetAmount())
		if err != nil {
			return 0, fmt.Errorf("Error totalling Inputs: %s", err.Error())
		}
	}
	return
}

func (t Transaction) TotalOutputs() (sum uint64, err error) {
	if len(t.Outputs) > 255 {
		return 0, fmt.Errorf("The number of outputs must be less than 255")
	}
	for _, output := range t.Outputs {
		sum, err = ValidateAmounts(sum, output.GetAmount())
		if err != nil {
			return 0, fmt.Errorf("Error totalling Outputs: %s", err.Error())
		}
	}
	return
}

func (t Transaction) TotalECs() (sum uint64, err error) {
	if len(t.OutECs) > 255 {
		return 0, fmt.Errorf("The number of Entry Credit outputs must be less than 255")
	}
	for _, ec := range t.OutECs {
		sum, err = ValidateAmounts(sum, ec.GetAmount())
		if err != nil {
			return 0, fmt.Errorf("Error totalling Entry Credit outputs: %s", err.Error())
		}
	}
	return
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
//
// The index is the height of the transaction in a Factoid block.  When
// the index == 0, then it means this is the coinbase transaction.
// The coinbase transaction is the "payout" transaction which cannot have
// any inputs, unlike any other transaction which must have at least one
// input.  If the height of the transaction is known, then the index can
// be used to identify the transaction. Otherwise it simply must be > 0
// to indicate it isn't a coinbase transaction.
func (t Transaction) Validate(index int) error {

	// Inputs, outputs, and ecoutputs, must be valid,
	tInputs, err := t.TotalInputs()
	if err != nil {
		return err
	}
	tOutputs, err := t.TotalOutputs()
	if err != nil {
		return err
	}
	tecs, err := t.TotalECs()
	if err != nil {
		return err
	}

	// Inputs cover outputs and ecoutputs.
	if index != 0 && tInputs < tOutputs+tecs {
		return fmt.Errorf("The Inputs of the transaction do not cover the outputs")
	}
	// Cannot have zero inputs.  This means you cannot use this function
	// to validate coinbase transactions, because they cannot have any
	// inputs.
	if len(t.Inputs) == 0 {
		if index > 0 {
			return fmt.Errorf("Transactions (other than the coinbase) must have at least one input")
		}
	} else {
		if index == 0 {
			PrtStk()
			fmt.Println(index, t)
			return fmt.Errorf("Coinbase transactions cannot have inputs.")
		}
	}
	// Every input must have an RCD block
	if len(t.Inputs) != len(t.RCDs) {
		return fmt.Errorf("All inputs must have a cooresponding RCD")
	}
	// Every input must match the address of an RCD (which is the hash
	// of the RCD
	for i, rcd := range t.RCDs {
		// Get the address specified by the RCD.
		address, err := rcd.GetAddress()
		// If there is anything wrong with the RCD, then the transaction isn't
		// valid.
		if err != nil {
			return fmt.Errorf("RCD %d failed to provide an address to compare with its input", i)
		}
		// If the Address (which is really a hash) isn't equal to the hash of
		// the RCD, this transaction is bogus.
		if t.Inputs[i].GetAddress().IsEqual(address) != nil {
			return fmt.Errorf("The %d Input does not match the %d RCD", i, i)
		}
	}

	return nil
}

// This call ONLY checks signatures.  Call ITransaction.Validate() to check the structure of the
// transaction.
//
func (t Transaction) ValidateSignatures() error {
	missingCnt := 0
	sigBlks := t.GetSignatureBlocks()
	for i, rcd := range t.RCDs {
		if !rcd.CheckSig(&t, sigBlks[i]) {
			missingCnt++
		}
	}
	if missingCnt != 0 {
		return fmt.Errorf("Missing %d of %d signatures", missingCnt, len(t.RCDs))
	}

	return nil
}

// Tests if the transaction is equal in all of its structures, and
// in order of the structures.  Largely used to test and debug, but
// generally useful.
func (t1 *Transaction) IsEqual(trans IBlock) []IBlock {

	t2, ok := trans.(ITransaction)

	if !ok || // Not the right kind of IBlock
		len(t1.Inputs) != len(t2.GetInputs()) || // Size of arrays has to match
		len(t1.Outputs) != len(t2.GetOutputs()) || // Size of arrays has to match
		len(t1.OutECs) != len(t2.GetECOutputs()) { // Size of arrays has to match

		r := make([]IBlock, 0, 5)
		return append(r, t1)
	}

	for i, input := range t1.GetInputs() {
		adr, err := t2.GetInput(i)
		if err != nil {
			r := make([]IBlock, 0, 5)
			return append(r, t1)
		}
		r := input.IsEqual(adr)
		if r != nil {
			return append(r, t1)
		}

	}
	for i, output := range t1.GetOutputs() {
		adr, err := t2.GetOutput(i)
		if err != nil {
			r := make([]IBlock, 0, 5)
			return append(r, t1)
		}
		r := output.IsEqual(adr)
		if r != nil {
			return append(r, t1)
		}

	}
	for i, outEC := range t1.GetECOutputs() {
		adr, err := t2.GetECOutput(i)
		if err != nil {
			r := make([]IBlock, 0, 5)
			return append(r, t1)
		}
		r := outEC.IsEqual(adr)
		if r != nil {
			return append(r, t1)
		}

	}
	for i, a := range t1.RCDs {
		adr, err := t2.GetRCD(i)
		if err != nil {
			r := make([]IBlock, 0, 5)
			return append(r, t1)
		}
		r := a.IsEqual(adr)
		if r != nil {
			return append(r, t1)
		}

	}
	for i, s := range t1.SigBlocks {
		r := s.IsEqual(t2.GetSignatureBlock(i))
		if r != nil {
			return append(r, t1)
		}
	}

	return nil
}

func (t Transaction) GetInputs() []IInAddress       { return t.Inputs }
func (t Transaction) GetOutputs() []IOutAddress     { return t.Outputs }
func (t Transaction) GetECOutputs() []IOutECAddress { return t.OutECs }
func (t Transaction) GetRCDs() []IRCD               { return t.RCDs }

func (t *Transaction) GetSignatureBlocks() []ISignatureBlock {
	if len(t.SigBlocks) > len(t.Inputs) { // If too long, nil out
		for i := len(t.Inputs); i < len(t.SigBlocks); i++ { // the extra entries, and
			t.SigBlocks[i] = nil // cut it to length.
		}
		t.SigBlocks = t.SigBlocks[:len(t.Inputs)]
		return t.SigBlocks
	}
	for i := len(t.SigBlocks); i < len(t.Inputs); i++ { // If too short, then
		t.SigBlocks = append(t.SigBlocks, new(SignatureBlock)) // pad it with
	} // signature blocks.
	return t.SigBlocks
}

func (t *Transaction) GetInput(i int) (IInAddress, error) {
	if i > len(t.Inputs) {
		return nil, fmt.Errorf("Index out of Range")
	}
	return t.Inputs[i], nil
}

func (t *Transaction) GetOutput(i int) (IOutAddress, error) {
	if i > len(t.Outputs) {
		return nil, fmt.Errorf("Index out of Range")
	}
	return t.Outputs[i], nil
}

func (t *Transaction) GetECOutput(i int) (IOutECAddress, error) {
	if i > len(t.OutECs) {
		return nil, fmt.Errorf("Index out of Range")
	}
	return t.OutECs[i], nil
}

func (t *Transaction) GetRCD(i int) (IRCD, error) {
	if i > len(t.RCDs) {
		return nil, fmt.Errorf("Index out of Range")
	}
	return t.RCDs[i], nil
}

// UnmarshalBinary assumes that the Binary is all good.  We do error
// out if there isn't enough data, or the transaction is too large.
func (t *Transaction) UnmarshalBinaryData(data []byte) (newData []byte, err error) {

	// To catch memory errors, I capture the panic and turn it into
	// a reported error.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Error unmarshalling transaction: %v", r)
		}
	}()

	// To capture the panic, my code needs to be in a function.  So I'm
	// creating one here, and call it at the end of this function.
	v, data := DecodeVarInt(data)
	if v != t.GetVersion() {
		return nil, fmt.Errorf("Wrong Transaction Version encountered. Expected %v and found %v", t.GetVersion(), v)
	}
	hd, data := binary.BigEndian.Uint32(data[:]), data[4:]
	ld, data := binary.BigEndian.Uint16(data[:]), data[2:]
	t.MilliTimestamp = (uint64(hd) << 16) + uint64(ld)

	numInputs := int(data[0])
	data = data[1:]
	numOutputs := int(data[0])
	data = data[1:]
	numOutECs := int(data[0])
	data = data[1:]

	t.Inputs = make([]IInAddress, numInputs, numInputs)
	t.Outputs = make([]IOutAddress, numOutputs, numOutputs)
	t.OutECs = make([]IOutECAddress, numOutECs, numOutECs)

	for i, _ := range t.Inputs {
		t.Inputs[i] = new(InAddress)
		data, err = t.Inputs[i].UnmarshalBinaryData(data)
		if err != nil || t.Inputs[i] == nil {
			return nil, err
		}
	}
	for i, _ := range t.Outputs {
		t.Outputs[i] = new(OutAddress)
		data, err = t.Outputs[i].UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}
	for i, _ := range t.OutECs {
		t.OutECs[i] = new(OutECAddress)
		data, err = t.OutECs[i].UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}
	}

	t.RCDs = make([]IRCD, len(t.Inputs))
	t.SigBlocks = make([]ISignatureBlock, len(t.Inputs))

	for i := 0; i < len(t.Inputs); i++ {
		t.RCDs[i] = CreateRCD(data)
		data, err = t.RCDs[i].UnmarshalBinaryData(data)
		if err != nil {
			return nil, err
		}

		t.SigBlocks[i] = new(SignatureBlock)
		data, err = t.SigBlocks[i].UnmarshalBinaryData(data)
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
func (t *Transaction) MarshalBinarySig() (newData []byte, err error) {
	var out bytes.Buffer

	EncodeVarInt(&out, t.GetVersion())

	hd := uint32(t.MilliTimestamp >> 16)
	ld := uint16(t.MilliTimestamp & 0xFFFF)
	binary.Write(&out, binary.BigEndian, uint32(hd))
	binary.Write(&out, binary.BigEndian, uint16(ld))

	out.WriteByte(byte(len(t.Inputs)))
	out.WriteByte(byte(len(t.Outputs)))
	out.WriteByte(byte(len(t.OutECs)))

	for _, input := range t.Inputs {
		data, err := input.MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}

	for _, output := range t.Outputs {
		data, err := output.MarshalBinary()
		if err != nil {
			return nil, err
		}
		out.Write(data)
	}

	for _, outEC := range t.OutECs {
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

	for i, rcd := range t.RCDs {

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
		if len(t.SigBlocks) <= i {
			t.SigBlocks = append(t.SigBlocks, new(SignatureBlock))
		}
		data, err = t.SigBlocks[i].MarshalBinary()
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
func (t *Transaction) AddInput(input IAddress, amount uint64) {
	if t.Inputs == nil {
		t.Inputs = make([]IInAddress, 0, 5)
	}
	out := NewInAddress(input, amount)
	t.Inputs = append(t.Inputs, out)
	t.clearCaches()
}

// Helper function for building transactions.  Add an output to
// the transaction.  I'm guessing 5 outputs is about all anyone
// will need, so I'll default to 5.  Of course, go will grow
// past that if needed.
func (t *Transaction) AddOutput(output IAddress, amount uint64) {
	if t.Outputs == nil {
		t.Outputs = make([]IOutAddress, 0, 5)
	}
	out := NewOutAddress(output, amount)
	t.Outputs = append(t.Outputs, out)
	t.clearCaches()
}

// Add a EntryCredit output.  Validating this is going to require
// access to the exchange rate.  This is literally how many entry
// credits are being added to the specified Entry Credit address.
func (t *Transaction) AddECOutput(ecoutput IAddress, amount uint64) {
	if t.OutECs == nil {
		t.OutECs = make([]IOutECAddress, 0, 5)
	}
	out := NewOutECAddress(ecoutput, amount)
	t.OutECs = append(t.OutECs, out)
	t.clearCaches()
}

// Marshal to text.  Largely a debugging thing.
func (t *Transaction) CustomMarshalText() (text []byte, err error) {
	data, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	var out bytes.Buffer
	out.WriteString(fmt.Sprintf("Transaction (size %d):\n", len(data)))
	out.WriteString("                 Version: ")
	WriteNumber64(&out, uint64(t.GetVersion()))
	out.WriteString("\n          MilliTimestamp: ")
	WriteNumber64(&out, uint64(t.MilliTimestamp))
	ts := time.Unix(0, int64(t.MilliTimestamp*1000000))
	out.WriteString(ts.UTC().Format(" Jan 2, 2006 at 3:04am (MST)"))
	out.WriteString("\n                # Inputs: ")
	WriteNumber16(&out, uint16(len(t.Inputs)))
	out.WriteString("\n               # Outputs: ")
	WriteNumber16(&out, uint16(len(t.Outputs)))
	out.WriteString("\n   # EntryCredit Outputs: ")
	WriteNumber16(&out, uint16(len(t.OutECs)))
	out.WriteString("\n")
	for _, address := range t.Inputs {
		text, _ := address.CustomMarshalText()
		out.Write(text)
	}
	for _, address := range t.Outputs {
		text, _ := address.CustomMarshalText()
		out.Write(text)
	}
	for _, ecaddress := range t.OutECs {
		text, _ := ecaddress.CustomMarshalText()
		out.Write(text)
	}
	for i, rcd := range t.RCDs {
		text, err = rcd.CustomMarshalText()
		if err != nil {
			return nil, err
		}
		out.Write(text)

		for len(t.SigBlocks) <= i {
			t.SigBlocks = append(t.SigBlocks, new(SignatureBlock))
		}
		text, err := t.SigBlocks[i].CustomMarshalText()
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
	if t.RCDs == nil {
		t.RCDs = make([]IRCD, 0, 5)
	}
	t.RCDs = append(t.RCDs, auth)
}

func (e *Transaction) JSONByte() ([]byte, error) {
	return EncodeJSON(e)
}

func (e *Transaction) JSONString() (string, error) {
	return EncodeJSONString(e)
}

func (e *Transaction) JSONBuffer(b *bytes.Buffer) error {
	return EncodeJSONToBuffer(e, b)
}

func (e *Transaction) Spew() string {
	return Spew(e)
}
