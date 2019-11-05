package publishers

import "time"

// RoundRobin only sends events to 1 subscriber on a round robin basis.
type RoundRobin struct {
	*Threaded
	next int
}

func NewRoundRobinPublisher(buffer int) *RoundRobin {
	p := new(RoundRobin)
	p.Threaded = NewThreadedPublisherPublisher(buffer)

	return p
}

func (p *RoundRobin) Run() {
	for in := range p.inputs { // Run until close
		for len(p.Subscribers) == 0 {
			// TODO: This isn't the best way to handle this.
			// 		Someone can unsub after we exit this for too.
			time.Sleep(100 * time.Millisecond)
		}
		p.Subscribers[p.next%len(p.Subscribers)].Write(in)
		p.next++
	}
	_ = p.Threaded.Base.Close()
}
