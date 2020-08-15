package eventservices

import (
	"fmt"
	"testing"

	"github.com/PaulSnow/factom2d/common/globals"
	"github.com/PaulSnow/factom2d/events/eventconfig"
	"github.com/PaulSnow/factom2d/util"
	"github.com/stretchr/testify/assert"
)

func TestEventServiceParameters_DefaultParameters(t *testing.T) {
	config := &util.FactomdConfig{}
	factomParams := &globals.Params

	params := selectParameters(factomParams, config)

	assert.Equal(t, defaultProtocol, params.Protocol)
	assert.Equal(t, fmt.Sprintf("%s:%d", defaultConnectionHost, defaultConnectionPort), params.Address)
	assert.Equal(t, defaultOutputFormat, params.OutputFormat)

	assert.False(t, params.EnableLiveFeedAPI)
	assert.False(t, params.SendStateChangeEvents)
	assert.False(t, params.ReplayDuringStartup)
	assert.Equal(t, eventconfig.BroadcastOnce, params.BroadcastContent)
}

func TestEventServiceParameters_OverrideParameters(t *testing.T) {
	config := buildBaseConfig(
		false,
		"tcp",
		"127.0.0.1",
		8444,
		"protobuf",
		false,
		false,
		"never",
	)
	factomParams := &globals.FactomParams{
		EnableLiveFeedAPI:        true,
		EventReceiverProtocol:    "udp",
		EventReceiverHost:        "0.0.0.0",
		EventReceiverPort:        8888,
		EventFormat:              "json",
		EventReplayDuringStartup: true,
		EventSendStateChange:     true,
		EventBroadcastContent:    "always",
	}

	testParams := selectParameters(factomParams, config)

	assert.True(t, testParams.EnableLiveFeedAPI)
	assert.Equal(t, "udp", testParams.Protocol)
	assert.Equal(t, "0.0.0.0:8888", testParams.Address)
	assert.Equal(t, eventconfig.Json, testParams.OutputFormat)
	assert.True(t, testParams.ReplayDuringStartup)
	assert.True(t, testParams.SendStateChangeEvents)
	assert.Equal(t, eventconfig.BroadcastAlways, testParams.BroadcastContent)
}

func TestEventServiceParameters_ConfigParameters(t *testing.T) {
	config := buildBaseConfig(
		true,
		"tcp",
		"127.0.0.1",
		8444,
		"protobuf",
		true,
		true,
		"never",
	)
	factomParams := &globals.Params

	testParams := selectParameters(factomParams, config)

	assert.True(t, testParams.EnableLiveFeedAPI)
	assert.Equal(t, "tcp", testParams.Protocol)
	assert.Equal(t, "127.0.0.1:8444", testParams.Address)
	assert.Equal(t, eventconfig.Protobuf, testParams.OutputFormat)
	assert.True(t, testParams.ReplayDuringStartup)
	assert.True(t, testParams.SendStateChangeEvents)
	assert.Equal(t, eventconfig.BroadcastNever, testParams.BroadcastContent)
}

func TestEventServiceParameters_ParseBroadcastError(t *testing.T) {
	config := &util.FactomdConfig{}
	factomParams := &globals.FactomParams{
		EnableLiveFeedAPI:        true,
		EventReceiverProtocol:    "udp",
		EventReceiverHost:        "0.0.0.0",
		EventReceiverPort:        8888,
		EventFormat:              "json",
		EventReplayDuringStartup: true,
		EventSendStateChange:     true,
		EventBroadcastContent:    "alwayss",
	}
	params := selectParameters(factomParams, config)
	assert.Equal(t, eventconfig.BroadcastOnce, params.BroadcastContent)
}

func TestEventServiceParameters_ParseBroadcastErrorOverride(t *testing.T) {
	config := buildBaseConfig(
		true,
		"tcp",
		"127.0.0.1",
		8444,
		"protobuf",
		true,
		true,
		"nevers",
	)
	factomParams := &globals.Params
	params := selectParameters(factomParams, config)
	assert.Equal(t, eventconfig.BroadcastOnce, params.BroadcastContent)
}

func buildBaseConfig(enable bool, protocol string, address string, port int, format string, replay bool, stateChange bool, broadcast string) *util.FactomdConfig {
	return &util.FactomdConfig{
		LiveFeedAPI: struct {
			EnableLiveFeedAPI        bool
			EventReceiverProtocol    string
			EventReceiverHost        string
			EventReceiverPort        int
			EventFormat              string
			EventReplayDuringStartup bool
			EventSendStateChange     bool
			EventBroadcastContent    string
		}{
			EnableLiveFeedAPI:        enable,
			EventReceiverProtocol:    protocol,
			EventReceiverHost:        address,
			EventReceiverPort:        port,
			EventFormat:              format,
			EventReplayDuringStartup: replay,
			EventSendStateChange:     stateChange,
			EventBroadcastContent:    broadcast,
		},
	}
}
