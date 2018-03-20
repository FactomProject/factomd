// Copyright © 2018 Factom Inc. <clay@factom.com>

package cmd

import (
	"errors"
	"fmt"
	. "github.com/FactomProject/factomd/electionsCore/ET2/divefromfile"
	"github.com/spf13/cobra"
	"os"
)

var listen, connect, load string
var recursions, randomFactor, primeIdx, global int

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "divefromfilecli <statefile>",
	Short: "Test Factom Election from saved state",
	Long:  `Need some text here`,
	Args:  cobra.MinimumNArgs(1),
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			panic(errors.New("The inital state filename argument is required"))
		}
		name := args[0]
		// create a thing
		DiveFromFile(name, listen, connect, load, recursions, randomFactor, primeIdx, global)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	//rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.divefromfilecli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().StringVarP(&listen, "listen", "l", "", "Listen for mirrors on port")
	rootCmd.Flags().StringVarP(&connect, "connect", "c", "", "Connect to other dives to share mirrors")
	rootCmd.Flags().StringVarP(&load, "load", "m", "", "File name to load initial mirrors")
	rootCmd.Flags().IntVarP(&recursions, "recursionlimit", "r", 1000, "Number of recursions allowed")
	rootCmd.Flags().IntVarP(&randomFactor, "primeindex", "p", 0, "Pick a starting prime")
	rootCmd.Flags().IntVarP(&global, "reportfrequency", "g", 1000, "How many global nodes between prints")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

}
