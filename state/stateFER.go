// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/FactomProject/btcutil/base58"
	ed "github.com/FactomProject/ed25519"
	"github.com/FactomProject/factomd/common/entryBlock/specialEntries"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

// Go through the factoid exchange rate chain and determine if an FER change should be scheduled
func (this *State) ProcessRecentFERChainEntries() {
	// Find the FER entry chain
	FERChainHash, err := primitives.HexToHash(this.FERChainId)
	if err != nil {
		this.Println("The FERChainId couldn't be turned into a IHASH")
		return
	}

	//  Get the first eblock from the FERChain
	entryBlock, err := this.DB.FetchEBlockHead(FERChainHash)
	if err != nil {
		packageLogger.Debugf("FER Chain head found to be nil %v", this.FERChainId)
		// this.Println("Couldn't find the FER chain for id ", this.FERChainId)
		return
	}
	if entryBlock == nil {
		packageLogger.Debug("FER Chain head found to be nil")
		// this.Println("FER Chain head found to be nil")
		return
	}

	this.Println("Checking last e block of FER chain with height of: ", entryBlock.GetHeader().GetDBHeight())
	this.Println("Current block height: ", this.GetDBHeightComplete())
	this.Println("BEFORE processing recent block: ")
	this.Println("    FERChangePrice: ", this.FERChangePrice)
	this.Println("    FERChangeHeight: ", this.FERChangeHeight)
	this.Println("    FERPriority: ", this.FERPriority)
	this.Println("    FERPrioritySetHeight: ", this.FERPrioritySetHeight)
	this.Println("    FER current: ", this.GetFactoshisPerEC())

	// Check to see if a price change targets the next block
	if this.FERChangeHeight == (this.GetDBHeightComplete())+1 {
		this.FactoshisPerEC = this.FERChangePrice
		this.FERChangePrice = 0
		this.FERChangeHeight = 1
	}

	// Check for the need to clear the priority
	// (this.GetDBHeightComplete() >= 12) is import because height is a uint and can't break logic if subtracted into false sub-zero
	if (this.GetDBHeightComplete() >= 12) &&
		(this.GetDBHeightComplete()-12) >= this.FERPrioritySetHeight {
		this.FERPrioritySetHeight = 0
		this.FERPriority = 0
		// Now the next entry to come through with a priority of 1 or more will be considered
	}

	// Check last entry block method
	if entryBlock.GetHeader().GetDBHeight() == this.GetDBHeightComplete()-1 {
		entryHashes := entryBlock.GetEntryHashes()

		// this.Println("Found FER entry hashes in a block as: ", entryHashes)
		// create a map of possible minute markers that may be found in the EBlock Body
		mins := make(map[string]uint8)
		for i := byte(0); i <= 10; i++ {
			h := make([]byte, 32)
			h[len(h)-1] = i
			mins[hex.EncodeToString(h)] = i
		}

		// Loop through the hashes from the last blocks FER entries and evaluate them individually
		for _, entryHash := range entryHashes {
			// if this entryhash is a minute mark then continue
			if _, exist := mins[entryHash.String()]; exist {
				continue
			}

			// Make sure the entry exists
			anEntry, err := this.DB.FetchEntry(entryHash)
			if err != nil {
				this.Println("Error during FetchEntryByHash: ", err)
				continue
			}
			if anEntry == nil {
				this.Println("Nil entry during FetchEntryByHash: ", entryHash)
				continue
			}

			if !this.ExchangeRateAuthorityIsValid(anEntry) {
				this.Println("Skipping non-authority FER chain entry", entryHash)
				continue
			}

			entryContent := anEntry.GetContent()
			// this.Println("Found content of an FER entry is:  ", string(entryContent))
			ferEntry := new(specialEntries.FEREntry)
			err = ferEntry.UnmarshalBinary(entryContent)
			if err != nil {
				this.Println("A FEREntry messgae didn't unmarshal correctly: ", err)
				continue
			}

			// Set it's resident height for validity checking
			ferEntry.SetResidentHeight(this.GetDBHeightComplete())

			if (this.FerEntryIsValid(ferEntry)) && (ferEntry.Priority > this.FERPriority) {
				this.Println(" Processing FER entry : ", string(entryContent))
				this.FERPriority = ferEntry.GetPriority()
				this.FERPrioritySetHeight = this.GetDBHeightComplete()
				this.FERChangePrice = ferEntry.GetTargetPrice()
				this.FERChangeHeight = ferEntry.GetTargetActivationHeight()

				// Adjust the target if needed
				if this.FERChangeHeight < (this.GetDBHeightComplete() + 2) {
					this.FERChangeHeight = this.GetDBHeightComplete() + 2
				}
			} else {
				this.Println(" Failed FER entry : ", string(entryContent))
			}
		}
	}

	this.Println("AFTER processing recent block: ")
	this.Println("    FERChangePrice: ", this.FERChangePrice)
	this.Println("    FERChangeHeight: ", this.FERChangeHeight)
	this.Println("    FERPriority: ", this.FERPriority)
	this.Println("    FERPrioritySetHeight: ", this.FERPrioritySetHeight)
	this.Println("    FER current: ", this.GetFactoshisPerEC())
	this.Println("----------------------------------")

	return
}

