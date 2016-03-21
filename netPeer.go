// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"net"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

var _ = fmt.Print

type NetPeer struct {
	Conn     net.Conn
	ToName   string
	FromName string
}

//var _ interfaces.IPeer = (*NetPeer)(nil)

func (f *NetPeer) AddExistingConnection(conn net.Conn) {
	f.Conn = conn
}

func (f *NetPeer) Connect(network, address string) error {
	c, err := net.Dial(network, address)
	if err != nil {
		return err
	}
	f.Conn = c
	return nil
}

func (f *NetPeer) ConnectTCP(address string) error {
	return f.Connect("tcp", address)
}

func (f *NetPeer) ConnectUDP(address string) error {
	return f.Connect("udp", address)
}

func (f *NetPeer) Init(fromName, toName string) *NetPeer { // interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	return f
}

func (f *NetPeer) GetNameFrom() string {
	return f.FromName
}
func (f *NetPeer) GetNameTo() string {
	return f.ToName
}

func (f *NetPeer) Send(msg interfaces.IMsg) error {
	data, err := msg.MarshalBinary()
	if err != nil {
		return err
	}

	_, err = f.Conn.Write(data)
	return err
}

// Non-blocking return value from channel.
func (f *NetPeer) Recieve() (interfaces.IMsg, error) {
	data := make([]byte, 5000)

	n, err := f.Conn.Read(data)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		msg, err := messages.UnmarshalMessage(data)
		return msg, err
	}
	return nil, nil
}
