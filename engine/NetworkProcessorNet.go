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
	"github.com/FactomProject/factomd/globals"
	"github.com/FactomProject/factomd/log"
	"os"
)

var _ = log.Printf
var _ = fmt.Print

func NetworkProcessorNet(fnode *FactomNode) {
	go Peers(fnode)
	go NetworkOutputs(fnode)
	go InvalidOutputs(fnode)
}

var myfile2 *os.File

func Peers(fnode *FactomNode) {
	debugExec := (fnode.State.FactomNodeName == "xFNode0" || fnode.State.FactomNodeName == "FNode0")

	cnt := 0
	saltReplayFilterOn := true

	crossBootIgnore := func(amsg interfaces.IMsg) bool {
		// If we are not syncing, we may ignore some old messages if we are rebooting based on salts
		if saltReplayFilterOn {
			switch amsg.Type() {
			case constants.ACK_MSG:
				ack := amsg.(*messages.Ack)
				replaySalt := fnode.State.CrossReplay.ExistOldSalt(ack.Salt)
				if replaySalt {
					fmt.Println("Found a replay")
				}
				return replaySalt // true means replay and ignore
			}
		}
		return false
	}

	// ackHeight is used in ignoreMsg to determine if we should ignore an acknowledgment
	ackHeight := uint32(0)
	// When syncing from disk/network we want to selectively ignore certain msgs to allow
	// factom to focus on syncing. The following msgs will be ignored:
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
				} else if fnode.State.InMsgQueue().Length() > 4000 {
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
	}

	for {
		if primitives.NewTimestampNow().GetTimeSeconds()-fnode.State.BootTime > int64(constants.CROSSBOOT_SALT_REPLAY_DURATION.Seconds()) {
			saltReplayFilterOn = false
		}

		for i := 0; i < 100 && fnode.State.APIQueue().Length() > 0; i++ {
			msg := fnode.State.APIQueue().Dequeue()

			if debugExec {
				logName := globals.FactomNodeName + "_APIQueue_0" + ".txt"
				messages.LogMessage(logName, "", msg)
			}

			if msg != nil {
				if msg == nil {
					continue
				}
				repeatHash := msg.GetRepeatHash()
				if repeatHash == nil {
					fmt.Println("dddd ERROR!", msg.String())
					continue
				}
				cnt++
				msg.SetOrigin(0)

				// Make sure message isn't a FCT transaction in a block
				_, bv := fnode.State.Replay.Valid(constants.BLOCK_REPLAY,
					msg.GetRepeatHash().Fixed(),
					msg.GetTimestamp(),
					fnode.State.GetTimestamp())

				if bv && fnode.State.Replay.IsTSValid_(constants.NETWORK_REPLAY,
					repeatHash.Fixed(),
					msg.GetTimestamp(),
					fnode.State.GetTimestamp()) {
					//fnode.MLog.add2(fnode, false, fnode.State.FactomNodeName, "API", true, msg)
					if fnode.State.InMsgQueue().Length() < 9000 {
						fnode.State.InMsgQueue().Enqueue(msg)
					}
				} else {
					RepeatMsgs.Inc()
				}
			}
		}

		// Put any broadcasts from our peers into our BroadcastIn queue
		for i, peer := range fnode.Peers {
			for j := 0; j < 100; j++ {
				var msg interfaces.IMsg
				var err error

				preReceiveTime := time.Now()

				if !fnode.State.GetNetStateOff() {
					msg, err = peer.Recieve()
				}

				if msg == nil {
					// Recieve is not blocking; nothing to do, we get a nil.
					break
				}

				receiveTime := time.Since(preReceiveTime)
				TotalReceiveTime.Add(float64(receiveTime.Nanoseconds()))

				cnt++

				if fnode.State.MessageTally {
					fnode.State.TallyReceived(int(msg.Type()))
				}

				if err != nil {
					fmt.Println("ERROR recieving message on", fnode.State.FactomNodeName+":", err)
					break
				}

				msg.SetOrigin(i + 1)

				// Make sure message isn't a FCT transaction in a block
				_, bv := fnode.State.Replay.Valid(constants.BLOCK_REPLAY,
					msg.GetRepeatHash().Fixed(),
					msg.GetTimestamp(),
					fnode.State.GetTimestamp())

				if bv && fnode.State.Replay.IsTSValid_(constants.NETWORK_REPLAY,
					msg.GetRepeatHash().Fixed(),
					msg.GetTimestamp(),
					fnode.State.GetTimestamp()) {
					//if state.GetOut() {
					//	fnode.State.Println("In Coming!! ",msg)
					//}
					in := "PeerIn"
					if msg.IsPeer2Peer() {
						in = "P2P In"
					}
					nme := fmt.Sprintf("%s %d", in, i+1)

					fnode.MLog.Add2(fnode, false, peer.GetNameTo(), nme, true, msg)

					//TODO: Log here -- clay
					logName := globals.FactomNodeName + "_InMsgQueue_i.txt"
					messages.LogMessage(logName, "", msg)

					ignore := ignoreMsg(msg)
					if ignore {
						messages.LogMessage(logName, "ignore", msg)
					}
					if !(fnode.State.InMsgQueue().Length() < 9000) {
						messages.LogMessage(logName, "drop", msg)
					}
					if crossBootIgnore(msg) {
						messages.LogMessage(logName, "crossBootIgnore", msg)
					}

					// Ignore messages if there are too many.
					if fnode.State.InMsgQueue().Length() < 9000 && !ignore && !crossBootIgnore(msg) {
						//TODO: Log here -- clay
						messages.LogMessage(logName, "enqueue", msg)
						fnode.State.InMsgQueue().Enqueue(msg)
					}
				} else {
					RepeatMsgs.Inc()
					//fnode.MLog.add2(fnode, false, peer.GetNameTo(), "PeerIn", false, msg)
				}
			}
		}
		if cnt == 0 {
			time.Sleep(50 * time.Millisecond)
		}
		cnt = 0
	}
}

