package controlPanel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	htemp "html/template"
	"net/http"
	"strconv"
	"text/template"

	"github.com/FactomProject/btcutil/base58"
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/controlPanel/files"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"

	"strings"

	"github.com/FactomProject/factomd/common/factoid"
)

var _ = htemp.HTMLEscaper("sdf")

func HandleSearchResult(content *SearchedStruct, w http.ResponseWriter) {
	// Functions able to be used within the html
	funcMap := template.FuncMap{
		"truncate": func(s string) string {
			bytes := []byte(s)
			hash := sha256.Sum256(bytes)
			str := fmt.Sprintf(" - Bytes: %d <br /> - Hash: %x", len(bytes), hash)
			return str
		},
		"AddressFACorrect": func(s string) string {
			hash, err := primitives.HexToHash(s)
			if err != nil {
				return "There has been an error converting the address"
			}
			prefix := []byte{0x5f, 0xb1}
			addr := hash.Bytes()
			addr = append(prefix, addr[:]...)
			oneSha := sha256.Sum256(addr)
			twoSha := sha256.Sum256(oneSha[:])
			addr = append(addr, twoSha[:4]...)
			str := base58.Encode(addr)
			return str
		},
		"AddressECCorrect": func(s string) string {
			hash, err := primitives.HexToHash(s)
			if err != nil {
				return "There has been an error converting the address"
			}
			prefix := []byte{0x59, 0x2a}
			addr := hash.Bytes()
			addr = append(prefix, addr[:]...)
			oneSha := sha256.Sum256(addr)
			twoSha := sha256.Sum256(oneSha[:])
			addr = append(addr, twoSha[:4]...)
			str := base58.Encode(addr)
			return str
		},
		"TransactionAmountCorrect": func(u uint64) string {
			s := fmt.Sprintf("%d", u)
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return s
			}
			f = f / 1e8
			return fmt.Sprintf("%.8f", f)
		},
	}
	TemplateMutex.Lock()
	templates.Funcs(funcMap)
	files.CustomParseGlob(templates, "templates/searchresults/*.html")
	files.CustomParseFile(templates, "templates/searchresults/type/"+content.Type+".html")
	TemplateMutex.Unlock()

	var err error
	_ = err
	switch content.Type {
	case "entry":
		entry := getEntry(content.Input)
		if entry == nil {
			break
		}
		TemplateMutex.Lock()
		err = templates.ExecuteTemplate(w, content.Type, entry)
		TemplateMutex.Unlock()
		return
	case "chainhead":
		arr := getAllChainEntries(content.Input)
		if arr == nil {
			break
		}
		arr[0].Content = struct {
			Head   interface{}
			Length int
		}{arr[0].Content, len(arr) - 1}
		TemplateMutex.Lock()
		err = templates.ExecuteTemplate(w, content.Type, arr)
		TemplateMutex.Unlock()
		return
	case "eblock":
		eblk := getEblock(content.Input)
		if eblk == nil {
			break
		}
		TemplateMutex.Lock()
		err = templates.ExecuteTemplate(w, content.Type, eblk)
		TemplateMutex.Unlock()
		return
	case "dblock":
		dblk := getDblock(content.Input)
		if dblk == nil {
			break
		}
		TemplateMutex.Lock()
		err = templates.ExecuteTemplate(w, content.Type, dblk)
		TemplateMutex.Unlock()
		return
	case "ablock":
		ablk := getAblock(content.Input)
		if ablk == nil {
			break
		}
		TemplateMutex.Lock()
		err = templates.ExecuteTemplate(w, content.Type, ablk)
		TemplateMutex.Unlock()
		return
	case "fblock":
		fblk := getFblock(content.Input)
		if fblk == nil {
			break
		}
		TemplateMutex.Lock()
		err = templates.ExecuteTemplate(w, content.Type, fblk)
		TemplateMutex.Unlock()
		return
	case "ecblock":
		ecblock := getECblock(content.Input)
		if ecblock == nil {
			break
		}
		TemplateMutex.Lock()
		err = templates.ExecuteTemplate(w, content.Type, ecblock)
		TemplateMutex.Unlock()
		return
	case "entryack":
		entryAck := getEntryAck(content.Input)
		if entryAck == nil {
			break
		}
		TemplateMutex.Lock()
		err = templates.ExecuteTemplate(w, content.Type, entryAck)
		TemplateMutex.Unlock()
		return
	case "factoidack":
		factoidAck := getFactoidAck(content.Input)
		if factoidAck == nil {
			break
		}
		TemplateMutex.Lock()
		err = templates.ExecuteTemplate(w, content.Type, factoidAck)
		TemplateMutex.Unlock()
		return
	case "facttransaction":
		transaction := getFactTransaction(content.Input)
		if transaction == nil {
			break
		}
		TemplateMutex.Lock()
		err = templates.ExecuteTemplate(w, content.Type, transaction)
		TemplateMutex.Unlock()
		return
	case "ectransaction":
		transaction := getEcTransaction(content.Input)
		if transaction == nil {
			break
		}
		TemplateMutex.Lock()
		err = templates.ExecuteTemplate(w, content.Type, transaction)
		TemplateMutex.Unlock()
		return
	case "EC":
		hash := base58.Decode(content.Input)
		if len(hash) < 34 {
			break
		}
		var fixed [32]byte
		copy(fixed[:], hash[2:34])
		bal := fmt.Sprintf("%d", StatePointer.FactoidState.GetECBalance(fixed))
		TemplateMutex.Lock()
		templates.ExecuteTemplate(w, content.Type,
			struct {
				Balance string
				Address string
			}{bal, content.Input})
		TemplateMutex.Unlock()
		return
	case "FA":
		hash := base58.Decode(content.Input)
		if len(hash) < 34 {
			break
		}
		var fixed [32]byte
		copy(fixed[:], hash[2:34])
		bal := fmt.Sprintf("%.8f", float64(StatePointer.FactoidState.GetFactoidBalance(fixed))/1e8)
		TemplateMutex.Lock()
		templates.ExecuteTemplate(w, content.Type,
			struct {
				Balance string
				Address string
			}{bal, content.Input})
		TemplateMutex.Unlock()
		return
	}

	TemplateMutex.Lock()
	files.CustomParseFile(templates, "templates/searchresults/type/notfound.html")
	templates.ExecuteTemplate(w, "notfound", content.Input)
	TemplateMutex.Unlock()
}

