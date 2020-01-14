package leader

import (
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/modules/event"
	"github.com/FactomProject/factomd/pubsub"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/worker"
)

type Pub struct {
	MsgOut pubsub.IPublisher
}

// isolate deps on state package - eventually functions will be relocated
var GetFedServerIndexHash = state.GetFedServerIndexHash

// create and start all publishers
func (p *Pub) Init(nodeName string) {
	// REVIEW: will need to spawn/stop leader thread
	// based on federated set membership
	p.MsgOut = pubsub.PubFactory.Threaded(100).Publish(
		pubsub.GetPath(nodeName, event.Path.LeaderMsgOut),
	)
	go p.MsgOut.Start()
}

const (
	LEADER_ROLE = iota + 1
	FOLLOWER_ROLE
)

type role = int

type Sub struct {
	role
	MsgInput       *pubsub.SubChannel
	MovedToHeight  *pubsub.SubChannel
	BalanceChanged *pubsub.SubChannel
	DBlockCreated  *pubsub.SubChannel
	EomTicker      *pubsub.SubChannel
	LeaderConfig   *pubsub.SubChannel
	AuthoritySet   *pubsub.SubChannel
}

func (*Sub) mkChan() *pubsub.SubChannel {
	return pubsub.SubFactory.Channel(1000) // FIXME: should calibrate channel depths
}

// Create all subscribers
func (s *Sub) Init() {
	s.MovedToHeight = s.mkChan()
	s.MsgInput = s.mkChan()
	s.BalanceChanged = s.mkChan()
	s.DBlockCreated = s.mkChan()
	s.EomTicker = s.mkChan()
	s.LeaderConfig = s.mkChan()
	s.AuthoritySet = s.mkChan()
}

// start subscriptions
func (s *Sub) Start(nodeName string) {
	s.LeaderConfig.Subscribe(pubsub.GetPath(nodeName, event.Path.LeaderConfig))
	s.AuthoritySet.Subscribe(pubsub.GetPath(nodeName, event.Path.AuthoritySet))
	s.SetLeaderMode(nodeName)
}

// start listening to subscriptions for leader duties
func (s *Sub) SetLeaderMode(nodeName string) {
	if s.role == LEADER_ROLE {
		return
	}
	s.role = LEADER_ROLE
	s.MsgInput.Subscribe(pubsub.GetPath(nodeName, "bmv", "rest"))
	s.MovedToHeight.Subscribe(pubsub.GetPath(nodeName, event.Path.Seq))
	s.DBlockCreated.Subscribe(pubsub.GetPath(nodeName, event.Path.Directory))
	s.BalanceChanged.Subscribe(pubsub.GetPath(nodeName, event.Path.Bank))
}

// stop subscribers that we do not need as a follower
func (s *Sub) SetFollowerMode() {
	if s.role == FOLLOWER_ROLE {
		return
	}
	s.role = FOLLOWER_ROLE
	s.MsgInput.Unsubscribe()
	s.MovedToHeight.Unsubscribe()
	s.BalanceChanged.Unsubscribe()
	s.DBlockCreated.Unsubscribe()
	s.EomTicker.Unsubscribe()
}

type Events struct {
	Config              *event.LeaderConfig //
	*event.DBHT                             // from move-to-ht
	*event.Balance                          // REVIEW: does this relate to a specific VM
	*event.Directory                        //
	*event.Ack                              // record of last sent ack by leader
	*event.AuthoritySet                     //
}

func (l *Leader) Start(w *worker.Thread) {
	w.Spawn("LeaderThread", func(w *worker.Thread) {
		w.OnReady(func() {
			l.Sub.Start(l.Config.NodeName)
		})
		w.OnRun(l.Run)
		w.OnExit(func() {
			close(l.exit)
			l.Pub.MsgOut.Close()
		})
		l.Sub.Init()
		l.Pub.Init(l.Config.NodeName)
	})
}

func (l *Leader) processMin() (ok bool) {
	go func() {
		time.Sleep(time.Second * time.Duration(l.Config.BlocktimeInSeconds/10))
		l.ticker <- true
	}()

	for {
		select {
		case v := <-l.Sub.LeaderConfig.Updates:
			l.Config = v.(*event.LeaderConfig)
			// TODO: check for change of identity
			return false
		case v := <-l.Sub.AuthoritySet.Updates:
			l.Events.AuthoritySet = v.(*event.AuthoritySet)
			// TODO: check for change of authority
			return false
		case v := <-l.MsgInput.Updates:
			m := v.(interfaces.IMsg)
			// TODO: if message cannot be ack'd send to Dependent Holding
			if constants.NeedsAck(m.Type()) {
				log.LogMessage(logfile, "msgIn ", m)
				l.sendAck(m)
			}
		case <-l.ticker:
			log.LogPrintf(logfile, "Ticker:")
			return true
		case <-l.exit:
			return false
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

// get latest AuthoritySet event data
// and compare w/ leader config
func (l *Leader) currentAuthority() (isLeader bool, index int) {
	evt := l.Events.AuthoritySet

readLatestAuthSet:
	for {
		select {
		case v := <-l.Sub.AuthoritySet.Updates:
			{
				evt = v.(*event.AuthoritySet)
			}
		default:
			{
				l.Events.AuthoritySet = evt
				break readLatestAuthSet
			}
		}
	}

	return GetFedServerIndexHash(l.Events.AuthoritySet.FedServers, l.Config.IdentityChainID)
}

func (l *Leader) WaitForAuthority() (isLeader bool) {
	log.LogPrintf(logfile, "WaitForAuthority %v ", l.Events.AuthoritySet.LeaderHeight)

	defer func() {
		if isLeader {
			l.Sub.SetLeaderMode(l.Config.NodeName)
			log.LogPrintf(logfile, "GotAuthority %v ", l.Events.AuthoritySet.LeaderHeight)
		}
	}()

	for { // wait for AuthoritySetChange
		select {
		case v := <-l.Sub.LeaderConfig.Updates:
			l.Config = v.(*event.LeaderConfig)
		case v := <-l.Sub.AuthoritySet.Updates:
			l.Events.AuthoritySet = v.(*event.AuthoritySet)
			if isAuthority, index := l.currentAuthority(); isAuthority {
				l.VMIndex = index
				return true
			}
		case <-l.exit:
			return false
		}
	}
}

func (l *Leader) Run() {
	// TODO: wait until after boot height
	// ignore these events during DB loading
	l.waitForNextMinute()

	for { //blockLoop

		l.WaitForAuthority()
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
				if !l.processMin() {
					break minLoop
				}
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
}
