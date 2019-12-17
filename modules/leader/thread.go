package leader

import (
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/fnode"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/worker"
)

type Pub struct {
	MsgOut pubsub.IPublisher
}

type Sub struct {
	MsgInput       *pubsub.SubChannel
	MovedToHeight  *pubsub.SubChannel
	BalanceChanged *pubsub.SubChannel
	DBlockCreated  *pubsub.SubChannel
	EomTicker      *pubsub.SubChannel
	LeaderConfig   *pubsub.SubChannel
	// TODO: add InternalAuthoritySet listener - probably add an update auth set msg instead of listening to all
	// Eventually also track swapping audits and leaders AddRemoveServer Messages (come in quads)
	// currently sent to election process

}

// block level events
type Events struct {
	Config         *event.LeaderConfig //
	*event.DBHT                        // from move-to-ht
	*event.Balance                     // REVIEW: does this relate to a specific VM
	*event.Directory
	*event.Ack // record of last sent ack by leader
	*event.LeaderConfig
}

func (*Leader) mkChan() *pubsub.SubChannel {
	return pubsub.SubFactory.Channel(1000) // FIXME: should calibrate channel depths
}

func (l *Leader) Start(w *worker.Thread) {

	w.Spawn("LeaderThread", func(w *worker.Thread) {
		w.OnReady(l.Ready)
		w.OnRun(l.Run)
		w.OnExit(l.Exit)

		l.Pub.MsgOut = pubsub.PubFactory.Threaded(100).Publish(
			pubsub.GetPath("FNode0", event.Path.LeaderMsgOut),
		)
		go l.Pub.MsgOut.Start()

		l.Sub.MovedToHeight = l.mkChan()
		l.Sub.MsgInput = l.mkChan()
		l.Sub.BalanceChanged = l.mkChan()
		l.Sub.DBlockCreated = l.mkChan()
		l.Sub.EomTicker = l.mkChan()
		l.Sub.LeaderConfig = l.mkChan()
	})
}

func (l *Leader) Exit() {
	close(l.exit)
	l.Pub.MsgOut.Close()
}

func (l *Leader) Ready() {
	node0 := fnode.Get(0).State.GetFactomNodeName() // get follower name

	// network inputs
	l.Sub.MsgInput.Subscribe(pubsub.GetPath(node0, "bmv", "rest"))

	// internal events
	l.Sub.MovedToHeight.Subscribe(pubsub.GetPath(node0, event.Path.Seq))
	l.Sub.DBlockCreated.Subscribe(pubsub.GetPath(node0, event.Path.Directory))
	l.Sub.BalanceChanged.Subscribe(pubsub.GetPath(node0, event.Path.Bank))
}

func (l *Leader) processMin() {
	go func() {
		time.Sleep(time.Second * time.Duration(l.Config.BlocktimeInSeconds/10))
		l.ticker <- true
	}()

	for {
		select {
		case v := <-l.Sub.LeaderConfig.Updates:
			l.Config = v.(*event.LeaderConfig)
		case v := <-l.MsgInput.Updates:
			m := v.(interfaces.IMsg)
			// TODO: do leader work - actually validate the message by calling into Bank Service
			if constants.NeedsAck(m.Type()) {
				log.LogMessage(logfile, "msgIn ", m)
				l.sendAck(m)
			}
		case <-l.ticker:
			log.LogPrintf(logfile, "Ticker:")
			return
		case <-l.exit:
			return
		}
	}
}

func (l *Leader) waitForNextMinute() int {
	for {
		select {
		case v := <-l.MovedToHeight.Updates:
			evt := v.(*event.DBHT)
			log.LogPrintf(logfile, "DBHT: %v", evt)

			if evt.Minute == 10 {
				continue
			}
			if l.DBHT.Minute == evt.Minute && l.DBHT.DBHeight == evt.DBHeight {
				continue
			}

			l.DBHT = evt
			return l.DBHT.Minute
		case <-l.exit:
			return -1
		}
	}

}

// TODO: refactor to only get a single Directory event
func (l *Leader) WaitForDBlockCreated() {
	for { // wait on a new (unique) directory event
		v := <-l.Sub.DBlockCreated.Updates
		evt := v.(*event.Directory)
		if l.Directory != nil && evt.DBHeight == l.Directory.DBHeight {
			log.LogPrintf(logfile, "DUP Directory: %v", v)
			continue
		} else {
			log.LogPrintf(logfile, "Directory: %v", v)
		}
		l.Directory = v.(*event.Directory)
		return
	}
}

func (l *Leader) WaitForBalanceChanged() {
	v := <-l.Sub.BalanceChanged.Updates
	l.Balance = v.(*event.Balance)
	log.LogPrintf(logfile, "BalChange: %v", v)
}

func (l *Leader) Run() {
	// TODO: wait until after boot height
	// ignore these events during DB loading
	l.waitForNextMinute()

blockLoop:
	for {
		if false { // TODO: deal w/ new auth set
			//case v := <-l.NewAuthoritySet
			// if got a new Auth & no longer leader - break the block look
			l.VMIndex = 0 // KLUDGE hard coded for single leader
			break blockLoop
		}

		l.WaitForBalanceChanged()
		l.WaitForDBlockCreated()
		l.sendDBSig()

		log.LogPrintf(logfile, "MinLoopStart: %v", true)
	minLoop:
		for { // could be counted 1..9 to account for min
			select {
			case <-l.exit:
				return
			default:
				l.processMin()
				l.sendEOM()

				switch min := l.waitForNextMinute(); min {
				case -1: // caught exit
					return
				case 0:
					break minLoop
				default:
					continue minLoop
				}
			}
		}
		log.LogPrintf(logfile, "MinLoopEnd: %v", true)
	}

	// FIXME: block on waiting for an event to become a leader again
}
