package longtest

import (
	"github.com/FactomProject/factomd/registry"
	. "github.com/FactomProject/factomd/testHelper"
	"testing"
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
	state0 := SetupSim("L", params, 10, 0, 0, t)
	_ = state0
	t.Log("Graph of Thread Dependencies:")
	t.Log(registry.Graph())
}

