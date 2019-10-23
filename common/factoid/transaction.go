// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package factoid

import (
	"fmt"
	"os"
	"reflect"
	"runtime/debug"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = debug.PrintStack

type Transaction struct {
	// Not marshalled in MarshalBinary()
	Txid        interfaces.IHash `json:"txid"`
	BlockHeight uint32           `json:"blockheight"`
	sigValid    bool

	// Marshalled in MarshalBinary()
	// version     uint64         Version of transaction. Hardcoded, naturally.
	MilliTimestamp uint64 `json:"millitimestamp"`
	// #inputs     uint8          number of inputs
	// #outputs    uint8          number of outputs
	// #ecoutputs  uint8          number of outECs (Number of EntryCredits)
	Inputs    []interfaces.ITransAddress   `json:"inputs"`
	Outputs   []interfaces.ITransAddress   `json:"outputs"`
	OutECs    []interfaces.ITransAddress   `json:"outecs"`
	RCDs      []interfaces.IRCD            `json:"rcds"`
	SigBlocks []interfaces.ISignatureBlock `json:"sigblocks"`
}

var _ interfaces.ITransaction = (*Transaction)(nil)
var _ interfaces.Printable = (*Transaction)(nil)
var _ interfaces.BinaryMarshallableAndCopyable = (*Transaction)(nil)

func (t *Transaction) IsSameAs(trans interfaces.ITransaction) bool {
	if trans == nil {
		if t == nil {
			return true
		}
		return false
	}
	if t.GetTimestamp().GetTimeMilliUInt64() != trans.GetTimestamp().GetTimeMilliUInt64() {
		return false
	}
	ins := trans.GetInputs()
	if len(t.Inputs) != len(ins) {
		return false
	}
	outs := trans.GetOutputs()
	if len(t.Outputs) != len(outs) {
		return false
	}
	outECs := trans.GetECOutputs()
	if len(t.OutECs) != len(outECs) {
		return false
	}
	rcds := trans.GetRCDs()
	if len(t.RCDs) != len(ins) {
		return false
	}
	sigs := trans.GetSignatureBlocks()
	if len(t.SigBlocks) != len(ins) {
		return false
	}

	for i := range t.Inputs {
		if t.Inputs[i].IsSameAs(ins[i]) == false {
			return false
		}
	}
	for i := range t.Outputs {
		if t.Outputs[i].IsSameAs(outs[i]) == false {
			return false
		}
	}
	for i := range t.OutECs {
		if t.OutECs[i].IsSameAs(outECs[i]) == false {
			return false
		}
	}
	for i := range t.RCDs {
		if t.RCDs[i].IsSameAs(rcds[i]) == false {
			return false
		}
	}
	for i := range t.SigBlocks {
		if t.SigBlocks[i].IsSameAs(sigs[i]) == false {
			return false
		}
	}
	return true
}

func (w *Transaction) New() interfaces.BinaryMarshallableAndCopyable {
	return new(Transaction)
}

func (t *Transaction) SetBlockHeight(height uint32) {
	t.BlockHeight = height
}

func (t *Transaction) GetBlockHeight() (height uint32) {
	return t.BlockHeight
}

// Clears caches if they are no long valid.
func (t *Transaction) clearCaches() {
	return
}

func (*Transaction) GetVersion() uint64 {
	return 2
}

func (t *Transaction) GetTxID() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("Transaction.GetTxID() saw an interface that was nil")
		}
	}()

	return t.GetSigHash()
}

func (t *Transaction) GetHash() (rval interfaces.IHash) {
	defer func() {
		if rval != nil && reflect.ValueOf(rval).IsNil() {
			rval = nil // convert an interface that is nil to a nil interface
			primitives.LogNilHashBug("Transaction.GetHash() saw an interface that was nil")
		}
	}()

	m, err := t.MarshalBinary()
	if err != nil {
		return nil
	}
	return primitives.Sha(m)
}

