package state_test

import (
	"testing"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/state"
)

func TestSortedMsgCache(t *testing.T) {
	m := state.NewMsgCache()
	constants.NeedsAck = func(t byte) bool {
		return false
	}

	var _ = m
}
