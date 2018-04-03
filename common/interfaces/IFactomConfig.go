// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

// Holds the state information for factomd.  This does imply that we will be
// using accessors to access state information in the consensus algorithm.
// This is a bit tedious, but does provide single choke points where information
// can be logged about the execution of Factom.  Also ensures that we do not
// accidentally
type IFactomConfig interface {
}