func NetworkOutputs(fnode *FactomNode) {
	for {
		// if len(fnode.State.NetworkOutMsgQueue()) > 500 {
		// 	fmt.Print(fnode.State.GetFactomNodeName(), "-", len(fnode.State.NetworkOutMsgQueue()), " ")
		// }
		//msg := <-fnode.State.NetworkOutMsgQueue()
		msg := fnode.State.NetworkOutMsgQueue().BlockingDequeue()
		logName := globals.FactomNodeName + "_NetworkOutMsgQueue_o" + ".txt"
		messages.LogMessage(logName, "", msg)

		NetworkOutTotalDequeue.Inc()

		// Local Messages are Not broadcast out.  This is mostly the block signature
		// generated by the timer for the leaders which needs to be processed, but replaced
		// by an updated version when the block is ready.
		if !msg.IsLocal() {
			// Don't do a rand int if drop rate is 0
			if fnode.State.GetDropRate() > 0 && rand.Int()%1000 < fnode.State.GetDropRate() {
				//drop the message, rather than processing it normally
				messages.LogMessage(logName, "Drop onPurpose", msg)
			} else {
				if msg.GetRepeatHash() == nil {
					messages.LogMessage(logName, "Drop nilRepeatHash", msg)
					continue
				}

				// TODO: Used to do this only for ACK, really its appropriate for any message generated by this node
				// rather than use	msg.Type() and switch on the locally generated types we will just do it for all
				// it's cheap enough -- famous last words -- Clay
				//_, ok := msg.(*messages.Ack)
				//if ok {
				// Add the ACK to the replay filter as a side effect of calling this test
				fnode.State.Replay.IsTSValid_(
					constants.NETWORK_REPLAY,
					msg.GetRepeatHash().Fixed(),
					msg.GetTimestamp(),
					fnode.State.GetTimestamp())
				//}

				p := msg.GetOrigin() - 1

				if !fnode.State.GetNetStateOff() {
					if msg.IsPeer2Peer() {
						// Must have a Peer to send a message to a peer
						numPeers := len(fnode.Peers)
						if numPeers > 0 {
							if p < 0 {
								p = rand.Int() % len(fnode.Peers) // randomly pick a peer
							}
							fnode.MLog.Add2(fnode, true, fnode.Peers[p].GetNameTo(), "P2P out", true, msg)
							preSendTime := time.Now()
							messages.LogMessage(logName, "Send to peer", msg)
							fnode.Peers[p].Send(msg)
							sendTime := time.Since(preSendTime)
							TotalSendTime.Add(float64(sendTime.Nanoseconds()))
							if fnode.State.MessageTally {
								fnode.State.TallySent(int(msg.Type()))
							}
						} else {
							messages.LogMessage(logName, "Drop noPeers", msg)
						}
					} else {
						for i, peer := range fnode.Peers {
							wt := 1
							if p >= 0 {
								wt = fnode.Peers[p].Weight()
							}
							// Don't resend to the node that sent it to you.
							if i != p || wt > 1 {
								bco := fmt.Sprintf("%s/%d/%d", "BCast", p, i)
								fnode.MLog.Add2(fnode, true, peer.GetNameTo(), bco, true, msg)
								if !fnode.State.GetNetStateOff() {
									preSendTime := time.Now()
									peer.Send(msg)
									sendTime := time.Since(preSendTime)
									TotalSendTime.Add(float64(sendTime.Nanoseconds()))
									if fnode.State.MessageTally {
										fnode.State.TallySent(int(msg.Type()))
									}
								}
							}
						} // for all peers
					}
				} else {
					messages.LogMessage(logName, "Drop networkOff", msg)
				}
			}
		} else {
			messages.LogMessage(logName, "Drop isLocal", msg)
		}
	} // forever ...
}

// Just throw away the trash
func InvalidOutputs(fnode *FactomNode) {
	for {
		time.Sleep(1 * time.Millisecond) // TODO: Isn't this pointless since the channel will block? -- Clay
		invalidMsg := <-fnode.State.NetworkInvalidMsgQueue()
		//fmt.Println(invalidMsg)

		// The following code was giving a demerit for each instance of a message in the NetworkInvalidMsgQueue.
		// However the consensus system is not properly limiting the messages going into this queue to be ones
		//  indicating an attack.  So the demerits are turned off for now.
		if false && len(invalidMsg.GetNetworkOrigin()) > 0 {
			p2pNetwork.AdjustPeerQuality(invalidMsg.GetNetworkOrigin(), -2)
		}
	} // forever
}