func getEcTransaction(hash string) interfaces.IECBlockEntry {
	mr, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}

	dbase := StatePointer.GetDB()
	trans, err := dbase.FetchECTransaction(mr)

	if trans == nil || err != nil {
		return nil
	}
	if trans.GetEntryHash() == nil {
		return nil
	}
	return trans
}

func getFactTransaction(hash string) interfaces.ITransaction {
	mr, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}

	dbase := StatePointer.GetDB()
	trans, err := dbase.FetchFactoidTransaction(mr)

	if trans == nil || err != nil {
		return nil
	}
	if trans.GetInputs() == nil {
		return nil
	}
	status := getFactoidAck(hash)
	if status == nil {
		return struct {
			interfaces.ITransaction
			wsapi.FactoidTxStatus
		}{trans, *status}
	}
	return struct {
		interfaces.ITransaction
		wsapi.FactoidTxStatus
	}{trans, *status}
}

type FactoidAck struct {
	ID      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  struct {
		Status                string  `json:"status"`
		TransactionDate       float64 `json:"transactiondate"`
		TransactionDateString string  `json:"transactiondatestring"`
		Txid                  string  `json:"txid"`
	} `json:"result"`
}

func getFactoidAck(hash string) *wsapi.FactoidTxStatus {
	ackReq := new(wsapi.AckRequest)
	ackReq.TxID = hash
	answers, err := wsapi.HandleV2FactoidACK(StatePointer, ackReq)
	if answers == nil || err != nil {
		return nil
	}
	return answers.(*wsapi.FactoidTxStatus)
}

func getEntryAck(hash string) *wsapi.EntryStatus {
	ackReq := new(wsapi.AckRequest)
	ackReq.TxID = hash
	answers, err := wsapi.HandleV2EntryACK(StatePointer, ackReq)
	if answers == nil || err != nil {
		return nil
	}
	return (answers.(*wsapi.EntryStatus))
}

type ECBlockHolder struct {
	ECBlock interfaces.IEntryCreditBlock
	Length  int
}

func getECblock(hash string) *ECBlockHolder {
	mr, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}

	dbase := StatePointer.GetDB()
	ecblk, err := dbase.FetchECBlock(mr)

	if ecblk == nil || err != nil {
		return nil
	}
	if ecblk.GetHeader() == nil {
		return nil
	}

	holder := new(ECBlockHolder)
	holder.ECBlock = ecblk
	length := 0
	zero := primitives.NewZeroHash()
	for _, e := range ecblk.GetEntryHashes() {
		if e != nil && !e.IsSameAs(zero) {
			length++
		}
	}
	holder.Length = length

	return holder
}

