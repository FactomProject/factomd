package snapshot

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"

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
	NextHeight uint32
	Directory  string

	eblocksProcessed int
	entriesProcessed int
	chains           map[[32]byte]int

	OpenFiles map[[32]byte]io.WriteCloser
}

func NewEntrySnapshot(dir string) *entrySnapshot {
	return &entrySnapshot{
		Directory: dir,
		OpenFiles: make(map[[32]byte]io.WriteCloser),
		chains:    make(map[[32]byte]int),
	}
}

func (es *entrySnapshot) Close(log *logrus.Logger) {
	for cid, f := range es.OpenFiles {
		err := f.Close()
		if err != nil {
			log.WithFields(logrus.Fields{
				"error": err.Error(),
				"chain": fmt.Sprintf("%x", cid),
			}).Errorf("close file")
		}
	}
}

func (es *entrySnapshot) writeNewEntry(entry interfaces.IEBEntry) error {
	file, err := es.ChainFile(entry.GetChainID())
	if err != nil {
		return fmt.Errorf("write eblock: %w", err)
	}

	cid := common.NewHash()
	_ = cid.SetBytes(entry.GetChainID().Bytes())
	slimEntry := common.Entry{
		Version: 0,
		ChainID: cid,
		ExtIDs:  entry.ExternalIDs(),
		Content: entry.GetContent(),
	}

	// Wish I could write straight to the buffer
	data, err := slimEntry.MarshalBinary()
	if err != nil {
		return fmt.Errorf("marshal entry %s: %w", entry.GetHash().String(), err)
	}

	_, err = fmt.Fprint(file, EntryPreix)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(file, base64.StdEncoding.EncodeToString(data))
	if err != nil {
		return err
	}

	return err
}

// writeNewEblock will write the height and eblock hash to the file with the `eb:` prefix
func (es *entrySnapshot) writeNewEblock(eblock interfaces.IEntryBlock) error {
	file, err := es.ChainFile(eblock.GetChainID())
	if err != nil {
		return fmt.Errorf("write eblock: %w", err)
	}

	keyMR, err := eblock.KeyMR()
	if err != nil {
		return fmt.Errorf("keymr: %w", err)
	}
	_, err = file.Write([]byte(fmt.Sprintf("%s%d %s\n", EblockPrefix, eblock.GetDatabaseHeight(), keyMR.String())))
	if err != nil {
		return fmt.Errorf("write eblock: %w", err)
	}
	return err
}

func (es *entrySnapshot) ChainFile(chainID interfaces.IHash) (io.WriteCloser, error) {
	if file, ok := es.OpenFiles[chainID.Fixed()]; ok {
		return file, nil
	}

	file, err := os.OpenFile(filepath.Join(es.Directory, chainID.String()), os.O_CREATE|os.O_WRONLY, 0777)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	w := newBufWriteCloser(file)
	es.OpenFiles[chainID.Fixed()] = w

	return file, nil
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

		cid := eblock.GetChainID().Fixed()
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

			es.chains[cid]++
			es.entriesProcessed++
		}
		es.eblocksProcessed++

	}

	if diagnostic {
		log.WithFields(logrus.Fields{
			"height":  height,
			"entries": es.entriesProcessed,
			"eblocks": es.eblocksProcessed,
			"chains":  len(es.chains),
		}).Info("entry info")
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
