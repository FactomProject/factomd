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

	"strings"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/p2p"
)

var _ = fmt.Print

var ()

type P2PProxy struct {
	// A connection to this node:
	ToName   string
	FromName string
	// Channels that define the connection:
	BroadcastOut chan interface{} // factomMessage ToNetwork from factomd
	BroadcastIn  chan interface{} // factomMessage FromNetwork for Factomd

	ToNetwork   chan interface{} // p2p.Parcel From p2pProxy to the p2p Controller
	FromNetwork chan interface{} // p2p.Parcel Parcels from the network for the application

	logFile   os.File
	logWriter bufio.Writer
	debugMode int
	logging   chan interface{} // NODE_TALK_FIX
	NumPeers  int
	bytesOut  int // bandwidth used by applicaiton without netowrk fan out
	bytesIn   int // bandwidth recieved by application from network

	// If this is true, outbound messages will be sent out
	// via Etcd (as well as sent out over the p2p network normally)
	useEtcd              bool
	useEtcdExclusive     bool
	blockLeaseIdx        uint32
	EtcdManager          interfaces.IEtcdManager
	SuperVerboseMessages bool
}

type factomMessage struct {
	Message  []byte
	PeerHash string
	AppHash  string
	AppType  string
}

func (e *factomMessage) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *factomMessage) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *factomMessage) String() string {
	str, _ := e.JSONString()
	return str
}

var _ interfaces.IPeer = (*P2PProxy)(nil)

func (f *P2PProxy) SetUseEtcd(setVal bool) {
	f.useEtcd = setVal
}

func (f *P2PProxy) SetUseEtcdExclusive(setVal bool) {
	f.useEtcdExclusive = setVal
}

func (f *P2PProxy) UsingEtcd() bool {
	return f.useEtcd
}

func (f *P2PProxy) UsingEtcdExclusive() bool {
	return f.useEtcdExclusive
}

// Here we filter messages by type (only sending ProcessList-able messages)
// Messages are sent into etcd as marshaled byte-slices
func (f *P2PProxy) SendIntoEtcd(msg interfaces.IMsg) error {
	var err error
	if msg.Type() < 16 || (msg.Type() > 21 && msg.Type() < 25) {
		/* Let's ignore these message types:
		MISSING_MSG           // 16
		MISSING_DATA          // 17
		DATA_RESPONSE         // 18
		MISSING_MSG_RESPONSE  //19
		DBSTATE_MSG          // 20
		DBSTATE_MISSING_MSG  // 21

		BOUNCE_MSG      // 25
		BOUNCEREPLY_MSG // 26
		MISSING_ENTRY_BLOCKS //27
		ENTRY_BLOCK_RESPONSE //28
		*/
		if msg.Type() == constants.EOM_MSG {
			eomMsg := msg.(*messages.EOM)
			if eomMsg.DBHeight > f.blockLeaseIdx {
				f.blockLeaseIdx = eomMsg.DBHeight
				f.EtcdManager.NewBlockLease(f.blockLeaseIdx)
			}
		}
		if msg.Type() == constants.ACK_MSG {
			ackMsg := msg.(*messages.Ack)
			if ackMsg.DBHeight > f.blockLeaseIdx {
				f.blockLeaseIdx = ackMsg.DBHeight
				f.EtcdManager.NewBlockLease(f.blockLeaseIdx)
			}
		}
		if msg.Type() == constants.DIRECTORY_BLOCK_SIGNATURE_MSG {
			dbsigMsg := msg.(*messages.DirectoryBlockSignature)
			if dbsigMsg.DBHeight > f.blockLeaseIdx {
				f.blockLeaseIdx = dbsigMsg.DBHeight
				f.EtcdManager.NewBlockLease(f.blockLeaseIdx)
			}
		}

		msgBytes, err := msg.MarshalBinary()
		if err == nil {
			return f.EtcdManager.SendIntoEtcd(msgBytes)
		}
	}
	return err
}

func (f *P2PProxy) Reinitiate() error {
	return f.EtcdManager.Reinitiate()
}

