package eventservices

import (
	"fmt"

	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/events/eventconfig"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/util"
)

type EventServiceParams struct {
	EnableLiveFeedAPI     bool
	Protocol              string
	Address               string
	ClientPort            string
	OutputFormat          eventconfig.EventFormat
	ReplayDuringStartup   bool
	SendStateChangeEvents bool
	BroadcastContent      eventconfig.BroadcastContent
}

func selectParameters(factomParams *globals.FactomParams, config *util.FactomdConfig) *EventServiceParams {
	params := new(EventServiceParams)
	if factomParams != nil && len(factomParams.EventReceiverProtocol) > 0 {
		params.Protocol = factomParams.EventReceiverProtocol
	} else if config != nil && len(config.LiveFeedAPI.EventReceiverProtocol) > 0 {
		params.Protocol = config.LiveFeedAPI.EventReceiverProtocol
	} else {
		params.Protocol = defaultProtocol
	}
	if factomParams != nil && len(factomParams.EventReceiverHost) > 0 && factomParams.EventReceiverPort > 0 {
		params.Address = fmt.Sprintf("%s:%d", factomParams.EventReceiverHost, factomParams.EventReceiverPort)
	} else if config != nil && len(config.LiveFeedAPI.EventReceiverHost) > 0 && config.LiveFeedAPI.EventReceiverPort > 0 {
		params.Address = fmt.Sprintf("%s:%d", config.LiveFeedAPI.EventReceiverHost, config.LiveFeedAPI.EventReceiverPort)
	} else {
		params.Address = fmt.Sprintf("%s:%d", defaultConnectionHost, defaultConnectionPort)
	}
	if factomParams != nil && factomParams.EventSenderPort > 0 {
		params.ClientPort = fmt.Sprintf(":%d", factomParams.EventSenderPort)
	} else if config != nil && config.LiveFeedAPI.EventSenderPort > 0 {
		params.ClientPort = fmt.Sprintf(":%d", config.LiveFeedAPI.EventSenderPort)
	}
	if factomParams != nil && len(factomParams.EventFormat) > 0 {
		params.OutputFormat = eventconfig.EventFormatFrom(factomParams.EventFormat, defaultOutputFormat)
	} else if config != nil && len(config.LiveFeedAPI.EventFormat) > 0 {
		params.OutputFormat = eventconfig.EventFormatFrom(config.LiveFeedAPI.EventFormat, defaultOutputFormat)
	} else {
		params.OutputFormat = defaultOutputFormat
	}

	params.EnableLiveFeedAPI = (factomParams != nil && factomParams.EnableLiveFeedAPI) || (config != nil && config.LiveFeedAPI.EnableLiveFeedAPI)
	params.ReplayDuringStartup = (factomParams != nil && factomParams.EventReplayDuringStartup) || (config != nil && config.LiveFeedAPI.EventReplayDuringStartup)
	params.SendStateChangeEvents = (factomParams != nil && factomParams.EventSendStateChange) || (config != nil && config.LiveFeedAPI.EventSendStateChange)

	var err error
	if factomParams != nil && len(factomParams.EventBroadcastContent) > 0 {
		params.BroadcastContent, err = eventconfig.ParseBroadcastContent(factomParams.EventBroadcastContent)
		if err != nil {
			log.LogPrintf("livefeed", "Parameter EventBroadcastContent could not be parsed: %v\n", err)
			params.BroadcastContent = eventconfig.BroadcastOnce
		}
	} else if config != nil && len(config.LiveFeedAPI.EventBroadcastContent) > 0 {
		params.BroadcastContent, err = eventconfig.ParseBroadcastContent(config.LiveFeedAPI.EventBroadcastContent)
		if err != nil {
			log.LogPrintf("livefeed", "Configuration property LiveFeedAPI.EventBroadcastContent could not be parsed: %v", err)
			params.BroadcastContent = eventconfig.BroadcastOnce
		}
	} else {
		params.BroadcastContent = eventconfig.BroadcastOnce
	}

	return params
}
