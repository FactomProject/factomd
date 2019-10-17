package eventservices

import (
	"fmt"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/util"
)

type EventServiceParams struct {
	EnableLiveFeedAPI                bool
	Protocol                         string
	Address                          string
	OutputFormat                     EventFormat
	MuteEventReplayDuringStartup     bool
	ResendRegistrationsOnStateChange bool
	BroadcastContent                 BroadcastContent
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
	if len(factomParams.EventReceiverAddress) > 0 && factomParams.EventReceiverPort > 0 {
		params.Address = factomParams.EventReceiverAddress
	} else if len(config.LiveFeedAPI.EventReceiverAddress) > 0 && config.LiveFeedAPI.EventReceiverPort > 0 {
		params.Address = fmt.Sprintf("%s:%d", config.LiveFeedAPI.EventReceiverAddress, config.LiveFeedAPI.EventReceiverPort)
	} else {
		params.Address = fmt.Sprintf("%s:%d", defaultConnectionHost, defaultConnectionPort)
	}
	if len(factomParams.OutputFormat) > 0 {
		params.OutputFormat = EventFormatFrom(factomParams.OutputFormat, defaultOutputFormat)
	} else if len(config.LiveFeedAPI.OutputFormat) > 0 {
		params.OutputFormat = EventFormatFrom(config.LiveFeedAPI.OutputFormat, defaultOutputFormat)
	} else {
		params.OutputFormat = defaultOutputFormat
	}

	params.EnableLiveFeedAPI = factomParams.EnableLiveFeedAPI || config.LiveFeedAPI.EnableLiveFeedAPI
	params.MuteEventReplayDuringStartup = factomParams.MuteReplayDuringStartup || config.LiveFeedAPI.MuteReplayDuringStartup
	params.ResendRegistrationsOnStateChange = factomParams.ResendRegistrationsOnStateChange || config.LiveFeedAPI.ResendRegistrationsOnStateChange
	var err error
	if len(factomParams.BroadcastContent) > 0 {
		params.BroadcastContent, err = Parse(factomParams.BroadcastContent)
		if err != nil {
			log.Printf("Parameter BroadcastContent could not be parsed: %v", err)
			params.BroadcastContent = BroadcastOnRegistration
		}
	} else if len(config.LiveFeedAPI.BroadcastContent) > 0 {
		params.BroadcastContent, err = Parse(config.LiveFeedAPI.BroadcastContent)
		if err != nil {
			log.Printf("Configuration property LiveFeedAPI.BroadcastContent could not be parsed: %v", err)
			params.BroadcastContent = BroadcastOnRegistration
		}
	} else {
		params.BroadcastContent = BroadcastOnRegistration
	}
	return params
}
