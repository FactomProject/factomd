// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import ()

// Holds the state information for factomd.  This does imply that we will be
// using accessors to access state information in the consensus algorithm.  
// This is a bit tedious, but does provide single choke points where information
// can be logged about the execution of Factom.  Also ensures that we do not
// accidentally 
type IState interface {
	
	// Database
	DB IDatabase
	SetDB(IDatabase)
	
	// Directory Block State
	CurrentDirectoryBlock() IDirectoryBlock			// The directory block under construction
	SetCurrentDirectoryBlock(IDirectoryBlock)
	DBHeight() int                                  // The index of the directory block under construction.
	SetDBHeight(int)                                 
}

