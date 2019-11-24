package p2p

// ParcelType is a list of parcel types that this node understands
type ParcelType uint16

const ( // iota is reset to 0
	// TypeHeartbeat is deprecated
	TypeHeartbeat ParcelType = iota
	// TypePing is sent if no other parcels have been sent in a while
	TypePing
	// TypePong is a response to a Ping
	TypePong
	// TypePeerRequest indicates a peer wants to be be updated of endpoints
	TypePeerRequest
	// TypePeerResponse carries a payload with protocol specific endpoints
	TypePeerResponse
	// TypeAlert is deprecated
	TypeAlert
	// TypeMessage carries an application message in the payload
	TypeMessage
	// TypeMessagePart is a partial message. deprecated in p2p 2.0
	TypeMessagePart
	// TypeHandshake is the first parcel sent after making a connection
	TypeHandshake
	// TypeRejectAlternative is sent instead of a handshake if the server refuses connection
	TypeRejectAlternative
)

var typeStrings = map[ParcelType]string{
	TypeHeartbeat:         "Heartbeat",
	TypePing:              "Ping",
	TypePong:              "Pong",
	TypePeerRequest:       "Peer-Request",
	TypePeerResponse:      "Peer-Response",
	TypeAlert:             "Alert",
	TypeMessage:           "Message",
	TypeMessagePart:       "MessagePart",
	TypeHandshake:         "Handshake",
	TypeRejectAlternative: "Rejection-Alternative",
}

func (t ParcelType) String() string {
	return typeStrings[t]
}