type FBlockHolder struct {
	factoid.FBlock
	Length int
}

func getFblock(hash string) *FBlockHolder {
	mr, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}

	dbase := StatePointer.GetDB()
	fblk, err := dbase.FetchFBlock(mr)

	if fblk == nil || err != nil {
		return nil
	}
	bytes, err := fblk.MarshalBinary()
	if err != nil {
		return nil
	}
	holder := new(FBlockHolder)
	err = holder.UnmarshalBinary(bytes)
	if err != nil {
		return nil
	}

	holder.Length = len(holder.Transactions)
	return holder
}

type AblockHolder struct {
	Header struct {
		PrevBackRefHash     string `json:"prevbackrefhash"`
		DBHeight            int    `json:"dbheight"`
		HeaderExpansionSize int    `json:"headerexpansionsize"`
		HeaderExpansionArea string `json:"headerexpansionarea"`
		MessageCount        int    `json:"messagecount"`
		BodySize            int    `json:"bodysize"`
		AdminChainID        string `json:"adminchainid"`
		ChainID             string `json:"chainid"`
	} `json:"header"`
	JsonABEntries     []interface{} `json:"abentries"`
	BackReferenceHash string        `json:"backreferencehash"`
	LookupHash        string        `json:"lookuphash"`

	ABEntries []interfaces.IABEntry
	ABDisplay []ABDisplayHolder
}

type ABDisplayHolder struct {
	Type      string
	OtherInfo string
}

