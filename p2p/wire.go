package p2p

import (
    
)

type P2PConnection struct {
    conn net.Conn
    OutChannel chan []byte // Out means "towards the network"
	InChannel  chan []byte // In means "from the network"
    ConnectionID uint64 // Random number used for loopback protection 
    // and as "address" for sending messages to specific nodes.
}

func (p *P2PConnection) Init()   {
    	f.OutChannel = make(chan Parcel, 1000)
	f.InChannel = make(chan Parcel, 1000)

}


// ProcessNetworkMessages gets all the messages from the network and sends them to the application
func (p *P2PConnection) ProcessNetworkMessages()  {
    for -while there are more messages to recieve- {
        BUGBUG need to send these messages to application
        p.OutChannel <- message
    }   
}

// ProcessOutChannel gets all the messages from the application and sends them out over the network
func (p *P2PConnection) ProcessInChannel()  {
    for 0 < len(p.InChannel) {
        BUGBUG need to send these messages to the network
    }   
}

// Sender is a goroutine that handles sending parcels out this connection
func (p *P2PConnection) Sender()  {
    
}

func (p * P2PConnection) send()  {
    
   
}

// Reciever is a goroutine that handles recieving parcels out this connection
func (p * P2PConnection) Reciever()  {
}

func (p * P2PConnection) receive()  {
    parcel Parcel
    
    if err := binary.Read(p.c)
    
    // if PeerID is our connectionID then this message came from ourselves.
}