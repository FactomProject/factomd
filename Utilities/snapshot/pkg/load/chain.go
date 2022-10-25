package load

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/FactomProject/FactomCode/common"
)

type Eblock struct {
	Height int
	KeyMR  string
	Seq    int64
	// Indexes in the Entries slice [first:last]
	FirstEntry int
	LastEntry  int
}

type Chain struct {
	Height  int
	Eblocks []*Eblock
	Entries []*common.Entry
	Head    *Eblock
}

func (c Chain) EblockEntries(eblock *Eblock) []*common.Entry {
	return c.Entries[eblock.FirstEntry:eblock.LastEntry]
}

func LoadChain(file io.Reader) (chain *Chain, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic, file format invalid :%v", r)
		}
	}()

	c := new(Chain)
	c.Entries = make([]*common.Entry, 0)
	c.Eblocks = make([]*Eblock, 0)

	scanner := bufio.NewScanner(file)
	// entIdx is the index of the NEXT entry in the chain.
	// This idx has yet to actually be allocated.
	var entIdx = 0
	var lastEblock *Eblock

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}

		switch line[:2] {
		case "eb":
			data := strings.TrimPrefix(line, "eb:")
			parts := strings.Split(data, " ")
			height, err := strconv.Atoi(parts[0])
			if err != nil {
				return nil, fmt.Errorf("parse height %s: %w", parts[0], err)
			}
			if lastEblock != nil {
				lastEblock.LastEntry = entIdx
			}

			next := &Eblock{
				Height:     height,
				KeyMR:      parts[1],
				FirstEntry: entIdx,
				Seq:        int64(len(c.Eblocks)),
			}
			c.Eblocks = append(c.Eblocks, next)
			lastEblock = next
		case "et":
			data64 := strings.TrimPrefix(line, "et:")
			data, err := base64.StdEncoding.DecodeString(data64)
			if err != nil {
				return nil, err
			}

			entry := common.NewEntry()
			left, err := entry.UnmarshalBinaryData([]byte(data))
			if err != nil {
				return nil, fmt.Errorf("unmarshal entry: %w", err)
			}
			if len(left) > 0 {
				return nil, fmt.Errorf("%d left over bytes", len(left))
			}

			c.Entries = append(c.Entries, entry)
			entIdx++
		}
	}
	if lastEblock != nil {
		lastEblock.LastEntry = entIdx
	}
	c.Head = lastEblock

	return c, nil
}

// readLine has some complexity as entries can contain newlines
func readLine(scanner *bufio.Scanner) {

}