func getAblock(hash string) *AblockHolder {
	mr, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}

	holder := new(AblockHolder)

	dbase := StatePointer.GetDB()
	ablk, err := dbase.FetchABlock(mr)

	if ablk == nil || err != nil {

		return nil
	}
	bytes, err := ablk.JSONByte()
	if err != nil {
		return nil
	}
	err = json.Unmarshal(bytes, holder)
	if err != nil {
		return nil
	}

	holder.ABEntries = ablk.GetABEntries()

	for _, entry := range holder.ABEntries {
		disp := new(ABDisplayHolder)
		data, err := entry.MarshalBinary()
		if err != nil {
			return nil
		}
		switch entry.Type() {
		case constants.TYPE_MINUTE_NUM:
			r := new(adminBlock.EndOfMinuteEntry)
			err := r.UnmarshalBinary(data)
			if err != nil {
				continue
			}
			disp.Type = "Minute Number"
			disp.OtherInfo = fmt.Sprintf("%x", r.MinuteNumber)
		case constants.TYPE_DB_SIGNATURE:
			r := new(adminBlock.DBSignatureEntry)
			err := r.UnmarshalBinary(data)
			if err != nil {
				continue
			}
			disp.Type = "DB Signature"
			disp.OtherInfo = "Server: " + r.IdentityAdminChainID.String()
		case constants.TYPE_REVEAL_MATRYOSHKA:
			r := new(adminBlock.RevealMatryoshkaHash)
			err := r.UnmarshalBinary(data)
			if err != nil {
				continue
			}
			disp.Type = "Reveal Matryoshka Hash"
			disp.OtherInfo = "Identity ChainID: <a href='' id='factom-search-link' type='chainhead'>" + r.IdentityChainID.String() + "</a><br />MHash: " + r.MHash.String()
		case constants.TYPE_ADD_MATRYOSHKA:
			m := new(adminBlock.AddReplaceMatryoshkaHash)
			err := m.UnmarshalBinary(data)
			if err != nil {
				continue
			}
			disp.Type = "Add Matryoshka Hash"
			disp.OtherInfo = "Identity ChainID: <a href='' id='factom-search-link' type='chainhead'>" + m.IdentityChainID.String() + "</a><br />MHash: " + m.MHash.String()
		case constants.TYPE_ADD_SERVER_COUNT:
			s := new(adminBlock.IncreaseServerCount)
			err := s.UnmarshalBinary(data)
			if err != nil {
				continue
			}
			disp.Type = "Add Server Count"
			disp.OtherInfo = fmt.Sprintf("%x", s.Amount)
		case constants.TYPE_ADD_FED_SERVER:
			f := new(adminBlock.AddFederatedServer)
			err := f.UnmarshalBinary(data)
			if err != nil {
				continue
			}
			disp.Type = "Add Federated Server"
			disp.OtherInfo = "Identity ChainID: <a href='' id='factom-search-link' type='chainhead'>" + f.IdentityChainID.String() + "</a>"
		case constants.TYPE_ADD_AUDIT_SERVER:
			a := new(adminBlock.AddAuditServer)
			err := a.UnmarshalBinary(data)
			if err != nil {
				continue
			}
			disp.Type = "Add Audit Server"
			disp.OtherInfo = "Identity ChainID: <a href='' id='factom-search-link' type='chainhead'>" + a.IdentityChainID.String() + "</a>"
		case constants.TYPE_REMOVE_FED_SERVER:
			f := new(adminBlock.RemoveFederatedServer)
			err := f.UnmarshalBinary(data)
			if err != nil {
				continue
			}
			disp.Type = "Remove Server"
			disp.OtherInfo = "Identity ChainID: <a href='' id='factom-search-link' type='chainhead'>" + f.IdentityChainID.String() + "</a>"
		case constants.TYPE_ADD_FED_SERVER_KEY:
			f := entry.(*adminBlock.AddFederatedServerSigningKey)
			disp.Type = "Add Server Key"
			disp.OtherInfo = "Identity ChainID: <a href='' id='factom-search-link' type='chainhead'>" + f.IdentityChainID.String() + "</a><br />Key: " + f.PublicKey.String()
		case constants.TYPE_ADD_BTC_ANCHOR_KEY:
			b := new(adminBlock.AddFederatedServerBitcoinAnchorKey)
			err := b.UnmarshalBinary(data)
			if err != nil {
				continue
			}
			disp.Type = "Add Bitcoin Server Key"
			disp.OtherInfo = "Identity ChainID: <a href='' id='factom-search-link' type='chainhead'>" + b.IdentityChainID.String() + "</a>"

			//Coinbase related
		case constants.TYPE_COINBASE_DESCRIPTOR:
			f := entry.(*adminBlock.CoinbaseDescriptor)
			disp.Type = "Coinbase Descriptor"
			sep := ""
			for _, o := range f.Outputs {
				disp.OtherInfo += fmt.Sprintf("%sAddress: <a href='' id='factom-search-link' type='FA'>%s</a> Amount: %s", sep, primitives.ConvertFctAddressToUserStr(o.GetAddress()), FactoshiToFactoid(o.GetAmount()))
				sep = "<br />"
			}

		case constants.TYPE_COINBASE_DESCRIPTOR_CANCEL:
			//f := new(adminBlock.AddFederatedServerSigningKey)
			//err := f.UnmarshalBinary(data)
			//if err != nil {
			//	continue
			//}
			disp.Type = "Coinbase Descriptor Cancel"
			disp.OtherInfo = "Not implemented"
		case constants.TYPE_ADD_FACTOID_ADDRESS:
			f := entry.(*adminBlock.AddFactoidAddress)
			disp.Type = "Add Coinbase Address"
			disp.OtherInfo = "Identity ChainID: <a href='' id='factom-search-link' type='chainhead'>" + f.IdentityChainID.String() + "</a><br />"
			disp.OtherInfo += fmt.Sprintf("Address: <a href='' id='factom-search-link' type='FA'>%s</a>", primitives.ConvertFctAddressToUserStr(f.FactoidAddress))
		case constants.TYPE_ADD_FACTOID_EFFICIENCY:
			f := entry.(*adminBlock.AddEfficiency)
			disp.Type = "Add Authority Efficiency"
			disp.OtherInfo = "Identity ChainID: <a href='' id='factom-search-link' type='chainhead'>" + f.IdentityChainID.String() + "</a><br />"
			disp.OtherInfo += fmt.Sprintf("Efficiency: %s%%", primitives.EfficiencyToString(f.Efficiency))
		default:
			// Forward compatible
			_, ok := entry.(*adminBlock.ForwardCompatibleEntry)
			if !ok {
				disp.Type = "Unknown"
				data, _ := entry.MarshalBinary()
				disp.OtherInfo = fmt.Sprintf("Type: %x<br />Raw: %x", entry.Type(), data)
			} else {
				disp.Type = "Forward Compatible Type"
				data, _ := entry.MarshalBinary()
				disp.OtherInfo = fmt.Sprintf("This entry is not defined, you are probably running old software and should update.<br />Type: 0x%02x<br />Raw: %x\n", entry.Type(), data)
			}
		}
		holder.ABDisplay = append(holder.ABDisplay, *disp)
	}

	return holder
}

