package contentfiltermode

import (
	"fmt"
)

type ContentFilterMode int

const (
	SendNever          ContentFilterMode = 0
	SendOnRegistration ContentFilterMode = 1
	SendAlways         ContentFilterMode = 2
	Unknown            ContentFilterMode = -1
)

func Parse(value string) ContentFilterMode {
	switch value {
	case "SendAlways":
		return SendAlways
	case "SendOnRegistration":
		return SendOnRegistration
	case "SendNever":
		return SendNever
	}
	return Unknown
}

func (contentFilterMode ContentFilterMode) String() string {
	switch contentFilterMode {
	case SendAlways:
		return "send alwyas"
	case SendOnRegistration:
		return "send once"
	case SendNever:
		return "never send"
	default:
		return fmt.Sprintf("Unknown mode %d", int(contentFilterMode))
	}
}
