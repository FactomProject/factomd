package p2p

// Parcel is the atomic level of communication for the p2p network.  It contains within it the necessary info for 
// the networking protocol, plus the message that the Application is sending.
type Parcel struct {
     header ParcelHeader
     payload []byte
}


// ParcelHeaderSize is the number of bytes in a parcel header
const ParcelHeaderSize = 28

type ParcelHeader struct {
	Cookie   uint32     // 4 bytes - magic cookie "Fact"
    Network  NetworkID  // 4 bytes - the network we are on (eg testnet, main net, etc.)
    Version  uint16     // 2 bytes - the version of the protocol we are running.
	Type     uint16     // 2 bytes - network level commands (eg: ping/pong)
	Length   uint32     // 4 bytes - length of the payload (that follows this header) in bytes
    PeerID   uint64     // 8 bytes - 
	Hash [4]byte        // 4 bytes - data integrity hash (of the payload itself.)
}

// Parcel commands -- all new commands should be added to the *end* of the list!
const uint16 ( // iota is reset to 0
	TypeHeartbeat = iota    // "Note, I'm still alive"
	TypePing                // "Are you there?"
    TypePong                // "yes, I'm here"
    TypePang                // "Roger that"
    TypeNetworkError        // eg: "you sent me a message larger than max payload ParcelHeaderSize""
    TypeAlert               // network wide alerts (used in bitcoin to indicate criticalities)
    TypeMessage             // Application level message
)

// MaxPayloadSize is the maximum bytes a message can be at the networking level.
const MaxPayloadSize = (1024 * 1024 * 1) // 1MB

func (p *ParcelHeader) Init()  {
    p.Cookie = ProtocolCookie
    p.NetworkID = TestNet
    p.Version = ProtocolVersion
    p.Type = TypeMessage
}
func (p *Parcel) Init(header ParcelHeader)  {
    p.header = header
}