// FactoshiToFactoid converts a uint64 factoshi ammount into a fixed point
// number represented as a string
func FactoshiToFactoid(i uint64) string {
	d := i / 1e8
	r := i % 1e8
	ds := fmt.Sprintf("%d", d)
	rs := fmt.Sprintf("%08d", r)
	rs = strings.TrimRight(rs, "0")
	if len(rs) > 0 {
		ds = ds + "."
	}
	return fmt.Sprintf("%s%s", ds, rs)
}

type EblockHolder struct {
	Header struct {
		ChainID      string `json:"chainid"`
		BodyMR       string `json:"bodymr"`
		PrevKeyMR    string `json:"prevkeymr"`
		PrevFullHash string `json:"prevfullhash"`
		EBSequence   int    `json:"ebsequence"`
		DBHeight     int    `json:"dbheight"`
		EntryCount   int    `json:"entrycount"`
	} `json:"header"`
	Body struct {
		EBEntries []string `json:"ebentries"`
	} `json:"body"`

	KeyMR    string
	BodyMR   string
	FullHash string
	Entries  []EntryHolder
}

func getEblock(hash string) *EblockHolder {
	mr, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}
	holder := new(EblockHolder)

	dbase := StatePointer.GetDB()
	eblk, err := dbase.FetchEBlock(mr)

	if eblk == nil || err != nil {
		return nil
	}
	bytes, err := eblk.JSONByte()
	if err != nil {
		return nil
	}
	err = json.Unmarshal(bytes, holder)
	if err != nil {
		return nil
	}

	if keymr, err := eblk.KeyMR(); err != nil {
		holder.KeyMR = "Error"
	} else {
		holder.KeyMR = keymr.String()
	}
	holder.BodyMR = eblk.BodyKeyMR().String()
	holder.FullHash = eblk.GetHash().String()

	entries := eblk.GetEntryHashes()
	count := 0
	for _, entry := range entries {
		if len(entry.String()) < 32 {
			continue
		} else if entry.String()[:10] == "0000000000" {
			ent := new(EntryHolder)
			ent.Hash = "Minute Marker"
			num := entry.String()[63:]
			if num == "a" {
				num = "10"
			}
			ent.ChainID = num

			holder.Entries = append(holder.Entries, *ent)
			continue
		}
		ent := getEntry(entry.String())
		count++
		if ent != nil {
			ent.Hash = entry.String()
			holder.Entries = append(holder.Entries, *ent)
		}
	}
	holder.Header.EntryCount = count

	return holder
}

type DblockHolder struct {
	Header struct {
		Version      int    `json:"version"`
		NetworkID    int    `json:"networkid"`
		BodyMR       string `json:"bodymr"`
		PrevKeyMR    string `json:"prevkeymr"`
		PrevFullHash string `json:"prevfullhash"`
		Timestamp    uint32 `json:"timestamp"`
		DBHeight     int    `json:"dbheight"`
		BlockCount   int    `json:"blockcount"`
		ChainID      string `json:"chainid"`

		FormatedTimeStamp string
	} `json:"header"`
	DBEntries []struct {
		ChainID string `json:"chainid"`
		KeyMR   string `json:"keymr"`
	} `json:"dbentries"`
	JsonDBHash interface{} `json:"dbhash"`
	JsonKeyMR  interface{} `json:"keymr"`

	EBlocks    []EblockHolder
	AdminBlock struct {
		ChainID string
		KeyMr   string
	}
	FactoidBlock struct {
		ChainID string
		KeyMr   string
	}
	EntryCreditBlock struct {
		ChainID string
		KeyMr   string
	}
	FullHash string
	KeyMR    string
}

