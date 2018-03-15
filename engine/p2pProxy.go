// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bufio"
	"fmt"
	"os"
	"time"

	// "github.com/FactomProject/factomd/common/constants"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages/msgsupport"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/p2p"

	log "github.com/sirupsen/logrus"
)

var _ = fmt.Print

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

	logFile   os.File
	logWriter bufio.Writer
	debugMode int
	logging   chan interface{} // NODE_TALK_FIX
	NumPeers  int
	bytesOut  int // bandwidth used by applicaiton without netowrk fan out
	bytesIn   int // bandwidth recieved by application from network
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
	f.logging = make(chan interface{}, p2p.StandardChannelSize)

	return f
}
func (f *P2PProxy) SetDebugMode(netdebug int) {
	f.debugMode = netdebug
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
		//log.Println("ERROR on Send: ", err)
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
	if msg.IsPeer2Peer() && 1 < f.debugMode {
		log.Printf("%s Sending directed to: %s message: %+v\n", time.Now().String(), message.PeerHash, msg.String())
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
				msg, err := msgsupport.UnmarshalMessage(fmessage.Message)

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
	if 1 < p.debugMode {
		go p.ManageLogging()
	}
	go p.ManageOutChannel() // Bridges between network format Parcels and factomd messages (incl. addressing to peers)
	go p.ManageInChannel()
}

// NODE_TALK_FIX
func (p *P2PProxy) stopProxy() {
	if 0 < p.debugMode {
		p2p.BlockFreeChannelSend(p.logging, "stop")
	}
}

type MessageLog struct {
	Hash     string // string(GetMsgHash().Bytes())
	Received bool   // true if logging a recieved message, false if sending
	Time     int64
	Target   string // the id of the targetted node (value may only have local meaning)
	Mtype    byte   /// message type (types defined in constants.go)
}

func (e *MessageLog) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *MessageLog) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *MessageLog) String() string {
	str, _ := e.JSONString()
	return str
}

func (p *P2PProxy) logMessage(msg interfaces.IMsg, received bool) {
	if 2 < p.debugMode {
		// if constants.DBSTATE_MSG == msg.Type() {
		// fmt.Printf("AppMsgLogging: \n Type: %s \n Network Origin: %s \n Message: %s", msg.Type(), msg.GetNetworkOrigin(), msg.String())
		// }
		hash := fmt.Sprintf("%x", msg.GetMsgHash().Bytes())
		time := time.Now().Unix()
		ml := MessageLog{Hash: hash, Received: received, Time: time, Mtype: msg.Type(), Target: msg.GetNetworkOrigin()}
		p2p.BlockFreeChannelSend(p.logging, ml)
	}
}

func (p *P2PProxy) ManageLogging() {
	fmt.Printf("setting up message logging")
	file, err := os.OpenFile("message_log.csv", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0660)
	p.logFile = *file
	if nil != err {
		fmt.Printf("Unable to open logging file. %v", err)
		panic("unable to open logging file")
	}
	writer := bufio.NewWriter(&p.logFile)
	p.logWriter = *writer
	start := time.Now()
	for {
		item := <-p.logging
		switch item.(type) {
		case MessageLog:
			message := item.(MessageLog)
			elapsedMinutes := int(time.Since(start).Minutes())
			line := fmt.Sprintf("%d, %s, %t, %d, %s, %d\n", message.Mtype, message.Hash, message.Received, message.Time, message.Target, elapsedMinutes)
			_, err := p.logWriter.Write([]byte(line))
			if nil != err {
				fmt.Printf("Error writing to logging file. %v", err)
				panic("Error writing to logging file")
			}
		case string:
			message := item.(string)
			if "stop" == message {
				return
			}
		default:
			fmt.Printf("Garbage on p.logging. %+v", item)
			break
		}
	}
	p.logWriter.Flush()
	defer p.logFile.Close()
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

func (f *P2PProxy) PeriodicStatusReport(fnodes []*FactomNode) {
	time.Sleep(p2p.NetworkStatusInterval) // wait for things to spin up
	for {
		time.Sleep(p2p.NetworkStatusInterval)
		f.InstantaneousStatusReport(fnodes)
	}
}

func (f *P2PProxy) InstantaneousStatusReport(fnodes []*FactomNode) {
	fmt.Println("\n\n\n")
	fmt.Println("-------------------------------------------------------------------------------")
	fmt.Println(" Periodic Status Report")
	fmt.Println("-------------------------------------------------------------------------------")
	for _, f := range fnodes {
		f.State.Status = 1
	}
	time.Sleep(100 * time.Millisecond)
	for _, f := range fnodes {
		fmt.Printf("%s \n\n", f.State.ShortString())
	}
	now := time.Now().Format("01/02/2006 15:04:05")
	listenTo := 0
	if listenTo >= 0 && listenTo < len(fnodes) {
		fmt.Printf("%s:\n", now)
		fmt.Printf("      InMsgQueue             %d\n", fnodes[listenTo].State.InMsgQueue().Length())
		fmt.Printf("      AckQueue               %d\n", len(fnodes[listenTo].State.AckQueue()))
		fmt.Printf("      MsgQueue               %d\n", len(fnodes[listenTo].State.MsgQueue()))
		fmt.Printf("      TimerMsgQueue          %d\n", len(fnodes[listenTo].State.TimerMsgQueue()))
		fmt.Printf("      NetworkOutMsgQueue     %d\n", fnodes[listenTo].State.NetworkOutMsgQueue().Length())
		fmt.Printf("      NetworkInvalidMsgQueue %d\n", len(fnodes[listenTo].State.NetworkInvalidMsgQueue()))
		fmt.Printf("      HoldingQueue           %d\n", len(fnodes[listenTo].State.Holding))
	}
	fmt.Printf("      ToNetwork Queue:       %d\n", len(f.ToNetwork))
	fmt.Printf("      FromNetwork Queue:     %d\n", len(f.FromNetwork))
	fmt.Printf("      BroadcastOut Queue:    %d\n", len(f.BroadcastOut))
	fmt.Printf("      BroadcastIn Queue:     %d\n", len(f.BroadcastIn))
	fmt.Printf("      Weight:                %d\n", f.NumPeers)
	fmt.Println("-------------------------------------------------------------------------------")
	fmt.Println("-------------------------------------------------------------------------------")
}
