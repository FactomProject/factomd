package state

import (
	"github.com/FactomProject/factomd/common/factoid"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
)

type HardGrant struct {
	DBh     uint32
	Amount  uint64
	Address interfaces.IAddress
}

// Return the Hard Coded Grants. Buried in an func so other code cannot easily Address the array and change it
func GetHardCodedGrants() []HardGrant {
	var hardcodegrants = [...]HardGrant{
		// waiting for "real-ish" data from brian
		HardGrant{10, 2, factoid.NewAddress(primitives.ConvertUserStrToAddress("FA3oajkmHMfqkNMMShmqpwDThzMCuVrSsBwiXM2kYFVRz3MzxNAJ"))}, // Pay Clay 2 at dbheight 10
		HardGrant{12, 4, factoid.NewAddress(primitives.ConvertUserStrToAddress("FA3Ga2XcaheS5NgQ3q22gBpLgE6tXmPu1GhjdU2FsdN2QPMzKJET"))}, // Pay Bob 4 at dbheight 12
		HardGrant{10, 3, factoid.NewAddress(primitives.ConvertUserStrToAddress("FA3Ga2XcaheS5NgQ3q22gBpLgE6tXmPu1GhjdU2FsdN2QPMzKJET"))}, // Pay Bob 3 at dbheight 10
		HardGrant{9, 1, factoid.NewAddress(primitives.ConvertUserStrToAddress("FA3GH7VEFKqTdJcmwGgDrcY4Xh9njQ4EWiJxhJeim6BCA7QuB388"))},  // Pay Bill 1 at dbheight 9
	}
	// Passing an array to a function creates a copy and I return a slide anchored to that copy
	// so the caller does not have the Address of the array itself.
	// Closest thing to a constant array I can get in Go
	return func(x [len(hardcodegrants)]HardGrant) []HardGrant { return x[:] }(hardcodegrants)
}

//return a (possibly empty) of coinbase payouts to be scheduled at this height
func GetGrantPayoutsFor(currentDBHeight uint32) []interfaces.ITransAddress {

	outputs := make([]interfaces.ITransAddress, 0)
	// this is only but temporary, once the hard coded grants are payed this code will go away
	// I can't modify the grant list because in simulation it is shared across nodes so for now I just
	// scan the whole list once every 25 blocks
	// I opted for one list knowing it will have to be different for testnet vs mainnet because making it
	// network sensitive just add complexity to the code.
	// there is no need for activation height because the grants have inherent activation heights per grant
	for _, g := range GetHardCodedGrants() { // check every hardcoded grant
		if g.DBh == currentDBHeight { // if it's ready {...
			o := factoid.NewOutAddress(g.Address, g.Amount) // Create a payout
			outputs = append(outputs, o)                    // and add it to the list
		}
	}
	return outputs
}
