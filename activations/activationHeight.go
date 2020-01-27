// This set of functions is used to schedule when a specific feature or data format becomes effective

package activations

import (
	"fmt"
	"math"
	"os"

	"github.com/FactomProject/factomd/common/globals"
)

type ActivationType int

const (
	_                       ActivationType = iota // 0 Don't use ZERO
	TESTNET_COINBASE_PERIOD                = iota // 1 -- this is a passing activation and this ID may be reused once that height is passes and the references are removed
	//
	AUTHRORITY_SET_MAX_DELTA = iota
	ACTIVATION_TYPE_COUNT    = iota - 1 // Always Last
)

type Activation struct {
	Name             string
	Id               ActivationType
	Description      string
	DefaultHeight    int            // height of activation on nets not expressly listed (math.MaxInt32 means never)
	ActivationHeight map[string]int // this maps a network Name to the height for that network for the feature to activate
}

var ActivationMap map[ActivationType]Activation
var ActivationNameMap map[ActivationType]string

// init initializes the two global variables ActivationMap and ActivationNameMap above
func init() {

	// unordered list of activations
	var activations []Activation = []Activation{
		Activation{"TestNetCoinBasePeriod", TESTNET_COINBASE_PERIOD,
			"Change testnet coin base payout delay to 140 blocks",
			math.MaxInt32, // inactive unless overridden below
			map[string]int{
				"MAIN":                      math.MaxInt32,
				"LOCAL":                     25,
				"CUSTOM:fct_community_test": 45335, //  Monday morning September 17
			},
		},
		Activation{"AuthorityMaxDelta", AUTHRORITY_SET_MAX_DELTA,
			"Ensures fewer than half of federated notes are replaced in a single election",
			math.MaxInt32, // inactive unless overridden below
			map[string]int{
				"MAIN":                      222874,
				"LOCAL":                     25,
				"CUSTOM:fct_community_test": 109387,
			},
		},
	}

	if ACTIVATION_TYPE_COUNT != len(activations) {
		// Really a compile issue but I don't know how to catch it then
		panic("ACTIVATION_TYPE_COUNT does not match the list of Activations")
	}

	ActivationMap = make(map[ActivationType]Activation, len(activations))
	ActivationNameMap = make(map[ActivationType]string, len(activations))
	for _, a := range activations {
		ActivationMap[a.Id] = a
		ActivationNameMap[a.Id] = a.Name
	}
}

// String converts an Activation ID to a name
func (id ActivationType) String() string {

	n, ok := ActivationNameMap[id]
	if !ok {
		//n = fmt.Sprintf("ActivationId(%v)", id)
	}
	return n
}

var once bool
var netName string

// networkname returns the network name specified by the user at startup with the factomd '-network' option
// see constants.go for options: MAIN / LOCAL / TEST / CUSTOM
func networkname() string {
	if !once {
		once = true
		netName = globals.Params.NetworkName
		if netName == "CUSTOM" {
			netName = fmt.Sprintf("%s:%s", netName, globals.Params.CustomNetName)
		}
		fmt.Printf("Using NetworkName \"%s\"\n", netName)
	}
	return netName
}

// IsActive returns whether the input activation is 'on' at the input height
func IsActive(id ActivationType, height int) bool {
	netName := networkname()
	a, ok := ActivationMap[id]

	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid %v (%s)\n", id, id.String())
		return false
	}

	h, ok := a.ActivationHeight[netName]
	if !ok {
		a.ActivationHeight[netName] = a.DefaultHeight
		if a.DefaultHeight != math.MaxInt32 {
			fmt.Fprintf(os.Stderr, "Activation %s does not know network name \"%s\". Activating at %d.\n", id.String(), netName, a.DefaultHeight)
		} else {
			fmt.Fprintf(os.Stderr, "Activation %s does not know network name \"%s\". Never activating.\n", id.String(), netName)
		}
		return true
	}

	return height >= h
}
