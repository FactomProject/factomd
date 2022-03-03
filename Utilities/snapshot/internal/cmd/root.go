package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/FactomProject/factomd/Utilities/snapshot/internal/snapshot"
	"github.com/FactomProject/factomd/Utilities/tools"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func RootCmd() *cobra.Command {
	var (
		dbType        string
		dbPath        string
		logLvl        string
		stopHeight    int64
		dumpDirectory string
		debugHeights  uint32ArrayFlags
	)

	cmd := &cobra.Command{
		Use: "snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			log := logrus.New()
			lvl, err := logrus.ParseLevel(logLvl)
			if err != nil {
				return fmt.Errorf("parse log level %s: %w", logLvl, err)
			}
			log.SetLevel(lvl)

			dbPath = os.ExpandEnv(dbPath)
			db := tools.NewDBReader(dbType, dbPath)
			s := snapshot.New(snapshot.Config{
				Log:          log,
				DB:           db,
				DebugHeights: debugHeights,
				Stop:         stopHeight,
				DumpDir:      dumpDirectory,
			})

			err = s.WalkDB()
			if err != nil {
				return fmt.Errorf("snapshot: %w", err)
			}

			err = s.Dump()
			if err != nil {
				return fmt.Errorf("write data: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().Int64VarP(&stopHeight, "stop-height", "s", -1, "height to stop the snapshot at")
	cmd.Flags().StringVar(&logLvl, "debug", "debug", "set the log level")
	cmd.Flags().StringVar(&dbType, "db-type", "level", "optionally change the type to 'bolt'")
	cmd.Flags().StringVar(&dbPath, "db", "$HOME/.factom/m2/main-database/ldb/MAIN/factoid_level.db",
		"the location of the database to use")
	cmd.Flags().Var(&debugHeights, "debug-heights", "heights to print diagnostic information at")
	cmd.Flags().StringVarP(&dumpDirectory, "dump-dir", "d", "snapshot", "where to dump snapshot data. empty means do not dump")

	return cmd
}

type uint32ArrayFlags []uint32

func (i *uint32ArrayFlags) Type() string {
	return "uint32Array"
}

func (i *uint32ArrayFlags) String() string {
	return fmt.Sprintf("%#v", i)
}

func (i *uint32ArrayFlags) Set(value string) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	*i = append(*i, uint32(v))
	return nil
}