func (t Transaction) GetFullHash() interfaces.IHash {
	m, err := t.MarshalBinary()
	if err != nil {
		return nil
	}
	return primitives.Sha(m)
}

func (t Transaction) GetSigHash() interfaces.IHash {
	m, err := t.MarshalBinarySig()
	if err != nil {
		return nil
	}
	return primitives.Sha(m)
}

func (t Transaction) String() string {
	txt, err := t.CustomMarshalText()
	if err != nil {
		return "<error>"
	}
	return string(txt)
}

// MilliTimestamp is in milliseconds
func (t *Transaction) GetTimestamp() interfaces.Timestamp {
	return primitives.NewTimestampFromMilliseconds(t.MilliTimestamp)
}

func (t *Transaction) SetTimestamp(ts interfaces.Timestamp) {
	milli := ts.GetTimeMilliUInt64()

	if milli != t.MilliTimestamp {
		//		messages.LogPrintf("timestamps.txt", "transaction %p changed from %d to %d @ %s", t, t.MilliTimestamp, milli, atomic.WhereAmIString(1))
	}
	t.MilliTimestamp = milli
}

func (t *Transaction) SetSignatureBlock(i int, sig interfaces.ISignatureBlock) {
	for len(t.SigBlocks) <= i {
		t.SigBlocks = append(t.SigBlocks, new(SignatureBlock))
	}
	t.SigBlocks[i] = sig
}

func (t *Transaction) GetSignatureBlock(i int) interfaces.ISignatureBlock {
	for len(t.SigBlocks) <= i {
		t.SigBlocks = append(t.SigBlocks, new(SignatureBlock))
	}
	return t.SigBlocks[i]
}

