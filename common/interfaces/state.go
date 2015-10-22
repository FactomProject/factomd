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
	
	// Server
	Cfg()					IFactomConfig
	Init()		
	
	// Channels
	InMsgQueue()            chan IMsg
	LeaderInMsgQueue() 		chan IMsg
	FollowerInMsgQueue()	chan IMsg
	OutMsgQueue()			chan IMsg		
	
	//Network MAIN = 0, TEST = 1, LOCAL = 2, CUSTOM = 3
	NetworkNumber()		int			// Encoded into Directory Blocks
	NetworkName()		string		// Some networks have defined names
	
	// Number of Servers acknowledged by Factom
	TotalServers() int 
	ServerState()  int      		// (0 if client, 1 if server, 2 if audit server
	Matryoshka()   []IHash  		// Reverse Hash
	
	LeaderFor([]byte) bool			// Tests if this server is the leader for this key
	
	// Database
	DB() IDatabase
	SetDB(IDatabase)
	
	// Directory Block State
	CurrentDirectoryBlock() IDirectoryBlock			// The directory block under construction
	SetCurrentDirectoryBlock(IDirectoryBlock)
	DBHeight() int                                  // The index of the directory block under construction.
	SetDBHeight(int)    
	
	// Message State
	LastAck() IMsg					// Return the last Acknowledgement set by this server
	
	NewHash() IHash					// Return a new Hash object
}

