package p2p

import (
	"io"
	"sync/atomic"
)

type StatsCollector interface {
	Collect() (mw uint64, mr uint64, bw uint64, br uint64)
}
type ReadWriteCollector interface {
	io.Reader
	io.Writer
	StatsCollector
}

// MetricsReadWriter is a wrapper for net.Conn that allows the package to
// observe the actual amount of bytes passing through it
type MetricsReadWriter struct {
	rw              io.ReadWriter
	messagesWritten uint64
	messagesRead    uint64
	bytesWritten    uint64
	bytesRead       uint64
}

var _ StatsCollector = (*MetricsReadWriter)(nil)
var _ ReadWriteCollector = (*MetricsReadWriter)(nil)
var _ io.ReadWriter = (*MetricsReadWriter)(nil)

func NewMetricsReadWriter(rw io.ReadWriter) *MetricsReadWriter {
	sc := new(MetricsReadWriter)
	sc.rw = rw
	return sc
}

func (sc *MetricsReadWriter) Write(p []byte) (int, error) {
	n, e := sc.rw.Write(p)
	atomic.AddUint64(&sc.messagesWritten, 1)
	atomic.AddUint64(&sc.bytesWritten, uint64(n))
	return n, e
}

func (sc *MetricsReadWriter) Read(p []byte) (int, error) {
	n, e := sc.rw.Read(p)
	atomic.AddUint64(&sc.messagesRead, 1)
	atomic.AddUint64(&sc.bytesRead, uint64(n))
	return n, e
}

func (sc *MetricsReadWriter) Collect() (mw uint64, mr uint64, bw uint64, br uint64) {
	mw = atomic.SwapUint64(&sc.messagesWritten, 0)
	mr = atomic.SwapUint64(&sc.messagesRead, 0)
	bw = atomic.SwapUint64(&sc.bytesWritten, 0)
	br = atomic.SwapUint64(&sc.bytesRead, 0)
	return
}
