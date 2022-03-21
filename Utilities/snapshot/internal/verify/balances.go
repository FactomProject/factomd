package verify

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/Utilities/snapshot/pkg/balances"

	"github.com/FactomProject/factom"
	"github.com/FactomProject/factomd/Utilities/snapshot/internal/verify/api"
	"github.com/sirupsen/logrus"
)

func VerifyBalancesAgainstFactomPro(ctx context.Context, log *logrus.Logger, factomAddr string, balancesFile io.Reader) error {
	fetcher, err := api.NewFactomPro(ctx)
	if err != nil {
		return err
	}
	return VerifyBalances(ctx, log, fetcher, balancesFile)
}

func VerifyBalancesAgainstFactomd(ctx context.Context, log *logrus.Logger, factomAddr string, balancesFile io.Reader) error {
	fetcher := api.NewFactomFetcher(factomAddr, "")
	return VerifyBalances(ctx, log, fetcher, balancesFile)
}

func VerifyBalances(ctx context.Context, log *logrus.Logger, fetcher api.APIFetcher, balancesFile io.Reader) error {
	bals, err := balances.LoadBalances(balancesFile)
	if err != nil {
		return fmt.Errorf("load balances: %w", err)
	}

	fetcherHeight, err := fetcher.Height(ctx)
	if err != nil {
		return fmt.Errorf("fetch height: %w", err)
	}

	log.WithFields(logrus.Fields{
		"height":     bals.Height,
		"api_height": fetcherHeight,
		"fa_total":   len(bals.FCTAddressMap),
		"ec_total":   len(bals.ECAddressMap),
	}).Info("Verifying balances...")
	badEC, badFA := 0, 0

	var faDone, ecDone int
	last := time.Now()
	// debug is to print progress
	debug := func() {
		if time.Since(last) > time.Second*5 {
			log.WithFields(logrus.Fields{
				"fa_done":  faDone,
				"fa_total": len(bals.FCTAddressMap),
				"fa_bad":   badFA,

				"ec_done":  ecDone,
				"ec_total": len(bals.ECAddressMap),
				"ec_bad":   badEC,
			}).Debug()
			last = time.Now()
		}
	}

	for addr, exp := range bals.FCTAddressMap {
		human := primitives.ConvertFctAddressToUserStr(primitives.NewHash(addr[:]))

		bal, err := fetcher.FactoidBalance(ctx, human)
		if err != nil {
			return fmt.Errorf("fetch balance %s: %w", addr, err)
		}
		if !compare(log, human, exp, bal, " FCT") {
			badFA++
		}
		faDone++
		debug()
	}

	for addr, exp := range bals.ECAddressMap {
		human := primitives.ConvertECAddressToUserStr(primitives.NewHash(addr[:]))

		bal, err := fetcher.EntryCreditBalance(ctx, human)
		if err != nil {
			return fmt.Errorf("fetch balance %s: %w", addr, err)
		}
		if !compare(log, human, exp, bal, " EC") {
			badEC++
		}
		ecDone++
		debug()
	}

	fields := logrus.Fields{
		"fct_mismatches":  badFA,
		"fct_done":        faDone,
		"ec_done":         ecDone,
		"ec_mismatches":   badEC,
		"snapshot_height": bals.Height,
		"api_height":      fetcherHeight,
	}
	if badFA+badEC == 0 {
		log.WithFields(fields).Info("All balances ok")
	} else {
		log.WithFields(fields).Error("found balance mismatches")
	}

	return nil
}

func compare(log *logrus.Logger, addr string, snapshot, api int64, balSuffix string) bool {
	if snapshot != api {
		d := snapshot - api
		p := ""
		if d < 0 {
			d = d * -1
			p = "-"
		}
		delta := p + factom.FactoshiToFactoid(uint64(d))

		if snapshot < 0 || api < 0 {
			log.WithFields(logrus.Fields{
				"addr":     addr,
				"snapshot": snapshot,
				"api":      api,
				"delta":    delta,
			}).Error("below 0 and balance mismatch")
		} else {
			log.WithFields(logrus.Fields{
				"addr":     addr,
				"snapshot": factom.FactoshiToFactoid(uint64(snapshot)) + balSuffix,
				"api":      factom.FactoshiToFactoid(uint64(api)) + balSuffix,
				"delta":    delta + balSuffix,
			}).Error("balance mismatch")
		}
		return false
	}
	return true
}
