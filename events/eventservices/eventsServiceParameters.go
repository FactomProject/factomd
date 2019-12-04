package eventservices

import (
	"fmt"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/util"
)

type EventServiceParams struct {
	EnableLiveFeedAPI     bool
	Protocol              string
	Address               string
	OutputFormat          EventFormat
	ReplayDuringStartup   bool
	SendStateChangeEvents bool
	BroadcastContent      BroadcastContent
}

func selectParameters(factomParams *globals.FactomParams, config *util.FactomdConfig) *EventServiceParams {
	params := new(EventServiceParams)
	if len(factomParams.EventReceiverProtocol) > 0 {
		params.Protocol = factomParams.EventReceiverProtocol
	} else if len(config.LiveFeedAPI.EventReceiverProtocol) > 0 {
		params.Protocol = config.LiveFeedAPI.EventReceiverProtocol
	} else {
		params.Protocol = defaultProtocol
	}
	if len(factomParams.EventReceiverHost) > 0 && factomParams.EventReceiverPort > 0 {
		params.Address = fmt.Sprintf("%s:%d", factomParams.EventReceiverHost, factomParams.EventReceiverPort)
	} else if len(config.LiveFeedAPI.EventReceiverHost) > 0 && config.LiveFeedAPI.EventReceiverPort > 0 {
		params.Address = fmt.Sprintf("%s:%d", config.LiveFeedAPI.EventReceiverHost, config.LiveFeedAPI.EventReceiverPort)
	} else {
		params.Address = fmt.Sprintf("%s:%d", defaultConnectionHost, defaultConnectionPort)
	}
	if len(factomParams.EventFormat) > 0 {
		params.OutputFormat = EventFormatFrom(factomParams.EventFormat, defaultOutputFormat)
	} else if len(config.LiveFeedAPI.EventFormat) > 0 {
		params.OutputFormat = EventFormatFrom(config.LiveFeedAPI.EventFormat, defaultOutputFormat)
	} else {
		params.OutputFormat = defaultOutputFormat
	}

	params.EnableLiveFeedAPI = factomParams.EnableLiveFeedAPI || config.LiveFeedAPI.EnableLiveFeedAPI
	params.ReplayDuringStartup = factomParams.EventReplayDuringStartup || config.LiveFeedAPI.EventReplayDuringStartup
	params.SendStateChangeEvents = factomParams.EventSendStateChange || config.LiveFeedAPI.EventSendStateChange
	var err error
	if len(factomParams.EventBroadcastContent) > 0 {
		params.BroadcastContent, err = Parse(factomParams.EventBroadcastContent)
		if err != nil {
			//log.Printfln("Parameter EventBroadcastContent could not be parsed: %v\n", err)
			params.BroadcastContent = BroadcastOnce
		}
	} else if len(config.LiveFeedAPI.EventBroadcastContent) > 0 {
		params.BroadcastContent, err = Parse(config.LiveFeedAPI.EventBroadcastContent)
		if err != nil {
			//log.Printfln("Configuration property LiveFeedAPI.EventBroadcastContent could not be parsed: %v", err)
			params.BroadcastContent = BroadcastOnce
		}
	} else {
		params.BroadcastContent = BroadcastOnce
	}
	return params
}
