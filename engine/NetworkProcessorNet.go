// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/log"
)

var _ = log.Printf
var _ = fmt.Print

func NetworkProcessorNet(fnode *FactomNode) {
	go Peers(fnode)
	go NetworkOutputs(fnode)
	go InvalidOutputs(fnode)
}

func Peers(fnode *FactomNode) {
	saltReplayFilterOn := true

	crossBootIgnore := func(amsg interfaces.IMsg) bool {
		// If we are not syncing, we may ignore some old messages if we are rebooting based on salts
		if saltReplayFilterOn {
			//var ack *messages.Ack
			//switch amsg.Type() {
			//case constants.MISSING_MSG_RESPONSE:
			//	mmrsp := amsg.(*messages.MissingMsgResponse)
			//	if mmrsp.Ack == nil {
			//		return false
			//	}
			//	ack = mmrsp.Ack.(*messages.Ack)
			//case constants.ACK_MSG:
			//	ack = amsg.(*messages.Ack)
			//case constants.DIRECTORY_BLOCK_SIGNATURE_MSG:
			//	dbs := amsg.(*messages.DirectoryBlockSignature)
			//	if dbs.Ack == nil {
			//		return false
			//	}
			//	ack = dbs.Ack.(*messages.Ack)
			//}

			if amsg.Type() == constants.ACK_MSG && amsg != nil {
				ack := amsg.(*messages.Ack)
				if replaySalt := fnode.State.CrossReplay.ExistOldSalt(ack.Salt); replaySalt {
					return true
				}
			}

		}

		return false
	}

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
				} else if fnode.State.InMsgQueue().Length() > constants.INMSGQUEUE_HIGH {
					// If > 4000, we won't get to this in time anyway. Just drop it since we are behind
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
		if primitives.NewTimestampNow().GetTimeSeconds()-fnode.State.BootTime > int64(constants.CROSSBOOT_SALT_REPLAY_DURATION.Seconds()) {
			saltReplayFilterOn = false
		}
		cnt := 0
		now := fnode.State.GetTimestamp()

		for i := 0; i < 100 && fnode.State.APIQueue().Length() > 0; i++ {
			msg := fnode.State.APIQueue().Dequeue()

			if msg == nil {
				continue
			}
			if msg.GetHash().IsHashNil() {
				fnode.State.LogMessage("badMsgs", "Nil hash from APIQueue", msg)
				continue
			}

			// TODO: Is this as intended for 'x' command? -- clay
			if fnode.State.GetNetStateOff() { // drop received message if he is off
				fnode.State.LogMessage("NetworkInputs", "API drop, X'd by simCtrl", msg)
				continue // Toss any inputs from API
			}

			if fnode.State.InMsgQueue().Length() > constants.INMSGQUEUE_HIGH {
				fnode.State.LogMessage("NetworkInputs", "API Drop, Too Full", msg)
				continue
			}

			if fnode.State.GetNetStateOff() {
				fnode.State.LogMessage("NetworkInputs", "API drop, X'd by simCtrl", msg)
				continue
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
				if err != nil {
					fnode.State.LogPrintf("NetworkInputs", "error on receive from %v: %v", peer.GetNameFrom(), err)
					fmt.Println("ERROR receiving message on", fnode.State.FactomNodeName+":", err)
					// TODO: Maybe we should check the error type and/or count errors and change status to offline?
					break // move to next peer
				}

				if fnode.State.LLeaderHeight < fnode.State.DBHeightAtBoot+2 {
					if msg.GetTimestamp().GetTimeMilli() < fnode.State.TimestampAtBoot.GetTimeMilli() {
						fnode.State.LogMessage("NetworkInputs", "drop, too old", msg)
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
					fnode.State.LogMessage("badMsgs", "Nil hash from Peer", msg)
					continue
				}

				if fnode.State.GetNetStateOff() { // drop received message if he is off
					fnode.State.LogMessage("NetworkInputs", fromPeer+" Drop, X'd by simCtrl", msg)
					continue // Toss any inputs from this peer
				}

				if fnode.State.InMsgQueue().Length() > constants.INMSGQUEUE_HIGH {
					fnode.State.LogMessage("NetworkInputs", fromPeer+" Drop, Too Full", msg)
					continue
				}

				repeatHash := msg.GetRepeatHash()
				if repeatHash == nil {
					fnode.State.LogMessage("NetworkInputs", fromPeer+" Drop, Hash Error", msg)
					continue
				}

				msg.SetOrigin(i + 1) // Origin is 1 based but peer list is zero based.
				hash := repeatHash.Fixed()
				timestamp := msg.GetTimestamp()

				tsv := fnode.State.Replay.IsTSValidAndUpdateState(constants.TIME_TEST, hash, timestamp, now)
				if !tsv {
					fnode.State.LogMessage("NetworkInputs", fromPeer+" Drop, TS invalid", msg)
					continue
				}

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

				rv := fnode.State.Replay.IsTSValidAndUpdateState(constants.NETWORK_REPLAY, hash, timestamp, now)
				if !rv {
					fnode.State.LogMessage("NetworkInputs", fromPeer+" Drop, NETWORK_REPLAY", msg)
					RepeatMsgs.Inc()
					//fnode.MLog.add2(fnode, false, peer.GetNameTo(), "PeerIn", false, msg)
					continue
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

				if !crossBootIgnore(msg) {
					sendToExecute(msg, fnode, fromPeer)
				}
			} // For a peer read up to 100 messages {...}
		} // for each peer {...}
		if cnt == 0 {
			time.Sleep(50 * time.Millisecond) // handled no message, sleep a bit
		}
	} // forever {...}
}

var cacheReveals bool = false

func sendToExecute(msg interfaces.IMsg, fnode *FactomNode, source string) {
	t := msg.Type()
	switch t {
	case constants.COMMIT_CHAIN_MSG:
		fnode.State.ChainCommits.Add(msg) // keep last 100 chain commits
		Q1(fnode, source, msg)            // send it fast track
		if cacheReveals {                 // if we are caching reveals then look to see if we already have the matching reveal
			reveal := fnode.State.Reveals.Get(msg.GetHash().Fixed())
			if reveal != nil {
				Q1(fnode, source, reveal) // if we have it send it fast track
				// it will still arive from thr slow track but thats ok.
			}
		}

	case constants.REVEAL_ENTRY_MSG:
		// if this is a chain commit reveal send it fast track to allow processing of dependant reveals
		if fnode.State.ChainCommits.Get(msg.GetHash().Fixed()) != nil {
			Q1(fnode, source, msg)
		} else {
			Q2(fnode, source, msg) // all other reveals are slow track
			if cacheReveals {
				fnode.State.Reveals.Add(msg)
			}
		}

	case constants.COMMIT_ENTRY_MSG:
		Q2(fnode, source, msg) // slow track

	default:
		Q1(fnode, source, msg) // fast track

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
			fnode.State.LogMessage("NetworkOutputs", "drop, local", msg)
			continue
		}
		// Don't do a rand int if drop rate is 0
		if fnode.State.GetDropRate() > 0 && rand.Int()%1000 < fnode.State.GetDropRate() {
			//drop the message, rather than processing it normally

			fnode.State.LogMessage("NetworkOutputs", "drop, simCtrl", msg)
			continue
		}
		if msg.GetRepeatHash() == nil {
			fnode.State.LogMessage("NetworkOutputs", "drop, no repeat hash", msg)
			continue
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
					preSendTime := time.Now()
					fnode.State.LogMessage("NetworkOutputs", "Send P2P "+peer.GetNameTo(), msg)
					peer.Send(msg)
					sendTime := time.Since(preSendTime)
					TotalSendTime.Add(float64(sendTime.Nanoseconds()))
					if fnode.State.MessageTally {
						fnode.State.TallySent(int(msg.Type()))
					}
				} else {

					fnode.State.LogMessage("NetworkOutputs", "drop, simCtrl X", msg)
				}
			} else {
				fnode.State.LogMessage("NetworkOutputs", "drop, no peers", msg)
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
