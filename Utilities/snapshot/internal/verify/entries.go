package verify

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/FactomProject/factomd/Utilities/snapshot/load"

	"github.com/FactomProject/factomd/Utilities/snapshot/internal/verify/api"
	"github.com/sirupsen/logrus"
)

func VerifyChainAgainstFactomPro(ctx context.Context, log *logrus.Logger, factomAddr string, chainID string, chainFile io.Reader) error {
	fetcher, err := api.NewFactomPro(ctx)
	if err != nil {
		return err
	}
	return VerifyChain(ctx, log, fetcher, chainID, chainFile)
}

func VerifyChainAgainstFactomd(ctx context.Context, log *logrus.Logger, factomAddr string, chainID string, chainFile io.Reader) error {
	fetcher := api.NewFactomFetcher(factomAddr, "")
	return VerifyChain(ctx, log, fetcher, chainID, chainFile)
}

func VerifyChain(ctx context.Context, logger *logrus.Logger, fetcher api.APIFetcher, chainID string, chainFile io.Reader) error {
	fLog := logger.WithFields(logrus.Fields{"chain": chainID})

	chain, err := load.LoadChain(chainFile)
	if err != nil {
		return fmt.Errorf("load chain: %w", err)
	}

	head, err := fetcher.ChainHead(ctx, chainID)
	if err != nil {
		return fmt.Errorf("get chainhead: %w", err)
	}

	if chain.Head.KeyMR != head {
		fLog.WithFields(logrus.Fields{
			"api_head":     head,
			"snaphot_head": chain.Head.KeyMR,
		}).Errorf("chainhead incorrect")
	}

	for i := len(chain.Eblocks) - 1; i >= 0; i-- {
		sEblock := chain.Eblocks[i]

		ents := chain.EblockEntries(sEblock)

		aEblock, err := fetcher.Eblock(ctx, sEblock.KeyMR)
		if err != nil {
			return fmt.Errorf("fetch eblock %s: %w", sEblock.KeyMR, err)
		}

		fLog = fLog.WithFields(logrus.Fields{
			"key_mr": sEblock.KeyMR,

			"snapshot_ents": len(ents),
			"snapshot_seq":  sEblock.Seq,

			"api_seq":  aEblock.Header.BlockSequenceNumber,
			"api_ents": len(aEblock.EntryList),
		})

		mismatch := []string{}
		if aEblock.Header.BlockSequenceNumber != sEblock.Seq {
			mismatch = append(mismatch, "seq numbers")
		}

		if len(aEblock.EntryList) != len(ents) {
			mismatch = append(mismatch, "entry count")
		}

		if len(mismatch) > 0 {
			fLog.Errorf("mismatching: %s", strings.Join(mismatch, ","))
		}

		if len(ents) == len(aEblock.EntryList) {
			for i := range ents {
				snapshotEnt := ents[i]
				apiEnt := aEblock.EntryList[i]
				if apiEnt.EntryHash != snapshotEnt.Hash().String() {
					fLog.WithFields(logrus.Fields{
						"eblock":           sEblock.KeyMR,
						"eblock_seq":       sEblock.Seq,
						"entry_eblock_idx": i,
						"snapshot_hash":    snapshotEnt.Hash().String(),
						"api_hash":         apiEnt.EntryHash,
					}).Errorf("entry hash mismatch")
					fLog.Error("hash mismatch")
				}
			}
		}
	}

	return nil
}
