package p2p

import (
	"io"
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
	rw io.ReadWriter

	messagesWritten uint64
	messagesRead    uint64
	bytesWritten    uint64
	bytesRead       uint64
}

var _ StatsCollector = (*MetricsReadWriter)(nil)
var _ ReadWriteCollector = (*MetricsReadWriter)(nil)

func NewMetricsReadWriter(rw io.ReadWriter) *MetricsReadWriter {
	sc := new(MetricsReadWriter)
	sc.rw = rw
	return sc
}

func (sc *MetricsReadWriter) Write(p []byte) (int, error) {
	n, e := sc.rw.Write(p)
	sc.messagesWritten++
	sc.bytesWritten += uint64(n)
	return n, e
}

func (sc *MetricsReadWriter) Read(p []byte) (int, error) {
	n, e := sc.rw.Read(p)
	sc.messagesRead++
	sc.bytesRead += uint64(n)
	return n, e
}

func (sc *MetricsReadWriter) Collect() (mw uint64, mr uint64, bw uint64, br uint64) {
	mw = sc.messagesWritten
	sc.messagesWritten = 0
	mr = sc.messagesRead
	sc.messagesRead = 0
	bw = sc.bytesWritten
	sc.bytesWritten = 0
	br = sc.bytesRead
	sc.bytesRead = 0
	return
}
