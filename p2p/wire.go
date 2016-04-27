package p2p

import (
    "time"
    "math/rand"
)


type P2PConnection struct {
    conn net.Conn
    OutChannel chan Parcel // Out means "towards the network"
	InChannel  chan Parcel // In means "from the network"
    ConnectionID uint64 // Random number used for loopback protection 
    // and as "address" for sending messages to specific nodes.
}

func (p *P2PConnection) Init()   {
    	p.OutChannel = make(chan Parcel, 1000)
	p.InChannel = make(chan Parcel, 1000)
    r := rand.New(rand.NewSource(time.Now().UnixNano()))
    p.ConnectionID = r.Int63()
    // f.ConnectionID = rand.Int(rand.Reader, math.MaxInt64)
}

func (p *P2PConnection) SimpleSend(payload []byte)  {
    		header := new(ParcelHeader).Init().(*ParcelHeader)
		parcel := new(Parcel).Init(header).(*Parcel)
        parcel.payload = payload
        parcel.header.PeerID = p.ConnectionID
        p.OutChannel <- parcel
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

func (p * P2PConnection) dial(address string)  {
    
    
}

func (p * P2PConnection) listen(address string)  {
    
}



// // Sender is a goroutine that handles sending parcels out this connection
// func (p *P2PConnection) Sender()  {
    
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


