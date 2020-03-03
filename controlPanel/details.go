package controlpanel

import (
	"fmt"
	"strings"
)

func nodeInfo(nodeName string, identityChainID string, publicKey string) string {
	var info strings.Builder

	fmt.Fprintf(&info, "My Node: %s\n", nodeName)
	fmt.Fprintf(&info, "Identity ChainID: %s\n", identityChainID)
	fmt.Fprintf(&info, "Signing Key: %s\n", publicKey)

	return info.String()
}
