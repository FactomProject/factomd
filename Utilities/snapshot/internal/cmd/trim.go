package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"

	"github.com/FactomProject/factomd/Utilities/snapshot/pkg/balances"

	"github.com/FactomProject/factomd/Utilities/snapshot/internal"
	"github.com/spf13/cobra"
)

func trimCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trim",
		Short: "Trim balance output to only show non-zero factoid addresses",
		RunE: func(cmd *cobra.Command, args []string) error {
			dumpDirectory, err := cmd.Flags().GetString("dump-dir")
			if err != nil {
				return err
			}

			log, err := logger(cmd)
			if err != nil {
				return err
			}
			var _ = log

			fp := filepath.Join(dumpDirectory, internal.DefaultBalanceFile)
			file, err := os.OpenFile(fp, os.O_RDONLY, 0777)
			if err != nil {
				return fmt.Errorf("open file %s: %w", fp, err)
			}
			defer file.Close()

			bal, err := balances.LoadBalances(file)
			if err != nil {
				return fmt.Errorf("load balances: %w", err)
			}

			for k, v := range bal.FCTAddressMap {
				if v == 0 {
					delete(bal.FCTAddressMap, k)
				}
			}
			log.WithFields(logrus.Fields{
				"fct_count": len(bal.FCTAddressMap),
			}).Info("summary")
			// Remove all ex addresses
			bal.ECAddressMap = make(map[[32]byte]int64, 0)

			np := filepath.Join(dumpDirectory, "trimmed-"+internal.DefaultBalanceFile)
			newF, err := os.OpenFile(np, os.O_CREATE|os.O_WRONLY, 0777)
			if err != nil {
				return fmt.Errorf("open new file %s: %w", np, err)
			}

			err = bal.Dump(newF)
			if err != nil {
				return fmt.Errorf("dump trimmed: %w", err)
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringP("dump-dir", "d", internal.DefaultSnapshotDir, "where snapshot data lives")
	return cmd
}
