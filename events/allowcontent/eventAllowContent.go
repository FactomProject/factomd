package allowcontent

import (
	"errors"
	"fmt"
)

type AllowContent int

const (
	Never          AllowContent = 0
	OnRegistration AllowContent = 1
	Always         AllowContent = 2
)

func Parse(value string) (AllowContent, error) {
	switch value {
	case "Always":
		return Always, nil
	case "OnRegistration":
		return OnRegistration, nil
	case "Never":
		return Never, nil
	}
	return -1, errors.New(fmt.Sprintf("Could not parse %s to AllowContent", value))
}

func (allowContent AllowContent) String() string {
	switch allowContent {
	case Always:
		return "Always"
	case OnRegistration:
		return "OnRegistration"
	case Never:
		return "Never"
	default:
		return fmt.Sprintf("Unknown value %d", int(allowContent))
	}
}
