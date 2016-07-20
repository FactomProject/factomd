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
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"

	"github.com/FactomProject/factomd/common/factoid"
)

var _ = htemp.HTMLEscaper("sdf")

func handleSearchResult(content *SearchedStruct, w http.ResponseWriter) {
	if statePointer.GetIdentityChainID() == nil {
		return
	}
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
			return fmt.Sprintf("%f", f)
		},
	}
	templates.Funcs(funcMap)

	templates.ParseFiles(FILES_PATH + "templates/searchresults/type/" + content.Type + ".html")
	templates.ParseGlob(FILES_PATH + "templates/searchresults/*.html")

	var err error
	switch content.Type {
	case "entry":
		entry := getEntry(content.Input)
		if entry == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, entry)
	case "chainhead":
		arr := getAllChainEntries(content.Input)
		if arr == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, arr)
	case "eblock":
		eblk := getEblock(content.Input)
		if eblk == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, eblk)
	case "dblock":
		dblk := getDblock(content.Input)
		if dblk == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, dblk)
	case "ablock":
		ablk := getAblock(content.Input)
		if ablk == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, ablk)
	case "fblock":
		fblk := getFblock(content.Input)
		if fblk == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, fblk)
	case "ecblock":
		ecblock := getECblock(content.Input)
		if ecblock == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, ecblock)
	case "entryack":
		entryAck := getEntryAck(content.Input)
		if entryAck == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, entryAck)
	case "factoidack":
		factoidAck := getFactoidAck(content.Input)
		if factoidAck == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, factoidAck)
	case "facttransaction":
		transaction := getFactTransaction(content.Input)
		if transaction == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, transaction)
	case "ectransaction":
		transaction := getEcTransaction(content.Input)
		if transaction == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, transaction)
	case "EC":
		hash := base58.Decode(content.Input)
		if len(hash) < 34 {
			break
		}
		var fixed [32]byte
		copy(fixed[:], hash[2:34])
		bal := fmt.Sprintf("%d", statePointer.FactoidState.GetECBalance(fixed))
		templates.ExecuteTemplate(w, content.Type,
			struct {
				Balance string
				Address string
			}{bal, content.Input})
	case "FA":
		hash := base58.Decode(content.Input)
		if len(hash) < 34 {
			break
		}
		var fixed [32]byte
		copy(fixed[:], hash[2:34])
		bal := fmt.Sprintf("%.3f", float64(statePointer.FactoidState.GetFactoidBalance(fixed))/1e8)
		templates.ExecuteTemplate(w, content.Type,
			struct {
				Balance string
				Address string
			}{bal, content.Input})
	default:
		err = templates.ExecuteTemplate(w, "not-found", nil)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	templates.ExecuteTemplate(w, "not-found", nil)
}

func getEcTransaction(hash string) interfaces.IECBlockEntry {
	mr, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}
	trans, err := statePointer.DB.FetchECTransaction(mr)
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
	trans, err := statePointer.DB.FetchFactoidTransaction(mr)
	if trans == nil || err != nil {
		return nil
	}
	if trans.GetInputs() == nil {
		return nil
	}
	status := getFactoidAck(hash)
	if status == nil {
		status = new(FactoidAck)
		status.Result.Status = "Unknown"
		return struct {
			interfaces.ITransaction
			FactoidAck
		}{trans, *status}
	}
	return struct {
		interfaces.ITransaction
		FactoidAck
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

func getFactoidAck(hash string) *FactoidAck {
	ackReq := new(wsapi.AckRequest)
	ackReq.TxID = hash
	jReq := primitives.NewJSON2Request("factoid-ack", 0, ackReq)
	resp, err := v2Request(jReq, statePointer.GetPort())
	if err != nil {
		return nil
	}

	data, err := resp.JSONByte()
	if err != nil {
		return nil
	}
	temp := new(FactoidAck)
	err = json.Unmarshal(data, &temp)
	if err != nil {
		return nil
	}
	fmt.Println(resp.String())
	return temp
}

func getEntryAck(hash string) interface{} {
	ackReq := new(wsapi.AckRequest)
	ackReq.TxID = hash

	jReq := primitives.NewJSON2Request("entry-ack", 0, ackReq)
	resp, err := v2Request(jReq, statePointer.GetPort())
	if err != nil {
		return nil
	}

	data, err := resp.JSONByte()
	if err != nil {
		return nil
	}
	var temp struct {
		ID      int    `json:"id"`
		Jsonrpc string `json:"jsonrpc"`
		Result  struct {
			CommitData struct {
				Status string `json:"status"`
			} `json:"commitdata"`
			Committxid string `json:"committxid"`
			EntryData  struct {
				Status string `json:"status"`
			} `json:"entrydata"`
			EntryHash string `json:"entryhash"`
		} `json:"result"`
	}
	err = json.Unmarshal(data, &temp)
	if err != nil {
		return nil
	}
	return temp
}

func getECblock(hash string) interfaces.IEntryCreditBlock {
	mr, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}
	ecblk, err := statePointer.DB.FetchECBlock(mr)
	if ecblk == nil || err != nil {
		return nil
	}
	if ecblk.GetHeader() == nil {
		return nil
	}

	return ecblk
}

