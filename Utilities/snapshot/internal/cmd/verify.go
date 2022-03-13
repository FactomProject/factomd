package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/FactomProject/factomd/Utilities/snapshot/internal"

	"github.com/FactomProject/factomd/Utilities/snapshot/internal/verify"
	"github.com/spf13/cobra"
)

func verifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify",
		Short: "verifies the snapshotted data against factom",
		Long:  "Verifying the data will compare the snapshot data to a mainnet factom node.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.PersistentFlags().StringP("dump-dir", "d", internal.DefaultSnapshotDir, "where to dump snapshot data. empty means do not dump")
	cmd.AddCommand(verifyBalances())
	return cmd
}

func verifyBalances() *cobra.Command {
	var (
		apiAddr string
		apiType string
	)

	cmd := &cobra.Command{
		Use:   "balances",
		Short: "verify found balances",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			dumpDirectory, err := cmd.Flags().GetString("dump-dir")
			if err != nil {
				return err
			}

			log, err := logger(cmd)
			if err != nil {
				return err
			}

			fp := filepath.Join(dumpDirectory, internal.DefaultBalanceFile)
			file, err := os.OpenFile(fp, os.O_RDONLY, 0777)
			if err != nil {
				return fmt.Errorf("open file %s: %w", fp, err)
			}
			defer file.Close()
			switch apiType {
			case "factomd":
				err := verify.VerifyBalancesAgainstFactomd(ctx, log, apiAddr, file)
				if err != nil {
					return fmt.Errorf("verify: %w", err)
				}
			case "factom.pro":
				err := verify.VerifyBalancesAgainstFactomPro(ctx, log, apiAddr, file)
				if err != nil {
					return fmt.Errorf("verify: %w", err)
				}
			default:
				return fmt.Errorf("api type %s is not supported", apiType)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&apiAddr, "api-addr", "s", "localhost:8088", "address of the api server to verify against")
	cmd.Flags().StringVar(&apiType, "api-type", "factomd", "supports 'factomd', and 'factom.pro'")

	return cmd
}
