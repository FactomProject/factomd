package controlPanel

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"text/template"

	"github.com/FactomProject/factomd/common/primitives"
)

func handleSearchResult(content *SearchedStruct, w http.ResponseWriter) {
	funcMap := template.FuncMap{
		"truncate": func(s string) string {
			bytes := []byte(s)
			hash := sha256.Sum256(bytes)
			str := fmt.Sprintf(" - Bytes: %d <br /> - Hash: %x", len(bytes), hash)
			return str
			/*str := s
			ret := ""
			if len(s) > 100 {
				for len(str) > 100 {
					ret = ret + str[:101] + "\n"
					str = str[100:]
				}
			}
			ret = ret + str[:]
			return ret*/
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
		/*	case "address":
			if content.Input[:2] == "EC" {
				st.DB.
			} else if content.Input[:2] == "FA" {

			}*/
	default:
		err = templates.ExecuteTemplate(w, content.Type, content)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

type EblockHolder struct {
}

func getEblock(hash string) *EblockHolder {
	return nil
}

type EntryHolder struct {
	ChainID string   `json:"ChainID"`
	Content string   `json:"Content"`
	ExtIDs  []string `json:"ExtIDs"`
	Version int      `json:"Version"`

	Height string
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
