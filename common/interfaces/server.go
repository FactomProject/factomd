// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

import (
	"fmt"
)

// This object will hold the public keys for servers that are not
// us, and maybe other information about servers.
type IServer interface {
	GetChainID() IHash
	String() string
	GetName() string
	IsOnline() bool
	SetOnline(bool)
	LeaderToReplace() IHash
	SetReplace(IHash)
}

type Server struct {
	ChainID IHash
	Name    string
	Online  bool
	Replace IHash
}

var _ IServer = (*Server)(nil)

func (s *Server) GetName() string {
	return s.Name
}

func (s *Server) GetChainID() IHash {
	return s.ChainID
}

func (s *Server) String() string {
	return fmt.Sprintf("%s %s", "Server:", s.GetChainID().String())
}

func (s *Server) IsOnline() bool {
	return s.Online
}

func (s *Server) SetOnline(o bool) {
	s.Online = o
}

func (s *Server) LeaderToReplace() IHash {
	return s.Replace
}

func (s *Server) SetReplace(h IHash) {
	s.Replace = h
}
