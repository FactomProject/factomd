package eventservices

import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBroadcastContent_Parse(t *testing.T) {
	testCases := []struct {
		Input  string
		Output BroadcastContent
		Error  error
	}{
		{"always", BroadcastAlways, nil},
		{"once", BroadcastOnce, nil},
		{"never", BroadcastNever, nil},
		{"test", -1, errors.New("could not parse test to EventBroadcastContent")},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Input, func(t *testing.T) {
			content, err := Parse(testCase.Input)

			assert.Equal(t, testCase.Error, err)
			assert.Equal(t, testCase.Output, content)
		})
	}
}

func TestBroadcastContent_String(t *testing.T) {
	testCases := []struct {
		Input  BroadcastContent
		Output string
	}{
		{BroadcastAlways, "always"},
		{BroadcastOnce, "once"},
		{BroadcastNever, "never"},
		{-1, "unknown value -1"},
		{4, "unknown value 4"},
	}

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("%d", testCase.Input), func(t *testing.T) {
			output := testCase.Input.String()

			assert.Equal(t, testCase.Output, output)
		})
	}
}
