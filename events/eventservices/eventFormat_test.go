package eventservices

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEventFormat_FormatFrom(t *testing.T) {
	testCases := []struct {
		Input  string
		Output EventFormat
	}{
		{"protobuf", Protobuf},
		{"json", Json},
		{"test", -1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Input, func(t *testing.T) {
			eventFormat := EventFormatFrom(testCase.Input, -1)
			assert.Equal(t, testCase.Output, eventFormat)
		})
	}
}

func TestEventFormat_String(t *testing.T) {
	testCases := []struct {
		Input  EventFormat
		Output string
	}{
		{Protobuf, "Protobuf"},
		{Json, "Json"},
		{-1, "unknown format -1"},
		{4, "unknown format 4"},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%d", testCase.Input), func(t *testing.T) {
			output := testCase.Input.String()

			assert.Equal(t, testCase.Output, output)
		})
	}
}
