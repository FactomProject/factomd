package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/FactomProject/factomd/Utilities/snapshot/internal/snapshot"
	"github.com/FactomProject/factomd/Utilities/tools"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func logger(cmd *cobra.Command) (*logrus.Logger, error) {
	logLvl, _ := cmd.Flags().GetString("log")
	log := logrus.New()
	lvl, err := logrus.ParseLevel(logLvl)
	if err != nil {
		return nil, fmt.Errorf("parse log level %s: %w", logLvl, err)
	}
	log.SetLevel(lvl)

	return log, nil
}

func RootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(snapshotCmd())
	cmd.AddCommand(cleanCmd())
	cmd.PersistentFlags().String("log", "debug", "set the log level")

	return cmd
}

func snapshotCmd() *cobra.Command {
	var (
		dbType        string
		dbPath        string
		stopHeight    int64
		dumpDirectory string
		debugHeights  uint32ArrayFlags
		recordEntries bool
	)

	cmd := &cobra.Command{
		Use:   "new",
		Short: "Take a new snapshot of a factom database",
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()
			log, err := logger(cmd)
			if err != nil {
				return err
			}

			defer func() {
				log.WithFields(logrus.Fields{
					"duration": time.Since(start),
				}).Info("complete")
			}()
			stat, err := os.Stat(dumpDirectory)
			if err == nil {
				if !stat.IsDir() {
					return fmt.Errorf("the file '%s' exists where the data is to be saved; you must delete this file", stat.Name())
				}
				if stat.IsDir() {
					return fmt.Errorf("the directory '%s' already exists where the data is to be saved; run 'snapshot clean'", stat.Name())
				}
			}

			dbPath = os.ExpandEnv(dbPath)
			db := tools.NewDBReader(dbType, dbPath)
			s, err := snapshot.New(snapshot.Config{
				Log:           log,
				DB:            db,
				DebugHeights:  debugHeights,
				Stop:          stopHeight,
				DumpDir:       dumpDirectory,
				RecordEntries: recordEntries,
			})
			if err != nil {
				return fmt.Errorf("new snapshotter: %w", err)
			}

			err = s.WalkDB()
			if err != nil {
				return fmt.Errorf("snapshot: %w", err)
			}

			err = s.Done()
			if err != nil {
				return fmt.Errorf("write data: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&recordEntries, "record-entries", "e", false, "enable snapshotting entry data")
	cmd.Flags().Int64VarP(&stopHeight, "stop-height", "s", -1, "height to stop the snapshot at")
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
