// This set of functions is used to schedule when a specific feature or data format becomes effective

package activations

import (
	"fmt"
	"os"

	"github.com/FactomProject/factomd/common/globals"
)

type ActivationType int

const (
	_                ActivationType = iota // 0 Don't use ZERO
	ELECTION_NO_SORT                = iota // 1 -- this is a passing activation and this ID may be reused once that height is passes and the references are removed

	//
	ACTIVATION_TYPE_COUNT = iota - 1 // Always Last
)

type Activation struct {
	Name             string
	Id               ActivationType
	Description      string
	ActivationHeight map[string]int // this maps a network Name to the height for that network for the feature to activate
}

var ActivationMap map[ActivationType]Activation
var ActivationNameMap map[ActivationType]string

func init() {

	// unordered list of activations
	var activations []Activation = []Activation{
		Activation{"ElectionNoSort", ELECTION_NO_SORT, "Disable sorting of severs after elections",
			map[string]int{
				"MAIN":                      146060 + 8*24*10 + 1, // On 6/20/18 11:45 mainnet was 146060, we want activation at 6/28/18 at ~12pm
				"TEST":                      0,                    // On 6/20/18 11:45 testnet was 146060, we want activation at 6/22/18 at ~12pm
				"LOCAL":                     10,                   // Must be > 6 for TestActivationHeightElection to pass
				"CUSTOM:fct_community_test": 33037 + 2*24*10 + 1,  // On 6/22/18 11:45 testnet was 33037, we want activation at 6/24/18 at 12:00pm
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

// convert an Activation ID to a name
func (id ActivationType) String() string {

	n, ok := ActivationNameMap[id]
	if !ok {
		n = fmt.Sprintf("ActivationId(%v)", id)
	}
	return n
}

func IsActive(id ActivationType, height int) bool {

	netName := globals.Params.NetworkName
	if netName == "CUSTOM" {
		netName = fmt.Sprintf("%s:%s", netName, globals.Params.CustomNetName)
	}
	a, ok := ActivationMap[id]

	if !ok {
		fmt.Fprintf(os.Stderr, "Invalid %v (%s)\n", id, id.String())
		return false
	}

	h, ok := a.ActivationHeight[netName]
	if !ok {
		fmt.Fprintf(os.Stderr, "Activation %s does not know network name %s. Activating at 0.\n", id.String(), netName)
		a.ActivationHeight[netName] = 0
		return true
	}

	return height >= h
}
