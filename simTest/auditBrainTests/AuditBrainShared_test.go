package auditBrainTests_test

import (
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/engine"
	"github.com/FactomProject/factomd/state"
	. "github.com/FactomProject/factomd/testHelper"
	"os"
	"testing"
)

func SetupConfigFiles(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	globals.Params.FactomHome = dir + "/TestBrainSwap"
	os.Setenv("FACTOM_HOME", globals.Params.FactomHome)

	t.Logf("Removing old run in %s", globals.Params.FactomHome)
	if err := os.RemoveAll(globals.Params.FactomHome); err != nil {
		t.Fatal(err)
	}

	// build config files for the test
	for i := 0; i < 6; i++ {
		WriteConfigFile(i, i, "", t) // just write the minimal config
	}
}

func SetupNodes(t *testing.T, givenNodes string) map[int]*state.State {
	states := map[int]*state.State{}
	states[0] = SetupSim(givenNodes, buildParmList(), 15, 0, 0, t)
	for i := 1; i <= len(givenNodes)-1; i++ {
		states[i] = engine.GetFnodes()[i].State
	}
	return states
}

func buildParmList() map[string]string {
	params := map[string]string{
		"--db":                  "LDB", // NOTE: using MAP causes an occasional error see FD-825
		"--network":             "LOCAL",
		"--net":                 "alot+",
		"--enablenet":           "true",
		"--blktime":             "10",
		"--startdelay":          "1",
		"--stdoutlog":           "out.txt",
		"--stderrlog":           "out.txt",
		"--checkheads":          "false",
		"--controlpanelsetting": "readwrite",
		"--debuglog":            ".",
		"--logPort":             "38000",
		"--port":                "38001",
		"--controlpanelport":    "38002",
		"--networkport":         "38003",
		"--peers":               "127.0.0.1:37003",
		"--factomhome":          globals.Params.FactomHome,
	}
	return params
}
