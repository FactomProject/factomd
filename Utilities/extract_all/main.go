package main

import (
	"fmt"
	"os"
	"os/user"
	"path"

	"time"

	"github.com/FactomProject/factomd/Utilities/extract_all/code"
	"github.com/FactomProject/factomd/common/constants/runstate"
	"github.com/FactomProject/factomd/engine"
)

func main() {

	// Make sure we have the output directory
	u, _ := user.Current()
	code.FullDir = path.Join(u.HomeDir + code.OutputDir)
	os.RemoveAll(code.FullDir)
	if err := os.MkdirAll(code.FullDir, os.ModePerm); err != nil {
		fmt.Println(err)
	}

	args := []string{"--enablenet=false"} // Enable or disable the network
	params := engine.ParseCmdLine(args)
	params.PrettyPrint()
	code.FactomdState = engine.Factomd(params)
	for code.FactomdState.GetRunState() != runstate.Running {
		time.Sleep(time.Second)
	}

	go func() {
		for code.FactomdState.GetRunState() != runstate.Stopped {
			time.Sleep(time.Second)
		}
		fmt.Println("Waiting to Shut Down") // This may not be necessary anymore with the new run state method
		time.Sleep(time.Second * 5)
	}()
	code.ProcessBalances()
	code.ProcessDictionaries()
}
