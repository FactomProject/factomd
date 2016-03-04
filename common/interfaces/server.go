// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import (
	"fmt"
	
)

// This object will hold the public keys for servers that are not
// us, and maybe other information about servers.
type IServer interface {
	GetChainID()	IHash
	String() 		string
}

type Server struct {
	ChainID	IHash
}

var _ IServer = (*Server)(nil)

func (s *Server) GetChainID() IHash {
	return s.ChainID
}

func (s *Server) String() string {
	return fmt.Sprintf("%s %s","Server:", s.GetChainID().String())
}

