// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bytes"
	"fmt"
	"os"

	// "github.com/FactomProject/factomd/common/constants"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/p2p"

	"github.com/FactomProject/factomd/common/messages/msgsupport"
	log "github.com/sirupsen/logrus"
)

var (
	proxyLogger = packageLogger.WithFields(log.Fields{
		"subpack":   "p2p-proxy",
		"component": "networking"})
)

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
	bytesOut int // bandwidth used by application without network fan out
	bytesIn  int // bandwidth received by application from network

	// logging
	logger *log.Entry
}

type FactomMessage struct {
	Message  []byte
	PeerHash string
	AppHash  string
	AppType  string
	msg      interfaces.IMsg // Keep the original message for debugging and peer selection optimization
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
	f.logger = proxyLogger.WithField("node", fromName)
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
	data, err := msg.MarshalBinary()
	if err != nil {
		f.logger.WithField("send-error", err).Error()
		return err
	}

	msgLogger := f.logger.WithFields(msg.LogFields())

	f.bytesOut += len(data)
	if msg.GetMsgHash() == nil || bytes.Equal(msg.GetMsgHash().Bytes(), constants.ZERO_HASH) {
		fmt.Fprintf(os.Stderr, "nil hash message in p2pProxy.Send() %s\n", msg.String())
		fmt.Fprintf(os.Stderr, "nil hash message in p2pProxy.Send() %+v\n", msg)
	} else {
		hash := fmt.Sprintf("%x", msg.GetMsgHash().Bytes())
		appType := fmt.Sprintf("%d", msg.Type())
		message := FactomMessage{Message: data, PeerHash: msg.GetNetworkOrigin(), AppHash: hash, AppType: appType, msg: msg}
		switch {
		case !msg.IsPeer2Peer() && msg.IsFullBroadcast():
			msgLogger.Debug("Sending full broadcast message")
			message.PeerHash = p2p.FullBroadcastFlag
		case !msg.IsPeer2Peer() && !msg.IsFullBroadcast():
			msgLogger.Debug("Sending broadcast message")
			message.PeerHash = p2p.BroadcastFlag
		case msg.IsPeer2Peer() && 0 == len(message.PeerHash): // directed, with no direction of who to send it to
			msgLogger.Debug("Sending directed message to a random peer")
			message.PeerHash = p2p.RandomPeerFlag
		default:
			msgLogger.Debugf("Sending directed message to: %s", message.PeerHash)
		}
		p2p.BlockFreeChannelSend(f.BroadcastOut, message)
	}

	return nil
}

// Non-blocking return value from channel.
func (f *P2PProxy) Receive() (interfaces.IMsg, error) {
	select {
	case data, ok := <-f.BroadcastIn:
		if ok {
			BroadInCastQueue.Dec()

			switch data.(type) {
			case FactomMessage:
				fmessage := data.(FactomMessage)
				msg, err := msgsupport.UnmarshalMessage(fmessage.Message)

				if err != nil {
					proxyLogger.WithField("receive-error", err).Error()
				} else {
					proxyLogger.WithFields(msg.LogFields()).Debug("Received Message")
				}

				if nil == err {
					msg.SetNetworkOrigin(fmessage.PeerHash)
				}
				f.bytesIn += len(fmessage.Message)
				return msg, err
			default:
				f.logger.Errorf("Garbage on f.BroadcastIn. %+v", data)
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
	p.logger.Info("Starting P2PProxy")
	go p.ManageOutChannel() // Bridges between network format Parcels and factomd messages (incl. addressing to peers)
	go p.ManageInChannel()
}

func (p *P2PProxy) StopProxy() {
	p.logger.Info("Stopped P2PProxy")
}

// manageOutChannel takes messages from the f.broadcastOut channel and sends them to the network.
func (f *P2PProxy) ManageOutChannel() {
	for data := range f.BroadcastOut {
		switch data.(type) {
		case FactomMessage:
			fmessage := data.(FactomMessage)
			// Wrap it in a parcel and send it out channel ToNetwork.
			parcels := p2p.ParcelsForPayload(p2p.CurrentNetwork, fmessage.Message, fmessage.msg)
			for _, parcel := range parcels {
				if parcel.Header.Type != p2p.TypeMessagePart {
					parcel.Header.Type = p2p.TypeMessage
				}
				parcel.Header.TargetPeer = fmessage.PeerHash
				parcel.Header.AppHash = fmessage.AppHash
				parcel.Header.AppType = fmessage.AppType
				p2p.BlockFreeChannelSend(f.ToNetwork, parcel)
			}
		default:
			f.logger.Errorf("Garbage on f.BrodcastOut. %+v", data)
		}
	}
}

// manageInChannel takes messages from the network and stuffs it in the f.BroadcastIn channel
func (f *P2PProxy) ManageInChannel() {
	for data := range f.FromNetwork {
		switch data.(type) {
		case p2p.Parcel:
			parcel := data.(p2p.Parcel)
			message := FactomMessage{Message: parcel.Payload, PeerHash: parcel.Header.TargetPeer, AppHash: parcel.Header.AppHash, AppType: parcel.Header.AppType}
			removed := p2p.BlockFreeChannelSend(f.BroadcastIn, message)
			BroadInCastQueue.Inc()
			BroadInCastQueue.Add(float64(-1 * removed))
			BroadCastInQueueDrop.Add(float64(removed))
		default:
			f.logger.Errorf("Garbage on f.FromNetwork. %+v", data)
		}
	}
}