func getDblock(hash string) *DblockHolder {
	mr, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}
	holder := new(DblockHolder)

	dbase := StatePointer.GetDB()
	dblk, err := dbase.FetchDBlock(mr)

	if dblk == nil || err != nil {
		return nil
	}
	bytes, err := dblk.JSONByte()
	if err != nil {
		return nil
	}
	err = json.Unmarshal(bytes, holder)
	if err != nil {
		return nil
	}

	blocks := dblk.GetDBEntries()
	for _, block := range blocks {
		if len(block.GetKeyMR().String()) < 32 {
			continue
		} else if block.GetChainID().String()[:10] == "0000000000" {
			// Admin/FC/EC block
			switch block.GetChainID().String() {
			case "000000000000000000000000000000000000000000000000000000000000000a":
				holder.AdminBlock.ChainID = block.GetChainID().String()
				holder.AdminBlock.KeyMr = block.GetKeyMR().String()
			case "000000000000000000000000000000000000000000000000000000000000000c":
				holder.EntryCreditBlock.ChainID = block.GetChainID().String()
				holder.EntryCreditBlock.KeyMr = block.GetKeyMR().String()
			case "000000000000000000000000000000000000000000000000000000000000000f":
				holder.FactoidBlock.ChainID = block.GetChainID().String()
				holder.FactoidBlock.KeyMr = block.GetKeyMR().String()
			}
			continue
		}
		blk := getEblock(block.GetKeyMR().String())
		if blk != nil {
			holder.EBlocks = append(holder.EBlocks, *blk)
		}
	}

	holder.FullHash = dblk.GetHash().String()
	holder.KeyMR = dblk.GetKeyMR().String()

	ts := dblk.GetTimestamp()
	holder.Header.FormatedTimeStamp = ts.String()
	return holder
}

type EntryHolder struct {
	ChainID string   `json:"chainid"`
	Content string   `json:"content"`
	ExtIDs  []string `json:"extids"`
	Version int      `json:"version"`

	Height        string
	Hash          string
	ContentLength int
	ContentHash   string
	ECCost        string

	Time string
}

func getEntry(hash string) *EntryHolder {
	entryHash, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}
	dbase := StatePointer.GetDB()
	entry, err := dbase.FetchEntry(entryHash)

	if err != nil {
		return nil
	}
	if entry == nil {
		return nil
	}

	holder := new(EntryHolder)
	holder.Hash = hash
	holder.ChainID = entry.GetChainID().String()
	max := byte(0x80)
	for _, data := range entry.ExternalIDs() {
		hexString := false
		for _, bytes := range data {
			if bytes > max {
				hexString = true
				break
			}
		}
		if hexString {
			str := hex.EncodeToString(data)
			holder.ExtIDs = append(holder.ExtIDs[:], "<span id='encoding'><a>Hex  : </a></span><span id='data'>"+htemp.HTMLEscaper(str)+"</span>")
		} else {
			str := string(data)
			holder.ExtIDs = append(holder.ExtIDs[:], "<span id='encoding'><a>Ascii: </a></span><span id='data'>"+htemp.HTMLEscaper(str)+"</span>")
		}
	}
	holder.Version = 0
	holder.Height = fmt.Sprintf("%d", entry.GetDatabaseHeight())
	holder.ContentLength = len(entry.GetContent())
	data := sha256.Sum256(entry.GetContent())
	content := string(entry.GetContent())
	holder.Content = htemp.HTMLEscaper(content)
	if bytes, err := entry.MarshalBinary(); err != nil {
		holder.ECCost = "Error"
	} else {
		if eccost, err := util.EntryCost(bytes); err != nil {
			holder.ECCost = "Error"
		} else {
			holder.ECCost = fmt.Sprintf("%d", eccost)
		}
	}

	//holder.Content = string(entry.GetContent())
	holder.ContentHash = primitives.NewHash(data[:]).String()
	return holder
}

func getAllChainEntries(chainIDString string) []SearchedStruct {
	arr := make([]SearchedStruct, 0)
	chainID, err := primitives.HexToHash(chainIDString)
	if err != nil {
		return nil
	}
	s := new(SearchedStruct)
	s.Type = "chainhead"
	s.Input = chainID.String()

	dbase := StatePointer.GetDB()
	mr, err := dbase.FetchHeadIndexByChainID(chainID)

	if err != nil || mr == nil {
		return nil
	}
	s.Content = mr.String()
	arr = append(arr[:], *s)
	if err != nil {
		return nil
	}

	entries := make([]interfaces.IEBEntry, 0)

	dbase = StatePointer.GetDB()
	eblks, err := dbase.FetchAllEBlocksByChain(chainID)
	if err != nil {

		return nil
	}

	for _, eblk := range eblks {
		hashes := eblk.GetEntryHashes()
		for _, hash := range hashes {
			entry, err := dbase.FetchEntry(hash)
			if err != nil || entry == nil {
				continue
			}
			entries = append(entries, entry)
		}
	}
	//entries, err := dbase.FetchAllEntriesByChainID(chainID)

	if err != nil {
		return nil
	}

	for _, entry := range entries {
		s := new(SearchedStruct)
		s.Type = "entry"
		e := getEntry(entry.GetHash().String())
		s.Content = e
		s.Input = entry.GetHash().String()
		arr = append(arr[:], *s)
	}
	return arr
}
