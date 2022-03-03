package snapshot

import (
	"fmt"
	"io"

	"github.com/FactomProject/factomd/state"

	"github.com/sirupsen/logrus"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/entryCreditBlock"
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/database/databaseOverlay"
)

type balanceSnapshot struct {
	NextHeight uint32
	// Use int64s because temporarily during the math, we might have a negative balance
	FCTAddressMap map[[32]byte]int64
	ECAddressMap  map[[32]byte]int64
}

func newBalanceSnapshot() *balanceSnapshot {
	return &balanceSnapshot{
		NextHeight:    0,
		FCTAddressMap: make(map[[32]byte]int64),
		ECAddressMap:  make(map[[32]byte]int64),
	}
}

func (bs *balanceSnapshot) Dump(w io.Writer) error {
	height := bs.NextHeight - 1
	_, _ = fmt.Fprintf(w, "height %d\n", height)
	for k, v := range bs.FCTAddressMap {
		_, err := fmt.Fprintf(w, "%s: %d", primitives.ConvertFctAddressToUserStr(factoid.NewAddress(k[:])), v)
		if err != nil {
			return fmt.Errorf("write fct addr: %w", err)
		}
	}
	for k, v := range bs.ECAddressMap {
		_, err := fmt.Fprintf(w, "%s: %d", primitives.ConvertECAddressToUserStr(factoid.NewAddress(k[:])), v)
		if err != nil {
			return fmt.Errorf("write ec addr: %w", err)
		}
	}

	return nil
}

// Process will process the height specified and load the balance changes into the memory maps.
// We pass the entire database to allow this function to do w/e it needs.
// Passing the height in explicitly just ensures we are loading blocks sequentially
func (bs *balanceSnapshot) Process(log *logrus.Logger, db *databaseOverlay.Overlay, height uint32, diagnostic bool) error {
	defer func() {
		bs.NextHeight++
	}()

	if height != bs.NextHeight {
		return fmt.Errorf("heights must be processed in sequence, exp %d, got %d", bs.NextHeight, height)
	}

	fblock, err := db.FetchFBlockByHeight(height)
	if err != nil {
		return fmt.Errorf("fetch fblock %d: %w", height, err)
	}

	// Update all FCT & EC balances from the factoid transactions.
	addressesChanged := make(map[[32]byte]bool)
	for _, t := range fblock.GetTransactions() {
		for _, input := range t.GetInputs() {
			addr := input.GetAddress().Fixed()
			addressesChanged[addr] = true
			bs.FCTAddressMap[addr] -= int64(input.GetAmount())
		}
		for _, output := range t.GetOutputs() {
			addr := output.GetAddress().Fixed()
			addressesChanged[addr] = true
			bs.FCTAddressMap[addr] += int64(output.GetAmount())
		}
		for _, output := range t.GetECOutputs() {
			fctAmt := output.GetAmount()

			addr := output.GetAddress().Fixed()
			addressesChanged[addr] = false
			bs.ECAddressMap[addr] += int64(fctAmt / fblock.GetExchRate())
		}
	}

	// Debug if any negative balances at the end of an FBlock
	for addr, fct := range addressesChanged {
		var amt int64
		if fct {
			amt = bs.FCTAddressMap[addr]
		} else {
			amt = bs.ECAddressMap[addr]
		}

		debugIfNeg(log, addr, amt, fct, height)
	}

	ecBlock, err := db.FetchECBlockByHeight(height)
	if err != nil {
		return fmt.Errorf("fetch ecblock %d: %w", height, err)
	}
	if ecBlock == nil {
		// ECBlocks 70386-70411 do not exists
		if !(height >= 70386 && height < 70411) {
			return fmt.Errorf("missing ecblock %d", height)
		}
	}

	if ecBlock != nil {
		for _, entry := range ecBlock.GetEntries() {
			switch entry.ECID() {
			case constants.ECIDChainCommit:
				ent := entry.(*entryCreditBlock.CommitChain)
				bs.ECAddressMap[ent.ECPubKey.Fixed()] -= int64(ent.Credits)
				debugIfNeg(log, ent.ECPubKey.Fixed(), bs.ECAddressMap[ent.ECPubKey.Fixed()], false, height)
			case constants.ECIDEntryCommit:
				ent := entry.(*entryCreditBlock.CommitEntry)
				bs.ECAddressMap[ent.ECPubKey.Fixed()] -= int64(ent.Credits)
				debugIfNeg(log, ent.ECPubKey.Fixed(), bs.ECAddressMap[ent.ECPubKey.Fixed()], false, height)
			}
		}
	}

	if diagnostic {
		// I believe these hashes can only be compared to hashes made by this tool
		fctHash := state.GetMapHash(bs.FCTAddressMap)
		ecHash := state.GetMapHash(bs.ECAddressMap)

		log.WithFields(logrus.Fields{
			"height":        height,
			"fct_adr_count": len(bs.FCTAddressMap),
			"ec_adr_count":  len(bs.ECAddressMap),
			"fct_hash":      fctHash.String(),
			"ec_hash":       ecHash.String(),
		}).Info("balance info")
	}

	return nil
}

func debugIfNeg(log *logrus.Logger, addr [32]byte, amt int64, fct bool, height uint32) {
	// There are negative balances under height 97886
	if amt < 0 && height > 97886 {
		str := primitives.ConvertFctAddressToUserStr(factoid.NewAddress(addr[:]))
		if !fct {
			str = primitives.ConvertECAddressToUserStr(factoid.NewAddress(addr[:]))
		}
		log.WithFields(logrus.Fields{
			"address": str,
			"amt":     amt,
			"height":  height,
		}).Info("balance under 0")
	}
}
