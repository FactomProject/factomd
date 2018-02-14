// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"time"

	// "github.com/FactomProject/factomd/common/constants"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/p2p"

	log "github.com/sirupsen/logrus"
)

var ()

var proxyLogger = packageLogger.WithFields(log.Fields{"subpack": "p2p-proxy"})

type P2PProxy struct {
	// A connection to this node:
	ToName   string
	FromName string
	// Channels that define the connection:
	BroadcastOut chan interface{} // FactomMessage ToNetwork from factomd
	BroadcastIn  chan interface{} // FactomMessage FromNetwork for Factomd

	ToNetwork   chan interface{} // p2p.Parcel From p2pProxy to the p2p Controller
	FromNetwork chan interface{} // p2p.Parcel Parcels from the network for the application

	NumPeers int
	bytesOut int // bandwidth used by applicaiton without netowrk fan out
	bytesIn  int // bandwidth recieved by application from network
}

type FactomMessage struct {
	Message  []byte
	PeerHash string
	AppHash  string
	AppType  string
}

func (e *FactomMessage) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *FactomMessage) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *FactomMessage) String() string {
	str, _ := e.JSONString()
	return str
}

var _ interfaces.IPeer = (*P2PProxy)(nil)

func (f *P2PProxy) Weight() int {
	// should return the number of connections this peer represents.  For now, just say a lot
	return f.NumPeers
}

func (f *P2PProxy) SetWeight(w int) {
	// should return the number of connections this peer represents.  For now, just say a lot
	f.NumPeers = w
}

func (f *P2PProxy) BytesOut() int {
	return f.bytesOut
}

func (f *P2PProxy) BytesIn() int {
	return f.bytesIn
}

func (f *P2PProxy) Init(fromName, toName string) interfaces.IPeer {
	f.ToName = toName
	f.FromName = fromName
	f.BroadcastOut = make(chan interface{}, p2p.StandardChannelSize)
	f.BroadcastIn = make(chan interface{}, p2p.StandardChannelSize)

	return f
}

func (f *P2PProxy) GetNameFrom() string {
	return f.FromName
}

func (f *P2PProxy) GetNameTo() string {
	return f.ToName
}

func (f *P2PProxy) Send(msg interfaces.IMsg) error {
	f.logMessage(msg, false) // NODE_TALK_FIX
	data, err := msg.MarshalBinary()
	if err != nil {
		proxyLogger.WithField("send-error", err).Error()
		return err
	}

	proxyLogger.WithFields(msg.LogFields()).WithField("node-name", f.GetNameFrom()).Info("Send Message")

	f.bytesOut += len(data)
	hash := fmt.Sprintf("%x", msg.GetMsgHash().Bytes())
	appType := fmt.Sprintf("%d", msg.Type())
	message := FactomMessage{Message: data, PeerHash: msg.GetNetworkOrigin(), AppHash: hash, AppType: appType}
	switch {
	case !msg.IsPeer2Peer():
		message.PeerHash = p2p.BroadcastFlag
		f.trace(message.AppHash, message.AppType, "P2PProxy.Send() - BroadcastFlag", "a")
	case msg.IsPeer2Peer() && 0 == len(message.PeerHash): // directed, with no direction of who to send it to
		message.PeerHash = p2p.RandomPeerFlag
		f.trace(message.AppHash, message.AppType, "P2PProxy.Send() - RandomPeerFlag", "a")
	default:
		f.trace(message.AppHash, message.AppType, "P2PProxy.Send() - Addressed by hash", "a")
	}
	if msg.IsPeer2Peer() {
		proxyLogger.Info("%s Sending directed to: %s message: %+v\n", time.Now().String(), message.PeerHash, msg.String())
	}
	p2p.BlockFreeChannelSend(f.BroadcastOut, message)

	return nil
}