func (t *Transaction) AddRCD(rcd interfaces.IRCD) {
	t.RCDs = append(t.RCDs, rcd)
	t.clearCaches()
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
	if len(data) > constants.MAX_TRANSACTION_SIZE { // Can't be bigger than our limits
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
// a signed boundary.  Returns false if invalid, and the
// sum if valid.  Returns 0 and true if nothing is passed in.
func ValidateAmounts(amts ...uint64) (uint64, error) {
	var sum int64
	for _, amt := range amts {
		if int64(amt) < 0 {
			return 0, fmt.Errorf("Amount is out of range")
		}
		sum += int64(amt)
		if int64(sum) < 0 {
			return 0, fmt.Errorf("Amounts on the transaction are out of range")
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
			fmt.Println(index, t)
			return fmt.Errorf("Coinbase transactions cannot have inputs.")
		}
	}
	// Every input must have an RCD block
	if len(t.Inputs) != len(t.RCDs) {
		return fmt.Errorf("All inputs must have a corresponding RCD")
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
		if t.Inputs[i].GetAddress().IsSameAs(address) == false {
			return fmt.Errorf("The %d Input does not match the %d RCD", i, i)
		}
	}

	return nil
}

// This call ONLY checks signatures.  Call interfaces.ITransaction.Validate() to check the structure of the
// transaction.
//
func (t Transaction) ValidateSignatures() error {
	if !t.sigValid {
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
		t.sigValid = true
	}
	return nil
}

func (t Transaction) GetInputs() []interfaces.ITransAddress    { return t.Inputs }
func (t Transaction) GetOutputs() []interfaces.ITransAddress   { return t.Outputs }
func (t Transaction) GetECOutputs() []interfaces.ITransAddress { return t.OutECs }
func (t Transaction) GetRCDs() []interfaces.IRCD               { return t.RCDs }

func (t *Transaction) GetSignatureBlocks() []interfaces.ISignatureBlock {
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

func (t *Transaction) GetInput(i int) (interfaces.ITransAddress, error) {
	if i > len(t.Inputs) {
		return nil, fmt.Errorf("Index out of Range")
	}
	return t.Inputs[i], nil
}

func (t *Transaction) GetOutput(i int) (interfaces.ITransAddress, error) {
	if i > len(t.Outputs) {
		return nil, fmt.Errorf("Index out of Range")
	}
	return t.Outputs[i], nil
}

func (t *Transaction) GetECOutput(i int) (interfaces.ITransAddress, error) {
	if i > len(t.OutECs) {
		return nil, fmt.Errorf("Index out of Range")
	}
	return t.OutECs[i], nil
}

func (t *Transaction) GetRCD(i int) (interfaces.IRCD, error) {
	if i > len(t.RCDs) {
		return nil, fmt.Errorf("Index out of Range")
	}
	return t.RCDs[i], nil
}

// UnmarshalBinary assumes that the Binary is all good.  We do error
// out if there isn't enough data, or the transaction is too large.
func (t *Transaction) UnmarshalBinaryData(data []byte) ([]byte, error) {
	buf := primitives.NewBuffer(data)

	v, err := buf.PopVarInt()
	if err != nil {
		return nil, err
	}
	if v != t.GetVersion() {
		return nil, fmt.Errorf("Wrong Transaction Version encountered. Expected %v and found %v", t.GetVersion(), v)
	}

	hd, err := buf.PopUInt32()
	if err != nil {
		return nil, err
	}
	ld, err := buf.PopUInt16()
	if err != nil {
		return nil, err
	}
	t.MilliTimestamp = (uint64(hd) << 16) + uint64(ld)

	numInputs, err := buf.PopUInt8()
	if err != nil {
		return nil, err
	}
	numOutputs, err := buf.PopUInt8()
	if err != nil {
		return nil, err
	}
	numOutECs, err := buf.PopUInt8()
	if err != nil {
		return nil, err
	}

	t.Inputs = make([]interfaces.ITransAddress, int(numInputs), int(numInputs))
	t.Outputs = make([]interfaces.ITransAddress, int(numOutputs), int(numOutputs))
	t.OutECs = make([]interfaces.ITransAddress, int(numOutECs), int(numOutECs))

	for i, _ := range t.Inputs {
		t.Inputs[i] = new(TransAddress)
		err = buf.PopBinaryMarshallable(t.Inputs[i])
		if err != nil {
			return nil, err
		}
		t.Inputs[i].(*TransAddress).UserAddress = primitives.ConvertFctAddressToUserStr(t.Inputs[i].(*TransAddress).Address)
	}
	for i, _ := range t.Outputs {
		t.Outputs[i] = new(TransAddress)
		err = buf.PopBinaryMarshallable(t.Outputs[i])
		if err != nil {
			return nil, err
		}
		t.Outputs[i].(*TransAddress).UserAddress = primitives.ConvertFctAddressToUserStr(t.Outputs[i].(*TransAddress).Address)
	}
	for i, _ := range t.OutECs {
		t.OutECs[i] = new(TransAddress)
		err = buf.PopBinaryMarshallable(t.OutECs[i])
		if err != nil {
			return nil, err
		}
		t.OutECs[i].(*TransAddress).UserAddress = primitives.ConvertECAddressToUserStr(t.OutECs[i].(*TransAddress).Address)
	}

	t.RCDs = make([]interfaces.IRCD, len(t.Inputs))
	t.SigBlocks = make([]interfaces.ISignatureBlock, len(t.Inputs))

	for i := 0; i < len(t.Inputs); i++ {
		b, err := buf.PeekByte()
		if err != nil {
			return nil, err
		}
		t.RCDs[i] = CreateRCD([]byte{b})
		err = buf.PopBinaryMarshallable(t.RCDs[i])
		if err != nil {
			return nil, err
		}
		t.SigBlocks[i] = new(SignatureBlock)
		err = buf.PopBinaryMarshallable(t.SigBlocks[i])
		if err != nil {
			return nil, err
		}
	}

	t.Txid = t.GetSigHash()
	return buf.DeepCopyBytes(), nil
}

func (t *Transaction) UnmarshalBinary(data []byte) (err error) {
	data, err = t.UnmarshalBinaryData(data)
	return err
}

// This is what Gets Signed.  Yet signature blocks are part of the transaction.
// We don't include them here, and tack them on later.
func (t *Transaction) MarshalBinarySig() (rval []byte, err error) {
	defer func(pe *error) {
		if *pe != nil {
			fmt.Fprintf(os.Stderr, "Transaction.MarshalBinarySig err:%v", *pe)
		}
	}(&err)
	buf := primitives.NewBuffer(nil)

	err = buf.PushVarInt(t.GetVersion())
	if err != nil {
		return nil, err
	}

	hd := uint32(t.MilliTimestamp >> 16)
	ld := uint16(t.MilliTimestamp & 0xFFFF)

	err = buf.PushUInt32(hd)
	if err != nil {
		return nil, err
	}
	err = buf.PushUInt16(ld)
	if err != nil {
		return nil, err
	}

	err = buf.PushByte(byte(len(t.Inputs)))
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(len(t.Outputs)))
	if err != nil {
		return nil, err
	}
	err = buf.PushByte(byte(len(t.OutECs)))
	if err != nil {
		return nil, err
	}

	for _, input := range t.Inputs {
		err = buf.PushBinaryMarshallable(input)
		if err != nil {
			return nil, err
		}
	}
	for _, output := range t.Outputs {
		err = buf.PushBinaryMarshallable(output)
		if err != nil {
			return nil, err
		}
	}
	for _, outEC := range t.OutECs {
		err = buf.PushBinaryMarshallable(outEC)
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

// This just Marshals what gets signed, i.e. MarshalBinarySig(), then
// Marshals the signatures and the RCDs for this transaction.
func (t Transaction) MarshalBinary() ([]byte, error) {
	data, err := t.MarshalBinarySig()
	if err != nil {
		return nil, err
	}
	buf := primitives.NewBuffer(data)

	for i, rcd := range t.RCDs {
		// Write the RCD
		err = buf.PushBinaryMarshallable(rcd)
		if err != nil {
			return nil, err
		}

		// Then write its signature blocks.  This needs to be
		// reworked so we use the information from the RCD block
		// to control the writing of the signatures.  After all,
		// we don't want to restrict what might be required to
		// sign an input.
		if len(t.SigBlocks) <= i {
			t.SigBlocks = append(t.SigBlocks, new(SignatureBlock))
		}
		err = buf.PushBinaryMarshallable(t.SigBlocks[i])
		if err != nil {
			return nil, err
		}
	}

	return buf.DeepCopyBytes(), nil
}

// Helper function for building transactions.  Add an input to
// the transaction.  I'm guessing 5 inputs is about all anyone
// will need, so I'll default to 5.  Of course, go will grow
// past that if needed.
func (t *Transaction) AddInput(input interfaces.IAddress, amount uint64) {
	if t.Inputs == nil {
		t.Inputs = make([]interfaces.ITransAddress, 0, 5)
	}
	out := NewInAddress(input, amount)
	t.Inputs = append(t.Inputs, out)
	t.clearCaches()
}

// Helper function for building transactions.  Add an output to
// the transaction.  I'm guessing 5 outputs is about all anyone
// will need, so I'll default to 5.  Of course, go will grow
// past that if needed.
func (t *Transaction) AddOutput(output interfaces.IAddress, amount uint64) {
	if t.Outputs == nil {
		t.Outputs = make([]interfaces.ITransAddress, 0, 5)
	}
	out := NewOutAddress(output, amount)
	t.Outputs = append(t.Outputs, out)
	t.clearCaches()
}

// Add a EntryCredit output.  Validating this is going to require
// access to the exchange rate.  This is literally how many entry
// credits are being added to the specified Entry Credit address.
func (t *Transaction) AddECOutput(ecoutput interfaces.IAddress, amount uint64) {
	if t.OutECs == nil {
		t.OutECs = make([]interfaces.ITransAddress, 0, 5)
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

	txid := fmt.Sprintf("%64s", "coinbase") //make it the same length as a real TXID
	if t.Txid != nil {
		txid = fmt.Sprintf("%x", t.Txid.Bytes())
	}

	var out primitives.Buffer
	out.WriteString(fmt.Sprintf("Transaction TXID: %s (size %d):\n", txid, len(data)))
	out.WriteString("                 Version: ")
	primitives.WriteNumber64(&out, uint64(t.GetVersion()))
	out.WriteString("\n          MilliTimestamp: ")
	primitives.WriteNumber64(&out, uint64(t.MilliTimestamp))
	ts := time.Unix(0, int64(t.MilliTimestamp*1000000))
	out.WriteString(ts.UTC().Format(" Jan 2, 2006 at 15:04:05 (MST)"))
	out.WriteString("\n                # Inputs: ")
	primitives.WriteNumber16(&out, uint16(len(t.Inputs)))
	out.WriteString("\n               # Outputs: ")
	primitives.WriteNumber16(&out, uint16(len(t.Outputs)))
	out.WriteString("\n   # EntryCredit Outputs: ")
	primitives.WriteNumber16(&out, uint16(len(t.OutECs)))
	out.WriteString("\n")
	for _, address := range t.Inputs {
		text, _ := address.CustomMarshalTextInput()
		out.Write(text)
	}
	for _, address := range t.Outputs {
		text, _ := address.CustomMarshalTextOutput()
		out.Write(text)
	}
	for _, ecaddress := range t.OutECs {
		text, _ := ecaddress.CustomMarshalTextECOutput()
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

	return out.DeepCopyBytes(), nil
}

// Helper Function.  This simply adds an Authorization to a
// transaction.  DOES NO VALIDATION.  Not the job of construction.
// That's why we have a validation call.
func (t *Transaction) AddAuthorization(auth interfaces.IRCD) {
	if t.RCDs == nil {
		t.RCDs = make([]interfaces.IRCD, 0, 5)
	}
	t.RCDs = append(t.RCDs, auth)
}

func (e *Transaction) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *Transaction) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *Transaction) HasUserAddress(userAddr string) bool {
	//  do any of the inputs or outputs of this transaction belong to the inputed user address
	// Other than a minimal length check, this does not address validation of the requested user address
	// in some cases, the useraddress is not being filled in the address struct.  if it is blank, convert the address (hash)
	var matchString string

	//  I am filtering for this because I do not want to bother checking for an "EC" start if it is too short
	if len(userAddr) < 2 {
		return false
	}
	// if this address starts with EC it can only be an EC output address.  No need to check any others.

	if userAddr[0:2] == "EC" {
		ecoutputs := e.GetECOutputs()
		for _, addLine := range ecoutputs {
			if addLine.GetUserAddress() == "" {
				matchString = primitives.ConvertECAddressToUserStr(addLine.GetAddress())
			} else {
				matchString = addLine.GetUserAddress()
			}
			if matchString == userAddr {
				return true
			}
		}
	} else {
		// if it is NOT an ec address, it can't be a factoid address, so don't check those.
		// check input addresses
		inputs := e.GetInputs()
		for _, addLine := range inputs {
			if addLine.GetUserAddress() == "" {
				matchString = primitives.ConvertFctAddressToUserStr(addLine.GetAddress())
			} else {
				matchString = addLine.GetUserAddress()
			}
			if matchString == userAddr {
				return true
			}
		}

		//check output addresses
		outputs := e.GetOutputs()
		for _, addLine := range outputs {
			if addLine.GetUserAddress() == "" {
				matchString = primitives.ConvertFctAddressToUserStr(addLine.GetAddress())
			} else {
				matchString = addLine.GetUserAddress()
			}
			if matchString == userAddr {
				return true
			}
		}
	}
	// if it was found, it would have already returned
	return false
}
