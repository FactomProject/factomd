// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"math/rand"
	"reflect"
	"time"

	"github.com/PaulSnow/factom2d/common/constants"
	"github.com/PaulSnow/factom2d/common/globals"
	"github.com/PaulSnow/factom2d/common/interfaces"
	"github.com/PaulSnow/factom2d/common/messages"
	"github.com/PaulSnow/factom2d/common/primitives"
)

var _ = fmt.Print

func NetworkProcessorNet(fnode *FactomNode) {
	go Peers(fnode)
	go NetworkOutputs(fnode)
	go InvalidOutputs(fnode)
	go MissingData(fnode)
}

func Peers(fnode *FactomNode) {

	// ackHeight is used in ignoreMsg to determine if we should ignore an acknowledgment
	ackHeight := uint32(0)
	// When syncing from disk/network we want to selectively ignore certain msgs to allow
	// Factom to focus on syncing. The following msgs will be ignored:
	//		Acks:
	//				Ignore acks below the ackheight, which is set if we get an ack at a height higher than
	//			  	the ackheight. This is because Acks are for the current block, which we are not at,
	//				but acks also serve as an indicator as to which height the network is on. So we allow
	//				1 ack through to set out leader height.
	//
	//		Commit/Reveals:
	//				These fill up our holding map because we are not getting acks. If we have things in the
	//				holding map, that increases the amount of time it takes to process the holding map, slowing
	//				down our inmsg queue draining.
	//
	//		EOMs:
	//				Only helpful at the latest height
	//
	//		MissingData:
	//				We should fulfill some of these requests, but we should also focus on ourselves while we are syncing.
	//				If our inmsg queue has too many msgs, then don't help others.
	ignoreMsg := func(amsg interfaces.IMsg) bool {
		// Stop uint32 underflow
		if fnode.State.GetTrueLeaderHeight() < 35 {
			return false
		}
		// If we are syncing up, then apply the filter
		if fnode.State.GetHighestCompletedBlk() < fnode.State.GetTrueLeaderHeight()-35 {
			// Discard all commits, reveals, and acks <= the highest ack height we have seen.
			switch amsg.Type() {
			case constants.COMMIT_CHAIN_MSG:
				return true
			case constants.REVEAL_ENTRY_MSG:
				return true
			case constants.COMMIT_ENTRY_MSG:
				return true
			case constants.EOM_MSG:
				return true
			case constants.MISSING_DATA:
				if !fnode.State.DBFinished {
					return true
				}
			case constants.ACK_MSG:
				if amsg.(*messages.Ack).DBHeight <= ackHeight {
					return true
				}
				// Set the highest ack height seen and allow through
				ackHeight = amsg.(*messages.Ack).DBHeight
			}
		}
		return false
	} // func ignoreMsg(){...}

	for {
		now := fnode.State.GetTimestamp()
		cnt := 0

		for i := 0; i < 100 && fnode.State.APIQueue().Length() > 0; i++ {
			msg := fnode.State.APIQueue().Dequeue()

			if msg.GetRepeatHash() == nil || reflect.ValueOf(msg.GetRepeatHash()).IsNil() || msg.GetMsgHash() == nil || reflect.ValueOf(msg.GetMsgHash()).IsNil() { // Do not send pokemon messages
				fnode.State.LogMessage("badEvents", "PokeMon seen on APIQueue", msg)
				continue
			}

			if globals.Params.FullHashesLog {
				primitives.Loghash(msg.GetMsgHash())
				primitives.Loghash(msg.GetHash())
				primitives.Loghash(msg.GetRepeatHash())
			}

			if msg == nil {
				continue
			}
			if msg.GetHash().IsHashNil() {
				fnode.State.LogMessage("badEvents", "Nil hash from APIQueue", msg)
				continue
			}

			// TODO: Is this as intended for 'x' command? -- clay
			if fnode.State.GetNetStateOff() { // drop received message if he is off
				fnode.State.LogMessage("NetworkInputs", "API drop, X'd by simCtrl", msg)
				continue // Toss any inputs from API
			}

			repeatHash := msg.GetRepeatHash()
			if repeatHash == nil || repeatHash.PFixed() == nil {
				fnode.State.LogMessage("NetworkInputs", "API drop, Hash Error", msg)
				fmt.Println("dddd ERROR!", msg.String())
				continue
			}

			cnt++
			msg.SetOrigin(0)
			timestamp := msg.GetTimestamp()
			repeatHashFixed := repeatHash.Fixed()

			// Make sure message isn't a FCT transaction in a block
			_, BRValid := fnode.State.FReplay.Valid(constants.BLOCK_REPLAY, repeatHashFixed, timestamp, now)
			if !BRValid {
				fnode.State.LogMessage("NetworkInputs", "API Drop, BLOCK_REPLAY", msg)
				RepeatMsgs.Inc()
				continue
			}

			// Make sure the message isn't a duplicate
			NRValid := fnode.State.Replay.IsTSValidAndUpdateState(constants.NETWORK_REPLAY, repeatHashFixed, timestamp, now)
			if !NRValid {
				fnode.State.LogMessage("NetworkInputs", "API Drop, NETWORK_REPLAY", msg)
				RepeatMsgs.Inc()
				continue
			}

			if constants.NeedsAck(msg.Type()) {
				// send msg to MMRequest processing to suppress requests for messages we already have
				fnode.State.RecentMessage.NewMsgs <- msg
			}

			//fnode.MLog.add2(fnode, false, fnode.State.FactomNodeName, "API", true, msg)
			sendToExecute(msg, fnode, "from API")

		} // for the api queue read up to 100 messages {...}

		// Put any broadcasts from our peers into our BroadcastIn queue
		for i, peer := range fnode.Peers {
			fromPeer := fmt.Sprintf("peer-%d", i)
			for j := 0; j < 100; j++ {
				var msg interfaces.IMsg
				var err error

				preReceiveTime := time.Now()

				msg, err = peer.Receive()
				if msg == nil {
					// Receive is not blocking; nothing to do, we get a nil.
					break // move to next peer
				}
				msg.SetReceivedTime(preReceiveTime)

				if err != nil {
					fnode.State.LogPrintf("NetworkInputs", "error on receive from %v: %v", peer.GetNameFrom(), err)
					// TODO: Maybe we should check the error type and/or count errors and change status to offline?
					break // move to next peer
				}

				if msg.GetRepeatHash() == nil || reflect.ValueOf(msg.GetRepeatHash()).IsNil() || msg.GetMsgHash() == nil || reflect.ValueOf(msg.GetMsgHash()).IsNil() { // Do not send pokemon messages
					fnode.State.LogMessage("badEvents", fmt.Sprintf("PokeMon seen on Peer %s", peer.GetNameFrom()), msg)
					continue
				}

				if globals.Params.FullHashesLog {
					primitives.Loghash(msg.GetMsgHash())
					primitives.Loghash(msg.GetHash())
					primitives.Loghash(msg.GetRepeatHash())
				}

				if fnode.State.LLeaderHeight < fnode.State.DBHeightAtBoot+2 {
					s := fnode.State
					if s.GetMessageFilterTimestamp() != nil && msg.GetTimestamp().GetTimeMilli() < s.GetMessageFilterTimestamp().GetTimeMilli() {
						fnode.State.LogMessage("NetworkInputs", "Drop, too old", msg)
						continue
					}
				}

				receiveTime := time.Since(preReceiveTime)
				TotalReceiveTime.Add(float64(receiveTime.Nanoseconds()))

				cnt++

				if fnode.State.MessageTally {
					fnode.State.TallyReceived(int(msg.Type())) //TODO: Do we want to count dropped message?
				}

				if msg.GetHash().IsHashNil() {
					fnode.State.LogMessage("badEvents", "Nil hash from Peer", msg)
					continue
				}

				if fnode.State.GetNetStateOff() { // drop received message if he is off
					fnode.State.LogMessage("NetworkInputs", fromPeer+" Drop, X'd by simCtrl", msg)
					continue // Toss any inputs from this peer
				}

				repeatHash := msg.GetRepeatHash()
				if repeatHash == nil {
					fnode.State.LogMessage("NetworkInputs", fromPeer+" Drop, Hash Error", msg)
					continue
				}

				msg.SetOrigin(i + 1) // Origin is 1 based but peer list is zero based.
				hash := repeatHash.Fixed()
				timestamp := msg.GetTimestamp()

				ignore := ignoreMsg(msg)
				if ignore {
					fnode.State.LogMessage("NetworkInputs", fromPeer+" Drop, ignoreMsg()", msg)
					continue
				}

				//_, bv := fnode.State.Replay.Valid(constants.INTERNAL_REPLAY, hash, timestamp, now)
				//if !bv {
				//	fnode.State.LogMessage("NetworkInputs", fromPeer+" Drop, BLOCK_REPLAY", msg)
				//	RepeatMsgs.Inc()
				//	//fnode.MLog.add2(fnode, false, peer.GetNameTo(), "PeerIn", false, msg)
				//	continue
				//}

				regex, _ := fnode.State.GetInputRegEx()

				if regex != nil {
					t := ""
					if mm, ok := msg.(*messages.MissingMsgResponse); ok {
						t = fmt.Sprintf("%7d-:-%d %s", fnode.State.LLeaderHeight, fnode.State.CurrentMinute, mm.MsgResponse.String())
					} else {
						t = fmt.Sprintf("%7d-:-%d %s", fnode.State.LLeaderHeight, fnode.State.CurrentMinute, msg.String())
					}

					if mm, ok := msg.(*messages.MissingMsgResponse); ok {
						if eom, ok := mm.MsgResponse.(*messages.EOB); ok {
							t2 := fmt.Sprintf("%7d-:-%d %s", fnode.State.LLeaderHeight, fnode.State.CurrentMinute, eom.String())
							messageResult := regex.MatchString(t2)
							if messageResult {
								fnode.State.LogMessage("NetworkInputs", "Drop, matched filter Regex", msg)
								continue
							}
						}
					}
					messageResult := regex.MatchString(t)
					if messageResult {
						fnode.State.LogMessage("NetworkInputs", "Drop, matched filter Regex", msg)
						continue
					}
				}

				//if state.GetOut() {
				//	fnode.State.Println("In Coming!! ",msg)
				//}

				var in string
				if msg.IsPeer2Peer() {
					in = "P2P In"
				} else {
					in = "PeerIn"
				}
				fnode.MLog.Add2(fnode, false, peer.GetNameTo(), fmt.Sprintf("%s %d", in, i+1), true, msg)

				// don't resend peer to peer messages or responses
				if constants.NormallyPeer2Peer(msg.Type()) {
					msg.SetNoResend(true)
				}
				// check if any P2P msg types slip by
				if msg.IsPeer2Peer() && !msg.GetNoResend() {
					fnode.State.LogMessage("NetworkInputs", "unmarked P2P msg", msg)
					msg.SetNoResend(true)
				}

				// This should be the last check before sendtoexecute because it adds the message to the replay
				// as a side effect
				rv := fnode.State.Replay.IsTSValidAndUpdateState(constants.NETWORK_REPLAY, hash, timestamp, now)
				if !rv {
					fnode.State.LogMessage("NetworkInputs", fromPeer+" Drop, NETWORK_REPLAY", msg)
					RepeatMsgs.Inc()
					//fnode.MLog.add2(fnode, false, peer.GetNameTo(), "PeerIn", false, msg)
					continue
				}
				msg.SetNetwork(true)
				sendToExecute(msg, fnode, fromPeer)
			} // For a peer read up to 100 messages {...}
		} // for each peer {...}
		if cnt == 0 {
			time.Sleep(50 * time.Millisecond) // handled no message, sleep a bit
		}
	} // forever {...}
}

