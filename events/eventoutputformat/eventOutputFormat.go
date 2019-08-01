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
		return "Protocol buffers"
	case Json:
		return "JSON"
	default:
		return fmt.Sprintf("Unknown output format %d", int(outputFormat))
	}
}

func FormatFrom(value string, defaultFormat Format) Format {
	switch strings.ToLower(value) {
	case "protobof":
		return Protobuf
	case "json":
		return Json
	default:
		return defaultFormat
	}
}