// Non-blocking return value from channel.
func (f *P2PProxy) Recieve() (interfaces.IMsg, error) {
	select {
	case data, ok := <-f.BroadcastIn:
		if ok {
			BroadInCastQueue.Dec()

			switch data.(type) {
			case FactomMessage:
				fmessage := data.(FactomMessage)
				f.trace(fmessage.AppHash, fmessage.AppType, "P2PProxy.Recieve()", "N")
				msg, err := messages.UnmarshalMessage(fmessage.Message)

				if err != nil {
					proxyLogger.WithField("receive-error", err).Error()
				} else {
					proxyLogger.WithFields(msg.LogFields()).WithField("node-name", f.GetNameFrom()).Info("Receive Message")
				}

				if nil == err {
					msg.SetNetworkOrigin(fmessage.PeerHash)
				}
				//if 1 < f.debugMode {
				//	f.logMessage(msg, true) // NODE_TALK_FIX
				//	fmt.Printf(".")
				//}
				f.bytesIn += len(fmessage.Message)
				return msg, err
			default:
				//fmt.Printf("Garbage on f.BroadcastIn. %+v", data)
			}
		}
	default:
	}
	return nil, nil
}

// Is this connection equal to parm connection
func (f *P2PProxy) Equals(ff interfaces.IPeer) bool {
	f2, ok := ff.(*P2PProxy)
	if !ok {
		return false
	} // Different peer type can't be equal
	// Check If this is another connection from here to there
	if f.FromName == f2.FromName && f.ToName == f2.FromName {
		return true
	}
	// Check if this is a connection from there to here
	if f.FromName == f2.ToName && f.ToName == f2.FromName {
		return true
	}
	return false
}

// Returns the number of messages waiting to be read
func (f *P2PProxy) Len() int {
	return len(f.BroadcastIn)
}

//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////

func (p *P2PProxy) StartProxy() {
	go p.ManageOutChannel() // Bridges between network format Parcels and factomd messages (incl. addressing to peers)
	go p.ManageInChannel()
}

// manageOutChannel takes messages from the f.broadcastOut channel and sends them to the network.
func (f *P2PProxy) ManageOutChannel() {
	for data := range f.BroadcastOut {
		switch data.(type) {
		case FactomMessage:
			fmessage := data.(FactomMessage)
			// Wrap it in a parcel and send it out channel ToNetwork.
			parcels := p2p.ParcelsForPayload(p2p.CurrentNetwork, fmessage.Message)
			for _, parcel := range parcels {
				if parcel.Header.Type != p2p.TypeMessagePart {
					parcel.Header.Type = p2p.TypeMessage
				}
				parcel.Header.TargetPeer = fmessage.PeerHash
				parcel.Header.AppHash = fmessage.AppHash
				parcel.Header.AppType = fmessage.AppType
				parcel.Trace("P2PProxy.ManageOutChannel()", "b")
				p2p.BlockFreeChannelSend(f.ToNetwork, parcel)
			}
		default:
			fmt.Printf("Garbage on f.BrodcastOut. %+v", data)
		}
	}
}

// manageInChannel takes messages from the network and stuffs it in the f.BroadcastIn channel
func (f *P2PProxy) ManageInChannel() {
	for data := range f.FromNetwork {
		switch data.(type) {
		case p2p.Parcel:
			parcel := data.(p2p.Parcel)
			f.trace(parcel.Header.AppHash, parcel.Header.AppType, "P2PProxy.ManageInChannel()", "M")
			message := FactomMessage{Message: parcel.Payload, PeerHash: parcel.Header.TargetPeer, AppHash: parcel.Header.AppHash, AppType: parcel.Header.AppType}
			removed := p2p.BlockFreeChannelSend(f.BroadcastIn, message)
			BroadInCastQueue.Inc()
			BroadInCastQueue.Add(float64(-1 * removed))
			BroadCastInQueueDrop.Add(float64(removed))
		default:
			fmt.Printf("Garbage on f.FromNetwork. %+v", data)
		}
	}
}

func (p *P2PProxy) trace(appHash string, appType string, location string, sequence string) {
	if 10 < p.debugMode {
		time := time.Now().Unix()
		fmt.Printf("\nParcelTrace, %s, %s, %s, Message, %s, %d \n", appHash, sequence, appType, location, time)
	}
}
