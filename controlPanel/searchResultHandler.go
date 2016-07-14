package controlPanel

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"

	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"

	"github.com/FactomProject/factomd/common/factoid"
)

func handleSearchResult(content *SearchedStruct, w http.ResponseWriter) {
	funcMap := template.FuncMap{
		"truncate": func(s string) string {
			bytes := []byte(s)
			hash := sha256.Sum256(bytes)
			str := fmt.Sprintf(" - Bytes: %d <br /> - Hash: %x", len(bytes), hash)
			return str
		},
	}
	templates.Funcs(funcMap)

	templates.ParseFiles(TEMPLATE_PATH + "searchresults/type/" + content.Type + ".html")
	templates.ParseGlob(TEMPLATE_PATH + "searchresults/*.html")

	var err error
	switch content.Type {
	case "entry":
		entry := getEntry(content.Input)
		if entry == nil {
			break
		}
		content.Content = entry
		err = templates.ExecuteTemplate(w, content.Type, content)
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
		ecblock := getAblock(content.Input)
		if ecblock == nil {
			break
		}
		err = templates.ExecuteTemplate(w, content.Type, ecblock)
	default:
		err = templates.ExecuteTemplate(w, content.Type, content)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func getFblock(hash string) *factoid.FBlock {
	mr, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}
	fblk, err := st.DB.FetchFBlockByPrimary(mr)
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
	fmt.Println(holder.String())

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
	ablk, err := st.DB.FetchABlock(mr)
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
	eblk, err := st.DB.FetchEBlock(mr)
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
	dblk, err := st.DB.FetchDBlock(mr)
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

	Height string
	Hash   string
}

func getEntry(hash string) *EntryHolder {
	entryHash, err := primitives.HexToHash(hash)
	if err != nil {
		return nil
	}
	entry, err := st.DB.FetchEntry(entryHash)
	if err != nil {
		return nil
	}
	if entry == nil {
		return nil
	}
	holder := new(EntryHolder)
	holder.ChainID = entry.GetChainID().String()
	holder.Content = string(entry.GetContent())
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
			holder.ExtIDs = append(holder.ExtIDs[:], "<span id='encoding'><a>Hex  : </a></span><span id='data'>"+hex.EncodeToString(data)+"</span>")
		} else {
			holder.ExtIDs = append(holder.ExtIDs[:], "<span id='encoding'><a>Ascii: </a></span><span id='data'>"+string(data)+"</span>")
		}
	}
	holder.Version = 0
	holder.Height = fmt.Sprintf("%d", entry.GetDatabaseHeight())
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
	mr, err := st.DB.FetchHeadIndexByChainID(chainID)
	if err != nil || mr == nil {
		return nil
	}
	s.Content = mr.String()
	arr = append(arr[:], *s)
	if err != nil {
		return nil
	}

	entries, err := st.DB.FetchAllEntriesByChainID(chainID)
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
