package testHelper_test

import (
	"testing"

	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/testHelper"
)

func TestCounterDump(t *testing.T) {
	// Choose a random prometheus counter
	if f := testHelper.DumpPrometheusCounter(state.TotalHoldingQueueInputs); f != 0 {
		t.Errorf("Should find 0, found %f", f)
	}

	state.TotalHoldingQueueInputs.Inc()
	if f := testHelper.DumpPrometheusCounter(state.TotalHoldingQueueInputs); f != 1 {
		t.Errorf("Should find 1, found %f", f)
	}

	state.TotalHoldingQueueInputs.Add(100.1)
	if f := testHelper.DumpPrometheusCounter(state.TotalHoldingQueueInputs); f != 101.1 {
		t.Errorf("Should find 1, found %f", f)
	}

	next := 101.1
	for i := 0; i < 10; i++ {
		state.TotalHoldingQueueInputs.Inc()
		next += 1
		if f := testHelper.DumpPrometheusCounter(state.TotalHoldingQueueInputs); f != next {
			t.Errorf("Should find %f, found %f", next, f)
		}
	}
}
