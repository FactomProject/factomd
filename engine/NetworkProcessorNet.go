// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/FactomProject/factomd/modules/DependentHolding"
	"github.com/FactomProject/factomd/modules/bmv"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/state"

	"github.com/FactomProject/factomd/fnode"
	"github.com/FactomProject/factomd/worker"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/globals"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
)

var _ = fmt.Print

func NetworkProcessorNet(w *worker.Thread, fnode *fnode.FactomNode) {
	//Peers(w, fnode)
	FromPeerToPeer(w, fnode)
	stubs(w, fnode)

	BasicMessageValidation(w, fnode) // create instances of basic message validation
	startDependentHolding(w, fnode)  // create instances of dependent holding

	sort(w, fnode.State) // TODO: Replace this service entirely
	w.Run("NetworkOutputs", func() { NetworkOutputs(fnode) })
	w.Run("InvalidOutputs", func() { InvalidOutputs(fnode.State) })
}

// TODO: sort should not exist, we should have each module subscribing to the
//		msgs, rather than this one.
func sort(parent *worker.Thread, s *state.State) {
	parent.Spawn("MsgSort", func(w *worker.Thread) {

		// Run init conditions. Setup publishers
		sub := pubsub.SubFactory.Channel(50)

		w.OnReady(func() {
			sub.Subscribe(pubsub.GetPath(s.GetFactomNodeName(), "dependentholding", "msgout"))
		})

		w.OnRun(func() {
			for v := range sub.Channel() {
				msg, ok := v.(interfaces.IMsg)
				if !ok {
					continue
				}
				//fmt.Println("sorted ->", msg)
				sortMsg(msg, s, "pubsub")
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
		pub := pubsub.PubFactory.Base().Publish(fnode.State.GetFactomNodeName() + "/blocktime")

		w.OnReady(func() {
		})

		w.OnRun(func() {
		})

		w.OnExit(func() {
			_ = pub.Close()
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
	msgPub := pubsub.PubFactory.MsgSplit(100).Publish(s.GetFactomNodeName() + "/msgs")
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
				sendToExecute(msg, msgPub)

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
					sendToExecute(msg, msgPub)
				} // For a peer read up to 100 messages {...}
			} // for each peer {...}
			if cnt == 0 {
				time.Sleep(50 * time.Millisecond) // handled no message, sleep a bit
			}
		} // forever {...}
	})
}

func BasicMessageValidation(parent *worker.Thread, fnode *fnode.FactomNode) {
	for i := 0; i < 2; i++ { // 2 Basic message validators
		parent.Spawn(fmt.Sprintf("BMV%d", i), func(w *worker.Thread) {
			ctx, cancel := context.WithCancel(context.Background())
			// w.Name is my parent?
			// Init my name object?
			//			w.Init(&parent.Name, "bmv")

			// Run init conditions. Setup publishers
			msgIn := bmv.NewBasicMessageValidator(fnode.State.GetFactomNodeName())

			w.OnReady(func() {
				// Subscribe to publishers
				msgIn.Subscribe()
			})

			w.OnRun(func() {
				// TODO: Temporary print all messages out of bmv. We need to actually use them...
				//go func() {
				//	sub := pubsub.SubFactory.Channel(100).Subscribe(pubsub.GetPath(fnode.State.GetFactomNodeName(), "bmv", "rest"))
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

func startDependentHolding(parent *worker.Thread, fnode *fnode.FactomNode) {
	for i := 0; i < 2; i++ { // 2 Basic message validators
		parent.Spawn(fmt.Sprintf("DH%d", i), func(w *worker.Thread) {
			ctx, cancel := context.WithCancel(context.Background())
			// Run init conditions. Setup publishers
			dependentHolding := DependentHolding.NewDependentHolding(&fnode.Name, i)

			w.OnReady(func() {
				dependentHolding.Publish()
				dependentHolding.Subscribe()
			})

			w.OnRun(func() {
				// do work
				dependentHolding.Run(ctx)
				cancel() // If run is over, we can end the ctx
			})

			w.OnExit(func() {
				cancel()
			})

			w.OnComplete(func() {
				dependentHolding.ClosePublishing()
			})
		})
	}
}

func Peers(w *worker.Thread, fnode *fnode.FactomNode) {
	// TODO: All this logic should be removed. All messages should get filtered and sorted by the basic message
	// 		validators.
	// TODO: Construct the proper setup and teardown of this publisher.
	msgPub := pubsub.PubFactory.MsgSplit(100).Publish(fnode.State.GetFactomNodeName() + "/msgs")
	go msgPub.Start()

	// FIXME: bind to
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

	w.Run("FromAPI", func() {
		for {
			now := fnode.State.GetTimestamp()
			if now.GetTimeSeconds()-fnode.State.BootTime > int64(constants.CROSSBOOT_SALT_REPLAY_DURATION.Seconds()) {
				saltReplayFilterOn = false
			}
			cnt := 0

			for i := 0; i < 100 && fnode.State.APIQueue().Length() > 0; i++ {
				msg := fnode.State.APIQueue().Dequeue()

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
				sendToExecute(msg, msgPub)

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
						// Read is not blocking; nothing to do, we get a nil.
						break // move to next peer
					}
					if err != nil {
						fnode.State.LogPrintf("NetworkInputs", "error on receive from %v: %v", peer.GetNameFrom(), err)
						// TODO: Maybe we should check the error type and/or count errors and change status to offline?
						break // move to next peer
					}

					if globals.Params.FullHashesLog {
						primitives.Loghash(msg.GetMsgHash())
						primitives.Loghash(msg.GetHash())
						primitives.Loghash(msg.GetRepeatHash())
					}

					if fnode.State.LLeaderHeight < fnode.State.DBHeightAtBoot+2 {
						s := fnode.State
						// Allow 20 minute grace period
						if s.GetMessageFilterTimestamp() != nil && msg.GetTimestamp().GetTimeMilli() < s.GetMessageFilterTimestamp().GetTimeMilli() {
							fnode.State.LogMessage("NetworkInputs", "Drop, too old", msg)
							continue
						}
					}

					receiveTime := time.Since(preReceiveTime)
					var _ = receiveTime
					// TotalReceiveTime.Add(float64(receiveTime.Nanoseconds()))

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

					regex, _ := fnode.State.GetInputRegEx()

					if regex != nil {
						t := ""
						if mm, ok := msg.(*messages.MissingMsgResponse); ok {
							t = fmt.Sprintf("%7d-:-%d %s", fnode.State.LLeaderHeight, fnode.State.CurrentMinute, mm.MsgResponse.String())
						} else {
							t = fmt.Sprintf("%7d-:-%d %s", fnode.State.LLeaderHeight, fnode.State.CurrentMinute, msg.String())
						}

						if mm, ok := msg.(*messages.MissingMsgResponse); ok {
							if eom, ok := mm.MsgResponse.(*messages.EOM); ok {
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
					_ = in // KLUDGE golang error?
					if msg.IsPeer2Peer() {
						in = "P2P In"
					} else {
						in = "PeerIn"
					}

					// don't resend peer to peer messages or responses
					if constants.NormallyPeer2Peer(msg.Type()) {
						msg.SetNoResend(true)
					}
					// check if any P2P msg types slip by
					if msg.IsPeer2Peer() && !msg.GetNoResend() {
						fnode.State.LogMessage("NetworkInputs", "unmarked P2P msg", msg)
						msg.SetNoResend(true)
					}

					msg.SetNetwork(true)
					if !crossBootIgnore(msg) {
						sendToExecute(msg, msgPub)
					}
				} // For a peer read up to 100 messages {...}
			} // for each peer {...}
			if cnt == 0 {
				time.Sleep(50 * time.Millisecond) // handled no message, sleep a bit
			}
		} // forever {...}
	})
}

func sendToExecute(msg interfaces.IMsg, pub pubsub.IPublisher) {
	pub.Write(msg)
}

func sortMsg(msg interfaces.IMsg, s *state.State, source string) {
	// TODO: These state updates need to be moved to their modules
	t := msg.Type()
	switch t {
	case constants.MISSING_MSG:
		s.LogMessage("mmr_response", fmt.Sprintf("%s, enqueue %d", source, len(s.MissingMessageResponseHandler.MissingMsgRequests)), msg)
		s.MissingMessageResponseHandler.NotifyPeerMissingMsg(msg)

	case constants.COMMIT_CHAIN_MSG:
		s.ChainCommits.Add(msg) // keep last 100 chain commits
		Q1(s, source, msg)      // send it fast track
		reveal := s.Reveals.Get(msg.GetHash().Fixed())
		if reveal != nil {
			Q1(s, source, reveal) // if we have it send it fast track
			// it will still arrive from thr slow track but that is ok.
		}

	case constants.REVEAL_ENTRY_MSG:
		// if this is a chain commit reveal send it fast track to allow processing of dependant reveals
		if s.ChainCommits.Get(msg.GetHash().Fixed()) != nil {
			Q1(s, source, msg) // fast track chain reveals
		} else {
			Q2(s, source, msg) // all other reveals are slow track
			s.Reveals.Add(msg)
		}

	case constants.COMMIT_ENTRY_MSG:
		Q2(s, source, msg) // slow track

	default:
		//todo: Probably should send EOM/DBSig and their ACKs on a faster yet track
		// in general this makes ACKs more likely to arrive first.
		Q1(s, source, msg) // fast track
	}

	if constants.NeedsAck(msg.Type()) {
		// send msg to MMRequest processing to suppress requests for messages we already have
		s.RecentMessage.NewMsgs <- msg
	}
}

func Q1(s *state.State, source string, msg interfaces.IMsg) {
	s.LogMessage("NetworkInputs", source+", enqueue", msg)
	s.LogMessage("InMsgQueue", source+", enqueue", msg)
	s.InMsgQueue().Enqueue(msg)
}

func Q2(s *state.State, source string, msg interfaces.IMsg) {
	s.LogMessage("NetworkInputs", source+", enqueue2", msg)
	s.LogMessage("InMsgQueue2", source+", enqueue2", msg)
	s.InMsgQueue2().Enqueue(msg)
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
				if !s.GetNetStateOff() { // don't Send p2p messages if he is OFF
					// Don't do a rand int if drop rate is 0
					if s.GetDropRate() > 0 && rand.Int()%1000 < s.GetDropRate() {
						//drop the message, rather than processing it normally

						s.LogMessage("NetworkOutputs", "Drop, simCtrl", msg)
					} else {
						preSendTime := time.Now()
						s.LogMessage("NetworkOutputs", "Send P2P "+peer.GetNameTo(), msg)
						peer.Send(msg)
						sendTime := time.Since(preSendTime)
						TotalSendTime.Add(float64(sendTime.Nanoseconds()))
						if s.MessageTally {
							s.TallySent(int(msg.Type()))
						}
					}
				} else {

					s.LogMessage("NetworkOutputs", "Drop, simCtrl X", msg)
				}
			} else {
				s.LogMessage("NetworkOutputs", "Drop, no peers", msg)
			}
		} else {
			s.LogMessage("NetworkOutputs", "Send broadcast", msg)
			for i, peer := range fnode.Peers {
				wt := 1
				if p >= 0 {
					wt = fnode.Peers[p].Weight()
				}
				// Don't resend to the node that sent it to you.
				if i != p || wt > 1 {
					//bco := fmt.Sprintf("%s/%d/%d", "BCast", p, i)
					if !s.GetNetStateOff() { // Don't send him broadcast message if he is off
						if s.GetDropRate() > 0 && rand.Int()%1000 < s.GetDropRate() && !msg.IsFullBroadcast() {
							//drop the message, rather than processing it normally

							s.LogMessage("NetworkOutputs", "Drop, simCtrl", msg)
						} else {
							preSendTime := time.Now()
							peer.Send(msg)
							sendTime := time.Since(preSendTime)
							TotalSendTime.Add(float64(sendTime.Nanoseconds()))
							if s.MessageTally {
								s.TallySent(int(msg.Type()))
							}
						}
					}
				}
			}
		}
	}
}

// Just throw away the trash
func InvalidOutputs(s *state.State) {
	for {
		time.Sleep(1 * time.Millisecond)
		_ = <-s.NetworkInvalidMsgQueue()
		//fmt.Println(invalidMsg)

		// The following code was giving a demerit for each instance of a message in the NetworkInvalidMsgQueue.
		// However the consensus system is not properly limiting the messages going into this queue to be ones
		//  indicating an attack.  So the demerits are turned off for now.
		// if len(invalidMsg.GetNetworkOrigin()) > 0 {
		// 	p2pNetwork.AdjustPeerQuality(invalidMsg.GetNetworkOrigin(), -2)
		// }
	}
}
