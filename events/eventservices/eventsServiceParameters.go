package eventservices

import (
	"fmt"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/events/eventoutputformat"
	"github.com/FactomProject/factomd/util"
)

type EventServiceParams struct {
	EnableLiveFeedAPI       bool
	Protocol                string
	Address                 string
	OutputFormat            eventoutputformat.Format
	MuteEventsDuringStartup bool
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
	if len(factomParams.EventFormat) > 0 {
		params.OutputFormat = eventoutputformat.FormatFrom(factomParams.EventFormat, defaultOutputFormat)
	} else if len(config.LiveFeedAPI.EventFormat) > 0 {
		params.OutputFormat = eventoutputformat.FormatFrom(config.LiveFeedAPI.EventFormat, defaultOutputFormat)
	} else {
		params.OutputFormat = defaultOutputFormat
	}

	params.EnableLiveFeedAPI = factomParams.EnableLiveFeedAPI || config.LiveFeedAPI.EnableLiveFeedAPI
	params.MuteEventsDuringStartup = factomParams.MuteEventsDuringStartup || config.LiveFeedAPI.MuteEventsDuringStartup

	return params
}
