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
func (s *State) ProcessRecentFERChainEntries() {
	// Find the FER entry chain
	FERChainHash, err := primitives.HexToHash(s.FERChainId)
	if err != nil {
		s.Println("The FERChainId couldn't be turned into a IHASH")
		return
	}

	//  Get the first eblock from the FERChain
	entryBlock, err := s.DB.FetchEBlockHead(FERChainHash)
	if err != nil {
		//		packageLogger.Debugf("FER Chain head found to be nil %v", s.FERChainId)
		// s.Println("Couldn't find the FER chain for id ", s.FERChainId)
		return
	}
	if entryBlock == nil {
		//		packageLogger.Debug("FER Chain head found to be nil")
		// s.Println("FER Chain head found to be nil")
		return
	}

	s.Println("Checking last e block of FER chain with height of: ", entryBlock.GetHeader().GetDBHeight())
	s.Println("Current block height: ", s.GetDBHeightComplete())
	s.Println("BEFORE processing recent block: ")
	s.Println("    FERChangePrice: ", s.FERChangePrice)
	s.Println("    FERChangeHeight: ", s.FERChangeHeight)
	s.Println("    FERPriority: ", s.FERPriority)
	s.Println("    FERPrioritySetHeight: ", s.FERPrioritySetHeight)
	s.Println("    FER current: ", s.GetFactoshisPerEC())

	// Check to see if a price change targets the next block
	if s.FERChangeHeight == (s.GetDBHeightComplete())+1 {
		s.FactoshisPerEC = s.FERChangePrice
		s.FERChangePrice = 0
		s.FERChangeHeight = 1
	}

	// Check for the need to clear the priority
	// (s.GetDBHeightComplete() >= 12) is import because height is a uint and can't break logic if subtracted into false sub-zero
	if (s.GetDBHeightComplete() >= 12) &&
		(s.GetDBHeightComplete()-12) >= s.FERPrioritySetHeight {
		s.FERPrioritySetHeight = 0
		s.FERPriority = 0
		// Now the next entry to come through with a priority of 1 or more will be considered
	}

	// Check last entry block method
	if entryBlock.GetHeader().GetDBHeight() == s.GetDBHeightComplete()-1 {
		entryHashes := entryBlock.GetEntryHashes()

		// s.Println("Found FER entry hashes in a block as: ", entryHashes)
		// create a map of possible minute markers that may be found in the EBlock Body
		mins := make(map[string]uint8)
		for i := byte(0); i <= 10; i++ {
			h := make([]byte, 32)
			h[len(h)-1] = i
			mins[hex.EncodeToString(h)] = i
		}

		// Loop through the hashes from the last blocks FER entries and evaluate them individually
		for _, entryHash := range entryHashes {
			// if s entryhash is a minute mark then continue
			if _, exist := mins[entryHash.String()]; exist {
				continue
			}

			// Make sure the entry exists
			anEntry, err := s.DB.FetchEntry(entryHash)
			if err != nil {
				s.Println("Error during FetchEntryByHash: ", err)
				continue
			}
			if anEntry == nil {
				s.Println("Nil entry during FetchEntryByHash: ", entryHash)
				continue
			}

			if !s.ExchangeRateAuthorityIsValid(anEntry) {
				s.Println("Skipping non-authority FER chain entry", entryHash)
				continue
			}

			entryContent := anEntry.GetContent()
			// s.Println("Found content of an FER entry is:  ", string(entryContent))
			ferEntry := new(specialEntries.FEREntry)
			err = ferEntry.UnmarshalBinary(entryContent)
			if err != nil {
				s.Println("A FEREntry messgae didn't unmarshal correctly: ", err)
				continue
			}

			// Set it's resident height for validity checking
			ferEntry.SetResidentHeight(s.GetDBHeightComplete())

			if (s.FerEntryIsValid(ferEntry)) && (ferEntry.Priority > s.FERPriority) {
				s.Println(" Processing FER entry : ", string(entryContent))
				s.FERPriority = ferEntry.GetPriority()
				s.FERPrioritySetHeight = s.GetDBHeightComplete()
				s.FERChangePrice = ferEntry.GetTargetPrice()
				s.FERChangeHeight = ferEntry.GetTargetActivationHeight()

				// Adjust the target if needed
				if s.FERChangeHeight < (s.GetDBHeightComplete() + 2) {
					s.FERChangeHeight = s.GetDBHeightComplete() + 2
				}
			} else {
				s.Println(" Failed FER entry : ", string(entryContent))
			}
		}
	}

	s.Println("AFTER processing recent block: ")
	s.Println("    FERChangePrice: ", s.FERChangePrice)
	s.Println("    FERChangeHeight: ", s.FERChangeHeight)
	s.Println("    FERPriority: ", s.FERPriority)
	s.Println("    FERPrioritySetHeight: ", s.FERPrioritySetHeight)
	s.Println("    FER current: ", s.GetFactoshisPerEC())
	s.Println("----------------------------------")

	return
}

func (s *State) ExchangeRateAuthorityIsValid(e interfaces.IEBEntry) bool {
	pubStr, err := factoid.PublicKeyStringToECAddressString(s.ExchangeRateAuthorityPublicKey)
	if err != nil {
		return false
	}
	// convert the conf quthority Address into a
	authorityAddress := base58.Decode(pubStr)
	ecPubPrefix := []byte{0x59, 0x2a}

	if !bytes.Equal(authorityAddress[:2], ecPubPrefix) {
		fmt.Errorf("Invalid Entry Credit Private Address")
		return false
	}

	pub := new([32]byte)
	copy(pub[:], authorityAddress[2:34])

	// in case verify can't handle empty public key
	if s.ExchangeRateAuthorityPublicKey == "" {
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

func (s *State) FerEntryIsValid(passedFEREntry interfaces.IFEREntry) bool {
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
func (s *State) GetPredictiveFER() uint64 {
	currentFER := s.GetFactoshisPerEC()

	if (s.FERChangeHeight == 0) || // Check to see if no change has been registered
		(s.FERChangePrice <= currentFER) {
		return currentFER
	}

	return s.FERChangePrice
}
