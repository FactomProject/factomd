package state

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/FactomProject/btcutil/base58"
	ed "github.com/FactomProject/ed25519"
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
		this.Println("Couldn't find the FER chain for id ", this.FERChainId)
		return
	}
	if entryBlock == nil {
		this.Println("FER Chain head found to be nil")
		return
	}

	this.Println("Node node = ", this.FactomNodeName)
	this.Println("Checking last e block of FER chain with height of: ", entryBlock.GetHeader().GetDBHeight())
	this.Println("Current block height: ", this.GetDBHeightComplete())
	this.Println("BEFORE processing recent block: ")
	this.Println("    FERChangePrice: ", this.FERChangePrice)
	this.Println("    FERChangeHeight: ", this.FERChangeHeight)
	this.Println("    FERPriority: ", this.FERPriority)
	this.Println("    FERPrioritySetHeight: ", this.FERPrioritySetHeight)
	this.Println("    FER current: ", this.GetFactoshisPerEC())

	// Check to see if a price change targets the next block
	if this.FERChangeHeight == (this.GetDBHeightComplete() + 1) {
		this.FactoshisPerEC = this.FERChangePrice
		this.FERChangePrice = 100000000
		this.FERChangeHeight = 0
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
	if entryBlock.GetHeader().GetDBHeight() == this.GetDBHeightComplete() {
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
			anFEREntry := new(FEREntry)
			err = json.Unmarshal(entryContent, &anFEREntry)
			if err != nil {
				this.Println("A FEREntry messgae didn't unmarshall correctly: ", err)
				continue
			}

			// Set it's resident height for validity checking
			anFEREntry.SetResidentHeight(this.GetDBHeightComplete())

			if (this.FerEntryIsValid(anFEREntry)) && (anFEREntry.Priority > this.FERPriority) {

				fmt.Println(" Processing FER entry : ", string(entryContent))
				this.FERPriority = anFEREntry.GetPriority()
				this.FERPrioritySetHeight = this.GetDBHeightComplete()
				this.FERChangePrice = anFEREntry.GetTargetPrice()
				this.FERChangeHeight = anFEREntry.GetTargetActivationHeight()

				// Adjust the target if needed
				if this.FERChangeHeight < (this.GetDBHeightComplete() + 2) {
					this.FERChangeHeight = this.GetDBHeightComplete() + 2
				}
			} else {
				fmt.Println(" Failed FER entry : ", string(entryContent))
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

	// convert the conf quthority address into a
	authorityAddress := base58.Decode(this.ExchangeRateAuthorityAddress)
	ecPubPrefix := []byte{0x59, 0x2a}

	if !bytes.Equal(authorityAddress[:2], ecPubPrefix) {
		fmt.Errorf("Invalid Entry Credit Private Address")
		return false
	}

	pub := new([32]byte)
	copy(pub[:], authorityAddress[2:34])

	// in case verify can't handle empty public key
	if this.ExchangeRateAuthorityAddress == "" {
		return false
	}
	sig := new([64]byte)
	externalIds := e.ExternalIDs()

	// check for number of ext ids
	if len(externalIds) < 1 {
		return false
	}

	copy(sig[:], externalIds[0]) // First ext id needs to be the signature of the content

	if !ed.Verify(pub, e.GetContent(), sig) {
		return false
	}

	return true
}

func (this *State) FerEntryIsValid(passedFEREntry interfaces.IFEREntry) bool {
	// fail if expired
	if passedFEREntry.GetExpirationHeight() < passedFEREntry.GetResidentHeight() {
		return false
	}

	// fail if expired height is too far out
	if passedFEREntry.GetExpirationHeight() > (passedFEREntry.GetResidentHeight() + 6) {
		return false
	}

	// fail if target is out of range of the expire height
	// The check for expire height >= 6 is import because a lower value results in a uint binary wrap to near maxint, cracks logic
	if (passedFEREntry.GetTargetActivationHeight() > (passedFEREntry.GetExpirationHeight() + 6)) ||
		((passedFEREntry.GetExpirationHeight() >= 6) &&
			(passedFEREntry.GetTargetActivationHeight() < (passedFEREntry.GetExpirationHeight() - 6))) {

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
