// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/fnode"
	"github.com/FactomProject/factomd/modules/bmv"
	"github.com/FactomProject/factomd/modules/pubsub"
	"github.com/FactomProject/factomd/modules/worker"
)

var _ = fmt.Print

func NetworkProcessorNet(w *worker.Thread, fnode *fnode.FactomNode) {
	FromPeerToPeer(w, fnode)
	stubs(w, fnode)
	BasicMessageValidation(w, fnode, 2) // 2 validators
	sort(w, fnode)                      // TODO: Replace this service entirely
	w.Run("NetworkOutputs", func() { NetworkOutputs(fnode) })
	w.Run("InvalidOutputs", func() { InvalidOutputs(fnode) })

	// FIXME: replace queue w/ pub/sub for message delivery
	if false {
		w.Run("MissingData", func() { MissingData(fnode) })
	}
}

// TODO: sort should not exist, we should have each module subscribing to the
//		msgs, rather than this one.
func sort(parent *worker.Thread, fnode *fnode.FactomNode) {
	parent.Spawn("MsgSort", func(w *worker.Thread) {

		// Run init conditions. Setup publishers
		sub := pubsub.SubFactory.Channel(50)

		w.OnReady(func() {
			sub.Subscribe(pubsub.GetPath(fnode.State.GetFactomNodeName(), "bmv", "output"))
		})

		w.OnRun(func() {
			for v := range sub.Channel() {
				msg, ok := v.(interfaces.IMsg)
				if !ok {
					continue
				}
				//fmt.Println("sorted ->", msg)
				sortMsg(msg, fnode, "pubsub")
			}
		})

		w.OnExit(func() {
		})

		w.OnComplete(func() {
		})
	})
}

// stubs are things we need to implement
func stubs(parent *worker.Thread, fnode *fnode.FactomNode) {
	parent.Spawn("stubs", func(w *worker.Thread) {
		// Run init conditions. Setup publishers

		w.OnReady(func() {
		})

		w.OnRun(func() {
		})

		w.OnExit(func() {
		})

		w.OnComplete(func() {
		})
	})
}

