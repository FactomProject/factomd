package p2p

import (
	"sync"
	"time"
)

// Measure measures the per-second rates of messages and bandwidth based
// on individual calls
type Measure struct {
	parcelsIn  uint64
	parcelsOut uint64
	bytesIn    uint64
	bytesOut   uint64
	dataMtx    sync.Mutex

	rateParcelOut float64
	rateParcelIn  float64
	rateBytesOut  float64
	rateBytesIn   float64
	rateMtx       sync.RWMutex

	rate time.Duration
}

// NewMeasure initializes a new measuring tool based on the given rate
// Rate must be at least one second and should be multiples of seconds
func NewMeasure(rate time.Duration) *Measure {
	m := new(Measure)
	m.rate = rate
	go m.calculate()
	return m
}

// GetRate returns the current rates measured this interval
// (Parcels Received, Parcels Sent, Bytes Received, Bytes Sent)
func (m *Measure) GetRate() (float64, float64, float64, float64) {
	m.rateMtx.RLock()
	defer m.rateMtx.RUnlock()
	return m.rateParcelIn, m.rateParcelOut, m.rateBytesIn, m.rateBytesOut
}

func (m *Measure) calculate() {
	ticker := time.NewTicker(m.rate)
	sec := m.rate.Seconds()
	for range ticker.C {
		m.dataMtx.Lock()
		m.rateMtx.Lock()

		m.rateParcelOut = float64(m.parcelsOut) / sec
		m.rateBytesOut = float64(m.bytesOut) / sec
		m.rateParcelIn = float64(m.parcelsIn) / sec
		m.rateBytesIn = float64(m.bytesIn) / sec

		m.parcelsIn = 0
		m.parcelsOut = 0
		m.bytesIn = 0
		m.bytesOut = 0

		m.dataMtx.Unlock()
		m.rateMtx.Unlock()
	}
}

// Send signals that we sent a parcel of the given size
func (m *Measure) Send(size uint64) {
	m.dataMtx.Lock()
	m.parcelsOut++
	m.bytesOut += size
	m.dataMtx.Unlock()
}

// Receive signals that we received a parcel of the given size
func (m *Measure) Receive(size uint64) {
	m.dataMtx.Lock()
	m.parcelsIn++
	m.bytesIn += size
	m.dataMtx.Unlock()
}