func (this *State) ExchangeRateAuthorityIsValid(e interfaces.IEBEntry) bool {
	pubStr, err := factoid.PublicKeyStringToECAddressString(this.ExchangeRateAuthorityPublicKey)
	if err != nil {
		return false
	}
	// convert the conf quthority Address into a
	authorityAddress := base58.Decode(pubStr)
	ecPubPrefix := []byte{0x59, 0x2a}

	if !bytes.Equal(authorityAddress[:2], ecPubPrefix) {
		fmt.Println("Invalid Entry Credit Private Address")
		return false
	}

	pub := new([32]byte)
	copy(pub[:], authorityAddress[2:34])

	// in case verify can't handle empty public key
	if this.ExchangeRateAuthorityPublicKey == "" {
		return false
	}
	sig := new([64]byte)
	externalIds := e.ExternalIDs()

	// check for number of ext ids
	if len(externalIds) < 1 {
		return false
	}

	copy(sig[:], externalIds[0]) // First ext id needs to be the signature of the content

	if !ed.VerifyCanonical(pub, e.GetContent(), sig) {
		return false
	}

	return true
}

func (this *State) FerEntryIsValid(passedFEREntry interfaces.IFEREntry) bool {
	// fail if expired
	if passedFEREntry.GetExpirationHeight() < passedFEREntry.GetResidentHeight() {
		fmt.Println("FER Failed-fail if expired")
		return false
	}

	// fail if expired height is too far out
	if passedFEREntry.GetExpirationHeight() > (passedFEREntry.GetResidentHeight() + 12) {
		fmt.Println("FER Failed-fail if expired height is too far out")

		return false
	}

	// fail if target is out of range of the expire height
	// The check for expire height >= 6 is import because a lower value results in a uint binary wrap to near maxint, cracks logic
	if (passedFEREntry.GetTargetActivationHeight() > (passedFEREntry.GetExpirationHeight() + 6)) ||
		((passedFEREntry.GetExpirationHeight() >= 6) &&
			(passedFEREntry.GetTargetActivationHeight() < (passedFEREntry.GetExpirationHeight() - 6))) {
		fmt.Println("FER Failed-fail if target is out of range of the expire height")
		return false
	}

	return true
}

// Returns the higher of the current factoid exchange rate and what it knows will change in the future
func (this *State) GetPredictiveFER() uint64 {
	currentFER := this.GetFactoshisPerEC()

	if (this.FERChangeHeight == 0) || // Check to see if no change has been registered
		(this.FERChangePrice <= currentFER) {
		return currentFER
	}

	return this.FERChangePrice
}
