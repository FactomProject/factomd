package longtest

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"
	"time"

	"github.com/FactomProject/factomd/common"
	"github.com/FactomProject/factomd/modules/pubsub"

	"github.com/FactomProject/factomd/modules/registry"
)

// just load up and print out visualization of registered threads
func TestRegistryViz(t *testing.T) {
	homeDir := simulation.GetLongTestHome(t)
	simulation.ResetTestHome(homeDir, t)

	params := map[string]string{
		"--db":         "LDB",
		"--net":        "alot+",
		"--factomhome": homeDir,
	}
	state0 := simulation.SetupSim("LFFF", params, 10, 0, 0, t)
	simulation.WaitBlocks(state0, 2)
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
		"--factomhome":   simulation.GetLongTestHome(t),
	}
	state0 := simulation.StartSim(1, params)
	simulation.StatusEveryMinute(state0)
	t.Log("Graph of Thread Dependencies:")
	t.Log(registry.Graph())

	time.Sleep(500 * time.Minute)
}