func (f *P2PProxy) NewBlockLease(blockHeight uint32) error {
	return f.EtcdManager.NewBlockLease(blockHeight)
}

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
	f.blockLeaseIdx = 0
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
	if f.UsingEtcd() {
		err := f.SendIntoEtcd(msg)
		if err != nil {
			if f.SuperVerboseMessages {
				if !strings.Contains(err.Error(), "already exists") {
					fmt.Println(err)
				}
			}
			if strings.Contains(err.Error(), "connection is shut down") {
				err := f.EtcdManager.Reinitiate()
				if err != nil {
					fmt.Println("Tried to reinitiate plugin on failed send; err:", err)
				}
			}
		} else {
			// Send was successful (err was nil)
			if f.SuperVerboseMessages {
				fmt.Println("SVM S:", msg.String(), msg.GetHash().String()[:10])
			}
		}
	} //else {

	if f.UsingEtcdExclusive() {
		return nil
	}

	f.logMessage(msg, false) // NODE_TALK_FIX
	data, err := msg.MarshalBinary()
	if err != nil {
		fmt.Println("ERROR on Send: ", err)
		return err
	}
	if f.SuperVerboseMessages {
		fmt.Println("SVM S:", msg.String(), msg.GetHash().String()[:10])
	}
	f.bytesOut += len(data)
	hash := fmt.Sprintf("%x", msg.GetMsgHash().Bytes())
	appType := fmt.Sprintf("%d", msg.Type())
	message := factomMessage{Message: data, PeerHash: msg.GetNetworkOrigin(), AppHash: hash, AppType: appType}
	switch {
	case !msg.IsPeer2Peer():
		if f.UsingEtcd() && msg.Type() > 1 {
			return nil
		}
		message.PeerHash = p2p.BroadcastFlag
		f.trace(message.AppHash, message.AppType, "P2PProxy.Send() - BroadcastFlag", "a")
	case msg.IsPeer2Peer() && 0 == len(message.PeerHash): // directed, with no direction of who to send it to
		message.PeerHash = p2p.RandomPeerFlag
		f.trace(message.AppHash, message.AppType, "P2PProxy.Send() - RandomPeerFlag", "a")
	default:
		f.trace(message.AppHash, message.AppType, "P2PProxy.Send() - Addressed by hash", "a")
	}
	if msg.IsPeer2Peer() && 1 < f.debugMode {
		fmt.Printf("%s Sending directed to: %s message: %+v\n", time.Now().String(), message.PeerHash, msg.String())
	}
	p2p.BlockFreeChannelSend(f.BroadcastOut, message)
	//}

	return nil
}

// Non-blocking return value from channel.
func (f *P2PProxy) Recieve() (interfaces.IMsg, error) {
	select {
	case data, ok := <-f.BroadcastIn:
		if ok {
			BroadInCastQueue.Dec()
			if f.UsingEtcd() {
				dataBytes, areActuallyBytes := data.([]byte)
				if areActuallyBytes {
					msg, err := messages.UnmarshalMessage(dataBytes)
					if f.SuperVerboseMessages {
						if err != nil {
							fmt.Println("SVM err:", err.Error())
						} else {
							fmt.Println("SVM R:", msg.String(), msg.GetHash().String()[:10])
						}
					}
					return msg, err
				}
			}

			if f.UsingEtcdExclusive() {
				return nil, nil
			}

			//else {
			switch data.(type) {
			case factomMessage:
				fmessage := data.(factomMessage)
				f.trace(fmessage.AppHash, fmessage.AppType, "P2PProxy.Recieve()", "N")
				msg, err := messages.UnmarshalMessage(fmessage.Message)
				if f.SuperVerboseMessages {
					if err != nil {
						fmt.Println("SVM err:", err.Error())
					} else {
						fmt.Println("SVM R:", msg.String())
					}
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
			//}
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

func (f *P2PProxy) SweepEtcd() {
	var newMsgBytes []byte
	for {
		newMsgBytes = f.EtcdManager.GetData()
		if newMsgBytes != nil && len(newMsgBytes) > 0 {
			f.BroadcastIn <- newMsgBytes
		} else {
			time.Sleep(time.Second)
		}

	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////////////////////////////////////////////////////////////

func (p *P2PProxy) StartProxy() {
	if p.UsingEtcd() {
		go p.SweepEtcd()
	}
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

type messageLog struct {
	Hash     string // string(GetMsgHash().Bytes())
	Received bool   // true if logging a recieved message, false if sending
	Time     int64
	Target   string // the id of the targetted node (value may only have local meaning)
	Mtype    byte   /// message type (types defined in constants.go)
}

func (e *messageLog) JSONByte() ([]byte, error) {
	return primitives.EncodeJSON(e)
}

func (e *messageLog) JSONString() (string, error) {
	return primitives.EncodeJSONString(e)
}

func (e *messageLog) String() string {
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
		ml := messageLog{Hash: hash, Received: received, Time: time, Mtype: msg.Type(), Target: msg.GetNetworkOrigin()}
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
		case messageLog:
			message := item.(messageLog)
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
		case factomMessage:
			fmessage := data.(factomMessage)
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
			message := factomMessage{Message: parcel.Payload, PeerHash: parcel.Header.TargetPeer, AppHash: parcel.Header.AppHash, AppType: parcel.Header.AppType}
			p2p.BlockFreeChannelSend(f.BroadcastIn, message)
			BroadInCastQueue.Inc()
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
}
