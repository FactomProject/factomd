package bmv

import (
	"github.com/FactomProject/factomd/telemetry"
)

var (
	TotalMessagesReceived = telemetry.NewCounterVec(
		"factomd_basicmessagefilter_msgs_receive_total",
		"Total messages accepted from the p2p", []string{"type"})
)
