package activations

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	// override net name
	netNameOnce.Do(func() {})
	netName = "MAIN"
}

func TestDefaults(t *testing.T) {
	netName = "MAIN"

	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 0))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 1))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 25))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 45335))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 222874-1))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 222874))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 222874+1))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, math.MaxInt32-1))

	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 0))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 1))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 25))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 45335))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 222874-1))
	assert.True(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 222874))
	assert.True(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 222874+1))
	assert.True(t, IsActive(AUTHRORITY_SET_MAX_DELTA, math.MaxInt32-1))

	netName = "LOCAL"

	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 0))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 1))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 24-1))
	assert.True(t, IsActive(TESTNET_COINBASE_PERIOD, 25))
	assert.True(t, IsActive(TESTNET_COINBASE_PERIOD, 25+1))
	assert.True(t, IsActive(TESTNET_COINBASE_PERIOD, math.MaxInt32-1))

	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 0))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 1))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 25-1))
	assert.True(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 25))
	assert.True(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 25+1))
	assert.True(t, IsActive(AUTHRORITY_SET_MAX_DELTA, math.MaxInt32-1))

	netName = "CUSTOM:fct_community_test"

	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 0))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 1))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 45335-1))
	assert.True(t, IsActive(TESTNET_COINBASE_PERIOD, 45335))
	assert.True(t, IsActive(TESTNET_COINBASE_PERIOD, 45335+1))
	assert.True(t, IsActive(TESTNET_COINBASE_PERIOD, math.MaxInt32-1))

	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 0))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 1))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 109387-1))
	assert.True(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 109387))
	assert.True(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 109387+1))
	assert.True(t, IsActive(AUTHRORITY_SET_MAX_DELTA, math.MaxInt32-1))

	netName = "XXXXX:default"

	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 0))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 1))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 25))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 45335))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, 222874))
	assert.False(t, IsActive(TESTNET_COINBASE_PERIOD, math.MaxInt32-1))

	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 0))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 1))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 25))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 45335))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, 222874))
	assert.False(t, IsActive(AUTHRORITY_SET_MAX_DELTA, math.MaxInt32-1))
}
