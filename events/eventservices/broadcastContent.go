package eventservices

import (
	"errors"
	"fmt"
	"strings"
)

type BroadcastContent int

const (
	BroadcastNever          BroadcastContent = 0
	BroadcastOnRegistration BroadcastContent = 1
	BroadcastAlways         BroadcastContent = 2
)

func Parse(value string) (BroadcastContent, error) {
	switch strings.ToLower(value) {
	case "always":
		return BroadcastAlways, nil
	case "onregistration":
		return BroadcastOnRegistration, nil
	case "never":
		return BroadcastNever, nil
	}
	return -1, errors.New(fmt.Sprintf("could not parse %s to BroadcastContent", value))
}

func (broadcastContent BroadcastContent) String() string {
	switch broadcastContent {
	case BroadcastAlways:
		return "Always"
	case BroadcastOnRegistration:
		return "OnRegistration"
	case BroadcastNever:
		return "Never"
	default:
		return fmt.Sprintf("unknown value %d", int(broadcastContent))
	}
}
