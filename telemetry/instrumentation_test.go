package telemetry_test

import (
	"github.com/FactomProject/factomd/testHelper"
	"testing"
)

func TestSimulation(t *testing.T) {
	// Just load simulator
	testHelper.SetupSim("LFF", map[string]string{}, 10, 0, 0, t)
}
