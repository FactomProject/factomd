// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package state

// This file is for the simulator to attach identities properly to the state.
// Each state has its own set of keys that need to match the ones in the
// identity to properly test identities/authorities
import (
	"github.com/FactomProject/factomd/common/primitives"
)

func (s *State) SimSetNewKeys(p *primitives.PrivateKey) {
	s.serverPrivKey = p
	s.serverPubKey = p.Pub
}

func (s *State) SimGetSigKey() string {
	return s.serverPrivKey.Pub.String()
}
