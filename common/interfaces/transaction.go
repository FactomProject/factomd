// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

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

	// Get the hash of the signed portion (not including signatures)
	GetSigHash() IHash

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
	
	// Wallet Support (Not sure why we need some of these)
	SetBlockHeight(int)
}
