package leader

import (
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/fnode"
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
	return pubsub.SubFactory.Channel(50)
}

func (l *Leader) Start(w *worker.Thread) {

	w.Spawn(func(w *worker.Thread) {
		w.Init(&w.Name, "LeaderThread")
		w.OnReady(l.Ready)
		w.OnRun(l.Run)

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

func (l *Leader) Ready() {
	node0 := fnode.Get(0).State.GetFactomNodeName() // get follower name

	{ // temporary hooks for leader thread development
		// KLUDGE publish to Fnode01 bypassing networkOut
		// snoop on valid message inputs
		l.Sub.MsgInput.Subscribe(pubsub.GetPath(node0, "bmv", "rest"))
	}

	// subscribe to internal events
	l.Sub.MovedToHeight.Subscribe(pubsub.GetPath(node0, event.Path.Seq))
	l.Sub.DBlockCreated.Subscribe(pubsub.GetPath(node0, event.Path.Directory))
	l.Sub.BalanceChanged.Subscribe(pubsub.GetPath(node0, event.Path.Bank))
}

func (l *Leader) ProcessMin() {
	go func() {
		time.Sleep(time.Second * 3 / 2) // FIXME add min alignment
		l.ticker <- true
	}()

	for {

		select {
		case v := <-l.Sub.LeaderConfig.Updates:
			l.Config = v.(*event.LeaderConfig)
		case v := <-l.MsgInput.Updates:
			m := v.(interfaces.IMsg)
			// TODO: do leader work - actually validate the message
			if constants.NeedsAck(m.Type()) {
				log.LogMessage("leader.txt", "msgIn ", m)
				l.sendAck(m)
			}
		case <-l.ticker:
			log.LogPrintf("leader.txt", "Ticker:")
			return
		}
	}
}

func (l *Leader) WaitForMoveToHt() int {
	for { // could be counted 0..9 to account for min
		// possibly shut down this leader thread or maybe unsubscribe to events
		select {
		case v := <-l.MovedToHeight.Updates:
			evt := v.(*event.DBHT)
			log.LogPrintf("leader.txt", "DBHT: %v", evt)

			if evt.Minute == 10 {
				continue
			}
			if l.DBHT.Minute == evt.Minute && l.DBHT.DBHeight == evt.DBHeight {
				continue
			}

			l.DBHT = evt
			return l.DBHT.Minute
		}
	}

}

func (l *Leader) Run() {
	// TODO: wait until after boot height
	// ignore these events during DB loading
	l.WaitForMoveToHt()

blockLoop:
	for {

		if false { // TODO: deal w/ new auth set
			//case v := <-l.NewAuthoritySet
			// if got a new Auth & no longer leader - break the block look
			l.VMIndex = 0 // KLUDGE hard coded for single leader
			break blockLoop
		}

		{
			v := <-l.Sub.BalanceChanged.Updates
			l.Balance = v.(*event.Balance)
			log.LogPrintf("leader.txt", "BalChange: %v", v)
		}
		// TODO: refactor to only get a single Directory event
		for { // wait on a new (unique) directory event
			v := <-l.Sub.DBlockCreated.Updates
			evt := v.(*event.Directory)
			if l.Directory != nil && evt.DBHeight == l.Directory.DBHeight {
				log.LogPrintf("leader.txt", "DUP Directory: %v", v)
				continue
			} else {
				log.LogPrintf("leader.txt", "Directory: %v", v)
			}
			l.Directory = v.(*event.Directory)
			break
		}

		l.SendDBSig()

		log.LogPrintf("leader.txt", "MinLoopStart: %v", true)
	minLoop:
		for { // could be counted 1..9 to account for min
			l.ProcessMin()
			l.SendEOM()
			min := l.WaitForMoveToHt()
			if min == 0 {
				break minLoop
			}
		}
		log.LogPrintf("leader.txt", "MinLoopEnd: %v", true)
	}
}
