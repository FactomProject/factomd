package p2p

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Dialer is a construct to throttle dialing and limit by attempts
type Dialer struct {
	dialer      net.Dialer
	bindTo      string
	interval    time.Duration
	timeout     time.Duration
	attempts    map[Endpoint]time.Time
	attemptsMtx sync.RWMutex
}

// NewDialer creates a new Dialer
func NewDialer(bindTo string, interval, timeout time.Duration) (*Dialer, error) {
	d := new(Dialer)
	d.interval = interval
	d.timeout = timeout
	d.attempts = make(map[Endpoint]time.Time)

	err := d.Bind(bindTo)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (d *Dialer) Bind(to string) error {
	local, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:0", to))
	if err != nil {
		return err
	}
	d.bindTo = to
	d.dialer = net.Dialer{
		LocalAddr: local,
		Timeout:   d.timeout,
	}
	return nil
}

// CanDial checks if the given ip can be dialed yet
func (d *Dialer) CanDial(ep Endpoint) bool {
	d.attemptsMtx.RLock()
	defer d.attemptsMtx.RUnlock()
	if a, ok := d.attempts[ep]; !ok || time.Since(a) >= d.interval {
		return true
	}

	return false
}

// Dial an ip. Returns the active TCP connection or error if it failed to connect
func (d *Dialer) Dial(ep Endpoint) (net.Conn, error) {
	d.attemptsMtx.Lock() // don't unlock with defer so we can dial concurrently
	if t, ok := d.attempts[ep]; ok && time.Since(t) < d.interval {
		d.attemptsMtx.Unlock()
		return nil, fmt.Errorf("dialing too soon")
	}
	d.attempts[ep] = time.Now()
	d.attemptsMtx.Unlock()

	con, err := d.dialer.Dial("tcp", ep.String())
	if err != nil {
		return nil, err
	}
	return con, nil
}
