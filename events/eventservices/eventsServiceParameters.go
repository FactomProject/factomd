package eventservices

import (
	"fmt"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/events/eventoutputformat"
	"github.com/FactomProject/factomd/util"
)

func selectParameters(params *globals.FactomParams, config *util.FactomdConfig) (string, string, eventoutputformat.Format) {
	var protocol string
	if len(params.EventReceiverProtocol) > 0 {
		protocol = params.EventReceiverProtocol
	} else if len(config.LiveFeedAPI.EventReceiverProtocol) > 0 {
		protocol = config.LiveFeedAPI.EventReceiverProtocol
	} else {
		protocol = defaultProtocol
	}
	var address string
	if len(params.EventReceiverAddress) > 0 && params.EventReceiverPort > 0 {
		address = params.EventReceiverAddress
	} else if len(config.LiveFeedAPI.EventReceiverAddress) > 0 && config.LiveFeedAPI.EventReceiverPort > 0 {
		address = fmt.Sprintf("%s:%d", config.LiveFeedAPI.EventReceiverAddress, config.LiveFeedAPI.EventReceiverPort)
	} else {
		address = fmt.Sprintf("%s:%d", defaultConnectionHost, defaultConnectionPort)
	}
	var outputFormat eventoutputformat.Format
	if len(params.EventFormat) > 0 {
		outputFormat = eventoutputformat.FormatFrom(params.EventFormat, defaultOutputFormat)
	} else if len(config.LiveFeedAPI.EventFormat) > 0 {
		outputFormat = eventoutputformat.FormatFrom(config.LiveFeedAPI.EventFormat, defaultOutputFormat)
	} else {
		outputFormat = defaultOutputFormat
	}
	return protocol, address, outputFormat
}
