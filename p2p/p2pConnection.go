// Copyright 2016 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.


package p2p

import (
    "time"
    "math/rand"
    "encoding/binary"
)


type P2PConnection struct {
    conn net.Conn
    OutChannel chan Parcel // Out means "towards the network"
	InChannel  chan Parcel // In means "from the network"
    ConnectionID uint64 // Random number used for loopback protection 
    // and as "address" for sending messages to specific nodes.
}

func (c *P2PConnection) Init()   {
    	p.OutChannel = make(chan Parcel, 1000)
	p.InChannel = make(chan Parcel, 1000)
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    p.ConnectionID = r.Int63()
    // f.ConnectionID = rand.Int(rand.Reader, math.MaxInt64)
    go runloop()
}



func (c *P2PConnection) runloop()  {

    for {
           // For each peer, empty its network buffer (do recieves first)
        for _, connection := range peers {
            connection.processNetworkMessages()
        }
        // For each peer, empty its outbound channel (and send to network.)
        for _, connection := range peers {
            connection.processInChannel()
  
    }
          }
}
// ProcessNetworkMessages gets all the messages from the network and sends them to the application
func (c *P2PConnection) processNetworkMessages()  {
    for -while there are more messages to recieve- {
        BUGBUG need to send these messages to application
        p.OutChannel <- message
    }   
}

// ProcessOutChannel gets all the messages from the application and sends them out over the network
func (c *P2PConnection) processInChannel()  {
    for 0 < len(p.InChannel) {
        BUGBUG need to send these messages to the network
    }   
}

func (p * P2PConnection) dial(address string)  {
    
    
}

func (p * P2PConnection) listen(address string)  {
    
}



// // Sender is a goroutine that handles sending parcels out this connection
// func (c *P2PConnection) Sender()  {
    
// }

func (p * P2PConnection) send([]byte)  {
    
   
}

// // Reciever is a goroutine that handles recieving parcels out this connection
// func (p * P2PConnection) Reciever()  {
// }

func (p * P2PConnection) receive()  {
    parcel Parcel
    
    if err := binary.Read(p.c)
    
    // if PeerID is our connectionID then this message came from ourselves.
}

func (p * P2PConnection) gotBadMessage(address string)  {
|
    // TODO Track bad messages to ban bad peers at network level
    // Array of in P2PConnection of bad messages
    // Add this one to the array with timestamp
    // Filter all messages with timestamps over an hour (put value in protocol.go maybe an hour is too logn)
    // If count of bad messages in last hour exceeds threshold from protocol.go then we drop connection
    // Add this IP address to our banned peers (for an hour or day, also define in protocol.go)
|}
