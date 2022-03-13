package cmd

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/Utilities/snapshot/internal"

	survey "github.com/AlecAivazis/survey/v2"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func cleanCmd() *cobra.Command {
	var (
		dumpDirectory string
		skipYes       bool
	)

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Deletes the snapshotted data.",
		RunE: func(cmd *cobra.Command, args []string) error {
			log, err := logger(cmd)
			if err != nil {
				return err
			}

			_, err = os.Stat(dumpDirectory)
			if os.IsNotExist(err) {
				return fmt.Errorf("nothing to clean, no directory '%s' does not exist", dumpDirectory)
			}

			if !skipYes {
				prompt := &survey.Confirm{
					Message: fmt.Sprintf("Are you sure you wish to delete the contents at '%s'?", dumpDirectory),
				}
				resp := false
				err = survey.AskOne(prompt, &resp)
				if err != nil {
					return fmt.Errorf("confirm prompt: %w", err)
				}
				if !resp {
					log.Info("aborting")
					return nil
				}
			}

			return clean(log, dumpDirectory)
		},
	}

	cmd.Flags().BoolVarP(&skipYes, "yes", "y", false, "skip yes prompt")
	cmd.Flags().StringVarP(&dumpDirectory, "dump-dir", "d", internal.DefaultSnapshotDir, "where to dump snapshot data. empty means do not dump")

	return cmd
}

func clean(log *logrus.Logger, dumpDirectory string) error {
	err := os.RemoveAll(dumpDirectory)
	if err != nil {
		log.WithFields(logrus.Fields{
			"error":     err.Error(),
			"directory": dumpDirectory,
		}).Error("delete directory")
		return fmt.Errorf("clean failed")
	}
	return nil
}
