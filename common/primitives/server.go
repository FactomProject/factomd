// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package primitives

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives/random"
)

type Server struct {
	ChainID interfaces.IHash
	Name    string
	Online  bool
	Replace interfaces.IHash
}

var _ interfaces.IServer = (*Server)(nil)
var _ interfaces.BinaryMarshallable = (*Server)(nil)

func (s *Server) Init() {
	if s.ChainID == nil {
		s.ChainID = NewZeroHash()
	}
	if s.Replace == nil {
		s.Replace = NewZeroHash()
	}
}

func RandomServer() interfaces.IServer {
	s := new(Server)
	s.Init()
	s.ChainID = RandomHash()
	s.Name = random.RandomString()
	s.Online = (random.RandInt()%2 == 0)
	s.Replace = RandomHash()
	return s
}

func (s *Server) IsSameAs(b interfaces.IServer) bool {
	serv := b.(*Server)
	if s.ChainID.IsSameAs(serv.ChainID) == false {
		return false
	}
	if s.Name != serv.Name {
		return false
	}
	if s.Online != serv.Online {
		return false
	}
	if s.Replace.IsSameAs(serv.Replace) == false {
		return false
	}
	return true
}

func (s *Server) MarshalBinary() ([]byte, error) {
	buf := new(Buffer)

	b := s.ChainID.Bytes()
	_, err := buf.Write(b)
	if err != nil {
		return nil, err
	}

	l := uint32(len(s.Name))
	err = binary.Write(buf, binary.BigEndian, &l)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write([]byte(s.Name))
	if err != nil {
		return nil, err
	}
	if s.Online == false {
		err = buf.WriteByte(0x00)
		if err != nil {
			panic(err)
			return nil, err
		}
	} else {
		err = buf.WriteByte(0x01)
		if err != nil {
			panic(err)
			return nil, err
		}
	}

	b = s.Replace.Bytes()
	_, err = buf.Write(b)
	if err != nil {
		return nil, err
	}

	return buf.DeepCopyBytes(), nil
}

func (s *Server) UnmarshalBinaryData(p []byte) (newData []byte, err error) {
	s.Init()
	buf := bytes.NewBuffer(p)

	hash := make([]byte, 32)
	newData = p

	_, err = buf.Read(hash)
	if err != nil {
		return
	} else {
		err = s.ChainID.UnmarshalBinary(hash)
		if err != nil {
			panic(err)
			return
		}
	}

	var strLen uint32

	err = binary.Read(buf, binary.BigEndian, &strLen)
	if err != nil {
		return
	}
	fmt.Printf("Len - %v\n", strLen)

	str := make([]byte, int(strLen))

	_, err = buf.Read(str)
	if err != nil {
		return
	} else {
		s.Name = string(str)
	}

	b, err := buf.ReadByte()
	if err != nil {
		return
	} else {
		s.Online = b > 0x00
	}

	_, err = buf.Read(hash)
	if err != nil {
		return
	} else {
		err = s.Replace.UnmarshalBinary(hash)
		if err != nil {
			panic(err)
			return
		}
	}

	newData = buf.Bytes()
	return
}

func (s *Server) UnmarshalBinary(p []byte) error {
	_, err := s.UnmarshalBinaryData(p)
	return err
}

func (s *Server) GetName() string {
	return s.Name
}

func (s *Server) GetChainID() interfaces.IHash {
	return s.ChainID
}

func (s *Server) String() string {
	return fmt.Sprintf("Server[:4]: %x", s.GetChainID().Bytes()[:10])
}

func (s *Server) IsOnline() bool {
	return s.Online
}

func (s *Server) SetOnline(o bool) {
	s.Online = o
}

func (s *Server) LeaderToReplace() interfaces.IHash {
	return s.Replace
}

func (s *Server) SetReplace(h interfaces.IHash) {
	s.Replace = h
}
