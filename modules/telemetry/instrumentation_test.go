package telemetry_test

import (
	"github.com/FactomProject/factomd/testHelper/simulation"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimulation(t *testing.T) {
	// Just load simulator
	assert.NotPanics(t, func() {
		simulation.SetupSim("LFF", map[string]string{}, 10, 0, 0, t)
	})
}