func sendToExecute(msg interfaces.IMsg, fnode *FactomNode, source string) {
	t := msg.Type()
	switch t {
	case constants.MISSING_MSG:
		fnode.State.LogMessage("mmr_response", fmt.Sprintf("%s, enqueue %d", source, len(fnode.State.MissingMessageResponseHandler.MissingMsgRequests)), msg)
		fnode.State.MissingMessageResponseHandler.NotifyPeerMissingMsg(msg)

	case constants.COMMIT_CHAIN_MSG:
		fnode.State.ChainCommits.Add(msg) // keep last 100 chain commits
		Q1(fnode, source, msg)            // send it fast track
		reveal := fnode.State.Reveals.Get(msg.GetHash().Fixed())
		if reveal != nil {
			Q1(fnode, source, reveal) // if we have it send it fast track
			// it will still arrive from thr slow track but that is ok.
		}

	case constants.REVEAL_ENTRY_MSG:
		// if this is a chain commit reveal send it fast track to allow processing of dependant reveals
		if fnode.State.ChainCommits.Get(msg.GetHash().Fixed()) != nil {
			Q1(fnode, source, msg) // fast track chain reveals
		} else {
			Q1(fnode, source, msg) // all other reveals are slow track
			fnode.State.Reveals.Add(msg)
		}

	case constants.COMMIT_ENTRY_MSG:
		Q1(fnode, source, msg) // slow track

	case constants.MISSING_DATA:
		Q1(fnode, source, msg) // separated missing data queue

	default:
		//todo: Probably should send EOM/DBSig and their ACKs on a faster yet track
		// in general this makes ACKs more likely to arrive first.
		Q1(fnode, source, msg) // fast track
	}

	if constants.NeedsAck(msg.Type()) {
		// send msg to MMRequest processing to suppress requests for messages we already have
		fnode.State.RecentMessage.NewMsgs <- msg
	}
}

