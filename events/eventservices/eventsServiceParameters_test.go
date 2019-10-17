package eventservices

import (
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventServiceParameters(t *testing.T) {
	t.Run("Test parameter overrides", func(t *testing.T) {
		config := buildBaseConfig()
		params := buildOverrideParams()
		testParams := selectParameters(params, config)
		assert.True(t, testParams.EnableLiveFeedAPI)
		assert.Equal(t, "udp", testParams.Protocol)
		assert.Equal(t, "0.0.0.0", testParams.Address)
		assert.Equal(t, "Json", testParams.OutputFormat.String())
		assert.True(t, testParams.ReplayDuringStartup)
		assert.True(t, testParams.SendStateChangeEvents)
		assert.Equal(t, "always", testParams.BroadcastContent.String())
	})
}

func buildOverrideParams() *globals.FactomParams {
	return &globals.FactomParams{
		EnableLiveFeedAPI:        true,
		EventReceiverProtocol:    "udp",
		EventReceiverAddress:     "0.0.0.0",
		EventReceiverPort:        8888,
		EventFormat:              "json",
		EventReplayDuringStartup: true,
		EventSendStateChange:     true,
		EventBroadcastContent:    "always",
	}
}

func buildBaseConfig() *util.FactomdConfig {
	return &util.FactomdConfig{
		LiveFeedAPI: struct {
			EnableLiveFeedAPI        bool
			EventReceiverProtocol    string
			EventReceiverAddress     string
			EventReceiverPort        int
			EventFormat              string
			EventReplayDuringStartup bool
			EventSendStateChange     bool
			EventBroadcastContent    string
		}{
			EnableLiveFeedAPI:        false,
			EventReceiverProtocol:    "tcp",
			EventReceiverAddress:     "127.0.0.1",
			EventReceiverPort:        8444,
			EventFormat:              "protobuf",
			EventReplayDuringStartup: false,
			EventSendStateChange:     false,
			EventBroadcastContent:    "never",
		},
	}
}
