// This set of functions is used to schedule when a specific feature or data format becomes effective

package activations

import (
	"fmt"
	"math"
	"os"
	"sync"

	"github.com/FactomProject/factomd/common/globals"
)

type ActivationType int

const (
	_                       ActivationType = iota // 0 Don't use ZERO
	TESTNET_COINBASE_PERIOD                = iota // 1 -- this is a passing activation and this ID may be reused once that height is passes and the references are removed
	//
	AUTHRORITY_SET_MAX_DELTA = iota
	MAX_FACTOM_HEIGHT        = iota
	ACTIVATION_TYPE_COUNT    = iota - 1 // Always Last
)

type activation struct {
	name             string
	id               ActivationType
	description      string
	defaultHeight    int            // height of activation on nets not expressly listed (math.MaxInt32 means never)
	activationHeight map[string]int // this maps a network Name to the height for that network for the feature to activate
}

var activationMap map[ActivationType]activation
var activationNameMap map[ActivationType]string

// init initializes the two global variables activationMap and activationNameMap above
func init() {

	// unordered list of activations
	var activations []activation = []activation{
		{"TestNetCoinBasePeriod", TESTNET_COINBASE_PERIOD,
			"Change testnet coin base payout delay to 140 blocks",
			0, // always active for consistency
			map[string]int{
				"MAIN":                      math.MaxInt32,
				"LOCAL":                     25,
				"CUSTOM:fct_community_test": 45335, //  Monday morning September 17
			},
		},
		{"AuthorityMaxDelta", AUTHRORITY_SET_MAX_DELTA,
			"Ensures fewer than half of federated notes are replaced in a single election",
			0, // always active for consistency
			map[string]int{
				"MAIN":                      222874,
				"LOCAL":                     25,
				"CUSTOM:fct_community_test": 109387,
			},
		},
		{"FactomMaxHeight", MAX_FACTOM_HEIGHT,
			"The maximum Height of the Factom Blockchain prior to Accumulate Activation",
			math.MaxInt32, //                           Don't activate by default
			map[string]int{ //
				"MAIN":                      374000, // The Last Block of the Factom ERA
				"LOCAL":                     400000, // Factom will continue processing into the next block, but
				"CUSTOM:fct_community_test": 400000, // stop at minute 2.  That last block will not be included
			}, //                                       in the History provided to the Accumulate ERA.
		},
	}

	if ACTIVATION_TYPE_COUNT != len(activations) {
		// Really a compile issue but I don't know how to catch it then
		panic("ACTIVATION_TYPE_COUNT does not match the list of Activations")
	}

	activationMap = make(map[ActivationType]activation, len(activations))
	activationNameMap = make(map[ActivationType]string, len(activations))
	for _, a := range activations {
		activationMap[a.id] = a
		activationNameMap[a.id] = a.name
	}
}

// String converts an Activation ID to a name
func (id ActivationType) String() string {
	if n, ok := activationNameMap[id]; ok {
		return n
	}
	return fmt.Sprintf("ActivationId(%d)", id)
}

var netNameOnce sync.Once
var netName string

// networkname returns the network name specified by the user at startup with the factomd '-network' option
// see constants.go for options: MAIN / LOCAL / TEST / CUSTOM
func networkname() string {
	netNameOnce.Do(func() {
		netName = globals.Params.NetworkName
		if netName == "CUSTOM" {
			netName = fmt.Sprintf("CUSTOM:%s", globals.Params.CustomNetName)
		}
		fmt.Printf("Using NetworkName \"%s\"\n", netName)
	})
	return netName
}

// IsActive returns whether the input activation is 'on' at the input height
func IsActive(id ActivationType, height int) bool {
	netName := networkname()
	a, ok := activationMap[id]

	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid %d (%s)\n", id, id.String())
		return false
	}

	// has a custom entry
	if h, ok := a.activationHeight[netName]; ok {
		return height >= h
	}

	// use default
	if a.defaultHeight < math.MaxInt32 {
		return height >= a.defaultHeight
	}

	return false
}
