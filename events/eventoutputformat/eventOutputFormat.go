package eventoutputformat

import (
	"fmt"
	"strings"
)

type Format int

const (
	Protobuf Format = 1
	Json     Format = 2
)

func (outputFormat Format) String() string {
	switch outputFormat {
	case Protobuf:
		return "Protobuf"
	case Json:
		return "Json"
	default:
		return fmt.Sprintf("Unknown output format %d", int(outputFormat))
	}
}

func FormatFrom(value string, defaultFormat Format) Format {
	switch strings.ToLower(value) {
	case strings.ToLower(Protobuf.String()):
		return Protobuf
	case strings.ToLower(Json.String()):
		return Json
	default:
		return defaultFormat
	}
}