// TODO: Switch Peers() to this
func FromPeerToPeer(parent *worker.Thread, fnode *fnode.FactomNode) {
	// TODO: All this logic should be removed. All messages should get filtered and sorted by the basic message
	// 		validators.
	// TODO: Construct the proper setup and teardown of this publisher.
	s := fnode.State
	msgPub := pubsub.PubFactory.MsgSplit(100).Publish(pubsub.GetPath(s.GetFactomNodeName(), "bmv", "input"))
	go msgPub.Start()

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
		if s.GetTrueLeaderHeight() < 35 {
			return false
		}
		// If we are syncing up, then apply the filter
		if s.GetHighestCompletedBlk() < s.GetTrueLeaderHeight()-35 {
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
				if !s.DBFinished {
					return true
				} else if s.InMsgQueue().Length() > constants.INMSGQUEUE_HIGH {
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

	parent.Run("FromPeer", func() {
		for {
			// now := fnode.State.GetTimestamp()
			cnt := 0

			for i := 0; i < 100 && s.APIQueue().Length() > 0; i++ {
				msg := s.APIQueue().Dequeue()

				if globals.Params.FullHashesLog {
					primitives.Loghash(msg.GetMsgHash())
					primitives.Loghash(msg.GetHash())
					primitives.Loghash(msg.GetRepeatHash())
				}

				if msg == nil {
					continue
				}

				cnt++
				msg.SetOrigin(0)

				if constants.NeedsAck(msg.Type()) {
					// send msg to MMRequest processing to suppress requests for messages we already have
					s.RecentMessage.NewMsgs <- msg
				}

				//fnode.MLog.add2(fnode, false, fnode.State.FactomNodeName, "API", true, msg)
				sendToExecute(msg, fnode, "from API", msgPub)

			} // for the api queue read up to 100 messages {...}

			// Put any broadcasts from our peers into our BroadcastIn queue
			for i, peer := range fnode.Peers {
				fromPeer := fmt.Sprintf("peer-%d", i)
				for j := 0; j < 100; j++ {
					var msg interfaces.IMsg
					var err error

					msg, err = peer.Receive()
					if msg == nil {
						// Read is not blocking; nothing to do, we get a nil.
						break // move to next peer
					}
					if err != nil {
						s.LogPrintf("NetworkInputs", "error on receive from %v: %v", peer.GetNameFrom(), err)
						// TODO: Maybe we should check the error type and/or count errors and change status to offline?
						break // move to next peer
					}

					if globals.Params.FullHashesLog {
						primitives.Loghash(msg.GetMsgHash())
						primitives.Loghash(msg.GetHash())
						primitives.Loghash(msg.GetRepeatHash())
					}

					cnt++

					msg.SetOrigin(i + 1) // Origin is 1 based but peer list is zero based.

					ignore := ignoreMsg(msg)
					if ignore {
						s.LogMessage("NetworkInputs", fromPeer+" Drop, ignoreMsg()", msg)
						continue
					}

					// don't resend peer to peer messages or responses
					if constants.NormallyPeer2Peer(msg.Type()) {
						msg.SetNoResend(true)
					}
					// check if any P2P msg types slip by
					if msg.IsPeer2Peer() && !msg.GetNoResend() {
						s.LogMessage("NetworkInputs", "unmarked P2P msg", msg)
						msg.SetNoResend(true)
					}

					msg.SetNetwork(true)
					sendToExecute(msg, fnode, fromPeer, msgPub)
				} // For a peer read up to 100 messages {...}
			} // for each peer {...}
			if cnt == 0 {
				time.Sleep(50 * time.Millisecond) // handled no message, sleep a bit
			}
		} // forever {...}
	})
}

func BasicMessageValidation(parent *worker.Thread, fnode *fnode.FactomNode, validators int) {
	// grab block time for window size
	blocktime := time.Second * time.Duration(fnode.State.GetDirectoryBlockInSeconds())

	// one replay for all validators
	replay := bmv.NewMsgReplay(6, blocktime)

	for i := 0; i < validators; i++ {
		parent.Spawn(fmt.Sprintf("BMV%d", i), func(w *worker.Thread) {
			ctx, cancel := context.WithCancel(context.Background())
			// w.Name is my parent?
			// Init my name object?
			//w.Init(&parent.Name, "bmv")

			// Run init conditions. Setup publishers
			msgIn := bmv.NewBasicMessageValidator(fnode.State.GetFactomNodeName(), replay)

			w.OnReady(func() {
				// Subscribe to publishers
				msgIn.Subscribe()
			})

			w.OnRun(func() {
				// TODO: Temporary print all messages out of bmv. We need to actually use them...
				//go func() {
				//	sub := pubsub.SubFactory.Channel(100).Subscribe(pubsub.GetPath(fnode.State.GetFactomNodeName(), "bmv", "output"))
				//	for v := range sub.Channel() {
				//		fmt.Println("MESSAGE -> ", v)
				//	}
				//}()

				// do work
				msgIn.Run(ctx)
				cancel() // If run is over, we can end the ctx
			})

			w.OnExit(func() {
				cancel()
			})

			w.OnComplete(func() {
				msgIn.ClosePublishing()
			})
		})
	}
}

// FIXME: publish event rather than using nested state queue
func DataQ(node *fnode.FactomNode, source string, msg interfaces.IMsg) {
	q := node.State.DataMsgQueue()
	node.State.LogMessage("DataQueue", fmt.Sprintf(source+", enqueue %v", len(q)), msg)
	q <- msg
}

func sendToExecute(msg interfaces.IMsg, _ *fnode.FactomNode, _ string, pub pubsub.IPublisher) {
	pub.Write(msg)
}

func sortMsg(msg interfaces.IMsg, fnode *fnode.FactomNode, source string) {
	// TODO: These state updates need to be moved to their modules
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
			Q2(fnode, source, msg) // all other reveals are slow track
			fnode.State.Reveals.Add(msg)
		}

	case constants.COMMIT_ENTRY_MSG:
		Q2(fnode, source, msg) // slow track

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

func Q1(fnode *fnode.FactomNode, source string, msg interfaces.IMsg) {
	fnode.State.LogMessage("NetworkInputs", source+", enqueue", msg)
	fnode.State.LogMessage("InMsgQueue", source+", enqueue", msg)
	fnode.State.InMsgQueue().Enqueue(msg)
}

func Q2(fnode *fnode.FactomNode, source string, msg interfaces.IMsg) {
	fnode.State.LogMessage("NetworkInputs", source+", enqueue2", msg)
	fnode.State.LogMessage("InMsgQueue2", source+", enqueue2", msg)
	fnode.State.InMsgQueue2().Enqueue(msg)
}

func NetworkOutputs(fnode *fnode.FactomNode) {
	for {
		// if len(fnode.State.NetworkOutMsgQueue()) > 500 {
		// 	fmt.Print(fnode.State.GetFactomNodeName(), "-", len(fnode.State.NetworkOutMsgQueue()), " ")
		// }
		//msg := <-fnode.State.Ne
		//tworkOutMsgQueue()
		s := fnode.State
		msg := s.NetworkOutMsgQueue().Dequeue()

		NetworkOutTotalDequeue.Inc()
		s.LogMessage("NetworkOutputs", "Dequeue", msg)

		// Local Messages are Not broadcast out.  This is mostly the block signature
		// generated by the timer for the leaders which needs to be processed, but replaced
		// by an updated version when the block is ready.
		if msg.IsLocal() {
			// todo: Should be a dead case. Add tracking code to see if it ever happens -- clay
			s.LogMessage("NetworkOutputs", "Drop, local", msg)
			continue
		}
		if msg.GetRepeatHash() == nil {
			s.LogMessage("NetworkOutputs", "Drop, no repeat hash", msg)
			continue
		}

		regex, _ := s.GetOutputRegEx()
		if regex != nil {
			t := ""
			if mm, ok := msg.(*messages.MissingMsgResponse); ok {
				t = fmt.Sprintf("%7d-:-%d %s", s.LLeaderHeight, s.CurrentMinute, mm.MsgResponse.String())
			} else {
				t = fmt.Sprintf("%7d-:-%d %s", s.LLeaderHeight, s.CurrentMinute, msg.String())
			}

			if mm, ok := msg.(*messages.MissingMsgResponse); ok {
				if eom, ok := mm.MsgResponse.(*messages.EOM); ok {
					t2 := fmt.Sprintf("%7d-:-%d %s", s.LLeaderHeight, s.CurrentMinute, eom.String())
					messageResult := regex.MatchString(t2)
					if messageResult {
						s.LogMessage("NetworkOutputs", "Drop, matched filter Regex", msg)
						continue
					}
				}
			}
			messageResult := regex.MatchString(t)
			if messageResult {
				//fmt.Println("Found it!", t)
				s.LogMessage("NetworkOutputs", "Drop, matched filter Regex", msg)
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
					//bco := fmt.Sprintf("%s/%d/%d", "BCast", p, i)
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
func InvalidOutputs(fnode *fnode.FactomNode) {
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
func MissingData(fnode *fnode.FactomNode) {
	// FIXME bind pub/sub event handlers
	q := fnode.State.DataMsgQueue()
	for {
		select {
		case msg := <-q:
			fnode.State.LogMessage("DataQueue", fmt.Sprintf("dequeue %v", len(q)), msg)
			msg.(*messages.MissingData).SendResponse(fnode.State)
		}
	}
}
