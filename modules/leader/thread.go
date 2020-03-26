package leader

import (
	"time"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/log"
	"github.com/FactomProject/factomd/modules/events"
	"github.com/FactomProject/factomd/modules/pubsub"
	"github.com/FactomProject/factomd/modules/worker"
	"github.com/FactomProject/factomd/state"
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
		pubsub.GetPath(nodeName, events.Path.LeaderMsgOut),
	)
	go p.MsgOut.Start()
}

type role = int

const (
	FederatedRole role = iota + 1
	AuditRole
	FollowerRole
)

var _ = AuditRole // REVIEW: if Audit responsibilities are different from normal Fed then use this

type Sub struct {
	role
	MsgInput       *pubsub.SubChannel
	MovedToHeight  *pubsub.SubChannel
	BalanceChanged *pubsub.SubChannel
	DBlockCreated  *pubsub.SubChannel
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
	s.LeaderConfig = s.mkChan()
	s.AuthoritySet = s.mkChan()
}

// start subscriptions
func (s *Sub) Start(nodeName string) {
	s.LeaderConfig.Subscribe(pubsub.GetPath(nodeName, events.Path.LeaderConfig))
	s.AuthoritySet.Subscribe(pubsub.GetPath(nodeName, events.Path.AuthoritySet))
	{
		s.SetLeaderMode(nodeName) //  create initial subscriptions
	}
}

// start listening to subscriptions for leader duties
func (s *Sub) SetLeaderMode(nodeName string) {
	if s.role == FederatedRole {
		return
	}
	s.role = FederatedRole
	s.MsgInput.Subscribe(pubsub.GetPath(nodeName, "bmv", "rest"))
	s.MovedToHeight.Subscribe(pubsub.GetPath(nodeName, events.Path.Seq))
	s.DBlockCreated.Subscribe(pubsub.GetPath(nodeName, events.Path.Directory))
	s.BalanceChanged.Subscribe(pubsub.GetPath(nodeName, events.Path.Bank))
}

// stop subscribers that we do not need as a follower
func (s *Sub) SetFollowerMode() {
	if s.role == FollowerRole {
		return
	}
	s.role = FollowerRole
	s.MsgInput.Unsubscribe()
	s.MovedToHeight.Unsubscribe()
	s.BalanceChanged.Unsubscribe()
	s.DBlockCreated.Unsubscribe()
}

type Events struct {
	Config               *events.LeaderConfig //
	*events.DBHT                              // from move-to-ht
	*events.Balance                           // REVIEW: does this relate to a specific VM
	*events.Directory                         //
	*events.Ack                               // record of last sent ack by leader
	*events.AuthoritySet                      //
}

func (l *Leader) Start(w *worker.Thread) {
	if !state.EnableLeaderThread {
		panic("LeaderThreadDisabled")
	}

	w.Spawn("LeaderThread", func(w *worker.Thread) {
		w.OnReady(func() {
			l.Sub.Start(l.Config.NodeName)
		})
		w.OnRun(l.Run)
		w.OnExit(func() {
			close(l.exit)
			l.Pub.MsgOut.Close()
		})
		l.Pub.Init(l.Config.NodeName)
		l.Sub.Init()
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
			l.Config = v.(*events.LeaderConfig)
		case v := <-l.MsgInput.Updates:
			m := v.(interfaces.IMsg)
			if constants.NeedsAck(m.Type()) {
				log.LogMessage(l.logfile, "msgIn ", m)
				l.sendAck(m)
			}
		case <-l.ticker:
			log.LogPrintf(l.logfile, "Ticker:")
			return true
		case <-l.exit:
			return false
		}
	}
}

func (l *Leader) waitForNextMinute() (min int, ok bool) {
	for {
		select {
		case v := <-l.MovedToHeight.Updates:
			evt := v.(*events.DBHT)
			log.LogPrintf(l.logfile, "DBHT: %v", evt)

			if evt.Minute == 10 {
				continue
			}
			if l.DBHT.Minute == evt.Minute && l.DBHT.DBHeight == evt.DBHeight {
				continue
			}

			l.DBHT = evt
			return l.DBHT.Minute, true
		case <-l.exit:
			return -1, false
		}
	}
}

// TODO: refactor to only get a single Directory event
func (l *Leader) WaitForDBlockCreated() (ok bool) {
	for { // wait on a new (unique) directory event
		select {
		case v := <-l.Sub.DBlockCreated.Updates:
			evt := v.(*events.Directory)
			if l.Directory != nil && evt.DBHeight == l.Directory.DBHeight {
				log.LogPrintf(l.logfile, "DUP Directory: %v", v)
				continue
			} else {
				log.LogPrintf(l.logfile, "Directory: %v", v)
			}
			l.Directory = v.(*events.Directory)
			return true
		case <-l.exit:
			return false
		}
	}
}

func (l *Leader) WaitForBalanceChanged() (ok bool) {
	select {
	case v := <-l.Sub.BalanceChanged.Updates:
		l.Balance = v.(*events.Balance)
		log.LogPrintf(l.logfile, "BalChange: %v", v)
		return true
	case <-l.exit:
		return false
	}
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
				evt = v.(*events.AuthoritySet)
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

// wait to become leader (possibly forever for followers)
func (l *Leader) WaitForAuthority() (isLeader bool) {
	// REVIEW: do we need to check block ht?
	log.LogPrintf(l.logfile, "WaitForAuthority %v ", l.Events.AuthoritySet.LeaderHeight)

	defer func() {
		if isLeader {
			l.Sub.SetLeaderMode(l.Config.NodeName)
			log.LogPrintf(l.logfile, "GotAuthority %v ", l.Events.AuthoritySet.LeaderHeight)
		}
	}()

	for {
		select {
		case v := <-l.Sub.LeaderConfig.Updates:
			l.Config = v.(*events.LeaderConfig)
		case v := <-l.Sub.AuthoritySet.Updates:
			l.Events.AuthoritySet = v.(*events.AuthoritySet)
		case <-l.exit:
			return false
		}
		if isAuthority, index := l.currentAuthority(); isAuthority {
			l.VMIndex = index
			return true
		}
	}
}

func (l *Leader) waitForNewBlock() (stillWaiting bool) {
	if min, ok := l.waitForNextMinute(); !ok {
		return false
	} else {
		return min != 0
	}
}

func (l *Leader) Run() {
	// TODO: wait until after boot height
	// ignore these events during DB loading
	l.waitForNextMinute()

blockLoop:
	for { //blockLoop
		ok := worker.RunSteps(
			l.WaitForAuthority,
			l.WaitForBalanceChanged,
			l.WaitForDBlockCreated,
		)
		if !ok {
			break blockLoop
		} else {
			l.sendDBSig()
		}
		log.LogPrintf(l.logfile, "MinLoopStart: %v", true)
	minLoop:
		for {
			ok := worker.RunSteps(
				l.processMin,
				l.sendEOM,
			)
			if !ok {
				break minLoop
			} else {
				l.waitForNewBlock()
			}
		}
		log.LogPrintf(l.logfile, "MinLoopEnd: %v", true)
	}
}
