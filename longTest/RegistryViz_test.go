package longtest

import (
	"testing"
	"time"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/pubsub"

	"github.com/FactomProject/factomd/registry"
	. "github.com/FactomProject/factomd/testHelper"
)

// just load up and print out visualization of registered threads
func TestRegistryViz(t *testing.T) {
	homeDir := GetLongTestHome(t)
	ResetTestHome(homeDir, t)

	params := map[string]string{
		"--db":         "LDB",
		"--net":        "alot+",
		"--factomhome": homeDir,
	}
	state0 := SetupSim("LFFF", params, 10, 0, 0, t)
	WaitBlocks(state0, 2)
	_ = state0

	// echo thread/pubsub/named obj hierarchies
	t.Log(registry.Graph())
	t.Log(pubsub.GlobalRegistry().PrintTree())
	common.PrintAllNames()

}

func TestRegistryVizExistingDB(t *testing.T) {
	params := map[string]string{
		"--db":           "LDB",
		"--fastsaverate": "100",
		"--net":          "alot+",
		"--factomhome":   GetLongTestHome(t),
	}
	state0 := StartSim(1, params)
	StatusEveryMinute(state0)
	t.Log("Graph of Thread Dependencies:")
	t.Log(registry.Graph())

	time.Sleep(500 * time.Minute)
}
