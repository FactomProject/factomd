package snapshot

import (
	"bufio"
	"fmt"
	"io"

	"github.com/FactomProject/FactomCode/common"
	"github.com/FactomProject/factomd/Utilities/tools"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/sirupsen/logrus"
)

const (
	EblockPrefix = "eb:"
	EntryPreix   = "et:"
)

type entrySnapshot struct {
	FirstHeight uint32
	NextHeight  uint32
	Directory   string

	eblocksProcessed int
	entriesProcessed int
	chainsCount      int
	Entries          io.WriteCloser
}

func NewEntrySnapshot(dir string) *entrySnapshot {
	es := &entrySnapshot{
		Directory: dir,
	}
	return es
}

func (es *entrySnapshot) Close(log *logrus.Logger) {
	err := es.Entries.Close()
	if err != nil {
		log.WithFields(logrus.Fields{
			"error": err.Error(),
			"file":  fmt.Sprintf("entries"),
		}).Errorf("close file")
	}

}

func (es *entrySnapshot) writeNewEntry(entry interfaces.IEBEntry) error {

	cid := common.NewHash()
	_ = cid.SetBytes(entry.GetChainID().Bytes())
	slimEntry := common.Entry{
		Version: 0,
		ChainID: cid,
		ExtIDs:  entry.ExternalIDs(),
		Content: entry.GetContent(),
	}

	// Wish I could write straight to the buffer
	data, e1 := slimEntry.MarshalBinary()
	L := len(data)
	_, e2 := es.Entries.Write(append([]byte{}, byte(L>>8), byte(L)))
	_, e3 := es.Entries.Write(data)
	switch {
	case e1 != nil:
		return e1
	case e2 != nil:
		return e1
	case e3 != nil:
		return e1
	}
	return nil
}

// writeNewEblock will write the height and eblock hash to the file with the `eb:` prefix
func (es *entrySnapshot) writeNewEblock(eblock interfaces.IEntryBlock) error {

	return nil
}

func (es *entrySnapshot) ChainFile(chainID interfaces.IHash) (io.WriteCloser, error) {

	return nil, nil
}

// Process will process the height specified and load new entries into their flat files.
// We pass the entire database to allow this function to do w/e it needs.
// Passing the height in explicitly just ensures we are loading blocks sequentially
func (es *entrySnapshot) Process(log *logrus.Logger, db tools.Fetcher, height uint32, diagnostic bool) error {
	defer func() {
		es.NextHeight++
	}()

	if height != es.NextHeight {
		return fmt.Errorf("heights must be processed in sequence, exp %d, got %d", es.NextHeight, height)
	}

	dblock, err := db.FetchDBlockByHeight(height)
	if err != nil {
		return fmt.Errorf("fetch fblock %d: %w", height, err)
	}

	eblocks := dblock.GetEBlockDBEntries()
	for _, dbEblock := range eblocks {
		eblock, err := db.FetchEBlock(dbEblock.GetKeyMR())
		if err != nil {
			return fmt.Errorf("fetch eblock (%s) %d: %w", dbEblock.GetKeyMR().String(), height, err)
		}
		err = es.writeNewEblock(eblock)
		if err != nil {
			return fmt.Errorf("write eblock: %w", err)
		}

		entries := eblock.GetEntryHashes()
		cidS := eblock.GetChainID().String()
		var _ = cidS
		for _, entryHash := range entries {
			if entryHash.IsMinuteMarker() {
				continue
			}
			entry, err := db.FetchEntry(entryHash)
			if err != nil {
				return fmt.Errorf("fetch entry (%s) %d: %w", entryHash.String(), height, err)
			}

			err = es.writeNewEntry(entry)
			if err != nil {
				return fmt.Errorf("write entry: %w", err)
			}

			es.chainsCount++
			es.entriesProcessed++
		}
		es.eblocksProcessed++

	}


	return nil
}

type bufWriteCloser struct {
	*bufio.Writer
	io.Closer
}

func newBufWriteCloser(w io.WriteCloser) *bufWriteCloser {
	return &bufWriteCloser{
		Writer: bufio.NewWriter(w),
		Closer: w,
	}
}

func (w *bufWriteCloser) Close() error {
	err := w.Writer.Flush()
	if err != nil {
		return fmt.Errorf("flush: %w", err)
	}
	return w.Closer.Close()
}
