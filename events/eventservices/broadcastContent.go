package eventservices

import (
	"errors"
	"fmt"
	"strings"
)

type BroadcastContent int

const (
	BroadcastNever  BroadcastContent = 0
	BroadcastOnce   BroadcastContent = 1
	BroadcastAlways BroadcastContent = 2
)

func Parse(value string) (BroadcastContent, error) {
	switch strings.ToLower(value) {
	case "always":
		return BroadcastAlways, nil
	case "once":
		return BroadcastOnce, nil
	case "never":
		return BroadcastNever, nil
	}
	return -1, errors.New(fmt.Sprintf("could not parse %s to EventBroadcastContent", value))
}

func (broadcastContent BroadcastContent) String() string {
	switch broadcastContent {
	case BroadcastAlways:
		return "always"
	case BroadcastOnce:
		return "once"
	case BroadcastNever:
		return "never"
	default:
		return fmt.Sprintf("unknown value %d", int(broadcastContent))
	}
}
