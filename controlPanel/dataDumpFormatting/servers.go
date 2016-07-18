package dataDumpFormatting

import (
	"fmt"

	"github.com/FactomProject/factomd/state"
)

func Identities(st *state.State) string {
	prt := ""
	prt = prt + fmt.Sprintf("=== Identity List ===  Total: %d Displaying: All\n", len(st.Identities))
	for c, i := range st.Identities {
		num := fmt.Sprintf("%d", c)
		prt = prt + "------------------------------------" + num + "----------------------------------------\n"
		stat := returnStatString(i.Status)
		prt = prt + fmt.Sprint("Server Status: ", stat, "\n")
		prt = prt + fmt.Sprint("Identity Chain: ", i.IdentityChainID, "\n")
		prt = prt + fmt.Sprint("Management Chain: ", i.ManagementChainID, "\n")
		prt = prt + fmt.Sprint("Matryoshka Hash: ", i.MatryoshkaHash, "\n")
		prt = prt + fmt.Sprint("Key 1: ", i.Key1, "\n")
		prt = prt + fmt.Sprint("Key 2: ", i.Key2, "\n")
		prt = prt + fmt.Sprint("Key 3: ", i.Key3, "\n")
		prt = prt + fmt.Sprint("Key 4: ", i.Key4, "\n")
		prt = prt + fmt.Sprint("Signing Key: ", i.SigningKey, "\n")
		for _, a := range i.AnchorKeys {
			prt = prt + fmt.Sprintf("Anchor Key: {'%s' L%x T%x K:%x}\n", a.BlockChain, a.KeyLevel, a.KeyType, a.SigningKey)
		}
	}
	return prt
}

func Authorities(st *state.State) string {
	prt := ""
	prt = prt + fmt.Sprintf("=== Authority List ===  Total: %d Displaying: All\n", len(st.Authorities))
	for c, i := range st.Authorities {
		num := fmt.Sprintf("%d", c)
		prt = prt + "--------------------------------------" + num + "---------------------------------------\n"
		var stat string
		switch i.Status {
		case 0:
			stat = "Unassigned"
		case 1:
			stat = "Federated Server"
		case 2:
			stat = "Audit Server"
		case 3:
			stat = "Full"
		case 4:
			stat = "Pending Federated Server"
		case 5:
			stat = "Pending Audit Server"
		case 6:
			stat = "Pending Full"
		case 7:
			stat = "Pending"
		}
		prt = prt + fmt.Sprint("Server Status: ", stat, "\n")
		prt = prt + fmt.Sprint("Identity Chain: ", i.AuthorityChainID, "\n")
		prt = prt + fmt.Sprint("Management Chain: ", i.ManagementChainID, "\n")
		prt = prt + fmt.Sprint("Matryoshka Hash: ", i.MatryoshkaHash, "\n")
		prt = prt + fmt.Sprint("Signing Key: ", i.SigningKey.String(), "\n")
		for _, a := range i.AnchorKeys {
			prt = prt + fmt.Sprintf("Anchor Key: {'%s' L%x T%x K:%x}\n", a.BlockChain, a.KeyLevel, a.KeyType, a.SigningKey)
		}
	}
	return prt
}

func returnStatString(i int) string {
	var stat string
	switch i {
	case 0:
		stat = "Unassigned"
	case 1:
		stat = "Federated Server"
	case 2:
		stat = "Audit Server"
	case 3:
		stat = "Full"
	case 4:
		stat = "Pending Federated Server"
	case 5:
		stat = "Pending Audit Server"
	case 6:
		stat = "Pending Full"
	case 7:
		stat = "Pending"
	}
	return stat
}