func getFblock(hash string) *factoid.FBlock {
	mr, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}
	fblk, err := statePointer.DB.FetchFBlock(mr)
	if fblk == nil || err != nil {
		return nil
	}
	bytes, err := fblk.MarshalBinary()
	if err != nil {
		return nil
	}
	holder := new(factoid.FBlock)
	err = holder.UnmarshalBinary(bytes)
	if err != nil {
		return nil
	}

	return holder
}

type AblockHolder struct {
	Header struct {
		PrevBackRefHash     string `json:"PrevBackRefHash"`
		DBHeight            int    `json:"DBHeight"`
		HeaderExpansionSize int    `json:"HeaderExpansionSize"`
		HeaderExpansionArea string `json:"HeaderExpansionArea"`
		MessageCount        int    `json:"MessageCount"`
		BodySize            int    `json:"BodySize"`
		AdminChainID        string `json:"AdminChainID"`
		ChainID             string `json:"ChainID"`
	} `json:"Header"`
	JsonABEntries     []interface{} `json:"ABEntries"`
	BackReferenceHash string        `json:"BackReferenceHash"`
	LookupHash        string        `json:"LookupHash"`

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
	ablk, err := statePointer.DB.FetchABlock(mr)
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
			disp.Type = "DB Signiture"
			disp.OtherInfo = " "
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
			f := new(adminBlock.AddFederatedServerSigningKey)
			err := f.UnmarshalBinary(data)
			if err != nil {
				continue
			}
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
		}
		holder.ABDisplay = append(holder.ABDisplay, *disp)
	}

	return holder
}

type EblockHolder struct {
	Header struct {
		ChainID      string `json:"ChainID"`
		BodyMR       string `json:"BodyMR"`
		PrevKeyMR    string `json:"PrevKeyMR"`
		PrevFullHash string `json:"PrevFullHash"`
		EBSequence   int    `json:"EBSequence"`
		DBHeight     int    `json:"DBHeight"`
		EntryCount   int    `json:"EntryCount"`
	} `json:"Header"`
	Body struct {
		EBEntries []string `json:"EBEntries"`
	} `json:"Body"`

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
	eblk, err := statePointer.DB.FetchEBlock(mr)
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
	for _, entry := range entries {
		if len(entry.String()) < 32 {
			continue
		} else if entry.String()[:10] == "0000000000" {
			continue
		}
		ent := getEntry(entry.String())
		if ent != nil {
			ent.Hash = entry.String()
			holder.Entries = append(holder.Entries, *ent)
		}
	}

	return holder
}

type DblockHolder struct {
	Header struct {
		Version      int    `json:"Version"`
		NetworkID    int    `json:"NetworkID"`
		BodyMR       string `json:"BodyMR"`
		PrevKeyMR    string `json:"PrevKeyMR"`
		PrevFullHash string `json:"PrevFullHash"`
		Timestamp    int    `json:"Timestamp"`
		DBHeight     int    `json:"DBHeight"`
		BlockCount   int    `json:"BlockCount"`
		ChainID      string `json:"ChainID"`
	} `json:"Header"`
	DBEntries []struct {
		ChainID string `json:"ChainID"`
		KeyMR   string `json:"KeyMR"`
	} `json:"DBEntries"`
	JsonDBHash interface{} `json:"DBHash"`
	JsonKeyMR  interface{} `json:"KeyMR"`

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
	dblk, err := statePointer.DB.FetchDBlock(mr)
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

	return holder
}

type EntryHolder struct {
	ChainID string   `json:"ChainID"`
	Content string   `json:"Content"`
	ExtIDs  []string `json:"ExtIDs"`
	Version int      `json:"Version"`

	Height        string
	Hash          string
	ContentLength int
	ContentHash   string
	ECCost        string
}

func getEntry(hash string) *EntryHolder {
	entryHash, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}
	entry, err := statePointer.DB.FetchEntry(entryHash)
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
	mr, err := statePointer.DB.FetchHeadIndexByChainID(chainID)
	if err != nil || mr == nil {
		return nil
	}
	s.Content = mr.String()
	arr = append(arr[:], *s)
	if err != nil {
		return nil
	}

	entries, err := statePointer.DB.FetchAllEntriesByChainID(chainID)
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