func Q1(fnode *FactomNode, source string, msg interfaces.IMsg) {
	fnode.State.LogMessage("NetworkInputs", source+", enqueue", msg)
	fnode.State.LogMessage("InMsgQueue", source+", enqueue", msg)
	fnode.State.InMsgQueue().Enqueue(msg)
}

func Q2(fnode *FactomNode, source string, msg interfaces.IMsg) {
	fnode.State.LogMessage("NetworkInputs", source+", enqueue2", msg)
	fnode.State.LogMessage("InMsgQueue2", source+", enqueue2", msg)
	fnode.State.InMsgQueue2().Enqueue(msg)
}

func DataQ(fnode *FactomNode, source string, msg interfaces.IMsg) {
	q := fnode.State.DataMsgQueue()
	fnode.State.LogMessage("DataQueue", fmt.Sprintf(source+", enqueue %v", len(q)), msg)
	q <- msg
}

func NetworkOutputs(fnode *FactomNode) {
	for {
		// if len(fnode.State.NetworkOutMsgQueue()) > 500 {
		// 	fmt.Print(fnode.State.GetFactomNodeName(), "-", len(fnode.State.NetworkOutMsgQueue()), " ")
		// }
		//msg := <-fnode.State.NetworkOutMsgQueue()
		msg := fnode.State.NetworkOutMsgQueue().BlockingDequeue()

		NetworkOutTotalDequeue.Inc()
		fnode.State.LogMessage("NetworkOutputs", "Dequeue", msg)

		// Local Messages are Not broadcast out.  This is mostly the block signature
		// generated by the timer for the leaders which needs to be processed, but replaced
		// by an updated version when the block is ready.
		if msg.IsLocal() {
			// todo: Should be a dead case. Add tracking code to see if it ever happens -- clay
			fnode.State.LogMessage("NetworkOutputs", "Drop, local", msg)
			continue
		}
		if msg.GetRepeatHash() == nil {
			fnode.State.LogMessage("NetworkOutputs", "Drop, no repeat hash", msg)
			continue
		}

		regex, _ := fnode.State.GetOutputRegEx()
		if regex != nil {
			t := ""
			if mm, ok := msg.(*messages.MissingMsgResponse); ok {
				t = fmt.Sprintf("%7d-:-%d %s", fnode.State.LLeaderHeight, fnode.State.CurrentMinute, mm.MsgResponse.String())
			} else {
				t = fmt.Sprintf("%7d-:-%d %s", fnode.State.LLeaderHeight, fnode.State.CurrentMinute, msg.String())
			}

			if mm, ok := msg.(*messages.MissingMsgResponse); ok {
				if eom, ok := mm.MsgResponse.(*messages.EOB); ok {
					t2 := fmt.Sprintf("%7d-:-%d %s", fnode.State.LLeaderHeight, fnode.State.CurrentMinute, eom.String())
					messageResult := regex.MatchString(t2)
					if messageResult {
						fnode.State.LogMessage("NetworkOutputs", "Drop, matched filter Regex", msg)
						continue
					}
				}
			}
			messageResult := regex.MatchString(t)
			if messageResult {
				//fmt.Println("Found it!", t)
				fnode.State.LogMessage("NetworkOutputs", "Drop, matched filter Regex", msg)
				continue
			}
		}

		//_, ok := msg.(*messages.Ack)
		//if ok {
		//// We don't care about the result, but we do want to log that we have
		//// seen this message before, because we might have generated the message
		//// ourselves.
		//	// Add the ack to our replay filter
		//	fnode.State.Replay.IsTSValidAndUpdateState(
		//		constants.NETWORK_REPLAY,
		//		msg.GetRepeatHash().Fixed(),
		//		msg.GetTimestamp(),
		//		fnode.State.GetTimestamp())
		//}

		p := msg.GetOrigin() - 1 // Origin is one based but peer list is zero based.

		if msg.IsPeer2Peer() {
			// Must have a Peer to send a message to a peer
			if len(fnode.Peers) > 0 {
				if p < 0 {
					fnode.P2PIndex = (fnode.P2PIndex + 1) % len(fnode.Peers)
					p = rand.Int() % len(fnode.Peers)
				}
				peer := fnode.Peers[p]
				fnode.MLog.Add2(fnode, true, peer.GetNameTo(), "P2P out", true, msg)
				if !fnode.State.GetNetStateOff() { // don't Send p2p messages if he is OFF
					// Don't do a rand int if drop rate is 0
					if fnode.State.GetDropRate() > 0 && rand.Int()%1000 < fnode.State.GetDropRate() {
						//drop the message, rather than processing it normally

						fnode.State.LogMessage("NetworkOutputs", "Drop, simCtrl", msg)
					} else {
						preSendTime := time.Now()
						fnode.State.LogMessage("NetworkOutputs", "Send P2P "+peer.GetNameTo(), msg)
						peer.Send(msg)
						sendTime := time.Since(preSendTime)
						TotalSendTime.Add(float64(sendTime.Nanoseconds()))
						if fnode.State.MessageTally {
							fnode.State.TallySent(int(msg.Type()))
						}
					}
				} else {

					fnode.State.LogMessage("NetworkOutputs", "Drop, simCtrl X", msg)
				}
			} else {
				fnode.State.LogMessage("NetworkOutputs", "Drop, no peers", msg)
			}
		} else {
			fnode.State.LogMessage("NetworkOutputs", "Send broadcast", msg)
			for i, peer := range fnode.Peers {
				wt := 1
				if p >= 0 {
					wt = fnode.Peers[p].Weight()
				}
				// Don't resend to the node that sent it to you.
				if i != p || wt > 1 {
					bco := fmt.Sprintf("%s/%d/%d", "BCast", p, i)
					fnode.MLog.Add2(fnode, true, peer.GetNameTo(), bco, true, msg)
					if !fnode.State.GetNetStateOff() { // Don't send him broadcast message if he is off
						if fnode.State.GetDropRate() > 0 && rand.Int()%1000 < fnode.State.GetDropRate() && !msg.IsFullBroadcast() {
							//drop the message, rather than processing it normally

							fnode.State.LogMessage("NetworkOutputs", "Drop, simCtrl", msg)
						} else {
							preSendTime := time.Now()
							peer.Send(msg)
							sendTime := time.Since(preSendTime)
							TotalSendTime.Add(float64(sendTime.Nanoseconds()))
							if fnode.State.MessageTally {
								fnode.State.TallySent(int(msg.Type()))
							}
						}
					}
				}
			}
		}
	}
}

// Just throw away the trash
func InvalidOutputs(fnode *FactomNode) {
	for {
		time.Sleep(1 * time.Millisecond)
		_ = <-fnode.State.NetworkInvalidMsgQueue()
		//fmt.Println(invalidMsg)

		// The following code was giving a demerit for each instance of a message in the NetworkInvalidMsgQueue.
		// However the consensus system is not properly limiting the messages going into this queue to be ones
		//  indicating an attack.  So the demerits are turned off for now.
		// if len(invalidMsg.GetNetworkOrigin()) > 0 {
		// 	p2pNetwork.AdjustPeerQuality(invalidMsg.GetNetworkOrigin(), -2)
		// }
	}
}

// Handle requests for missing data
func MissingData(fnode *FactomNode) {
	q := fnode.State.DataMsgQueue()
	for {
		select {
		case msg := <-q:
			fnode.State.LogMessage("DataQueue", fmt.Sprintf("dequeue %v", len(q)), msg)
			msg.(*messages.MissingData).SendResponse(fnode.State)
		}
	}
}
