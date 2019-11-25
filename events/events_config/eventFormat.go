package events_config

import (
	"fmt"
	"strings"
)

type EventFormat int

const (
	Protobuf EventFormat = 1
	Json     EventFormat = 2
)

func EventFormatFrom(value string, defaultFormat EventFormat) EventFormat {
	switch strings.ToLower(value) {
	case strings.ToLower(Protobuf.String()):
		return Protobuf
	case strings.ToLower(Json.String()):
		return Json
	default:
		return defaultFormat
	}
}

func (outputFormat EventFormat) String() string {
	switch outputFormat {
	case Protobuf:
		return "Protobuf"
	case Json:
		return "Json"
	default:
		return fmt.Sprintf("unknown format %d", int(outputFormat))
	}
}
