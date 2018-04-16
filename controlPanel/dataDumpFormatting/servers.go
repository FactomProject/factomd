package dataDumpFormatting

import (
	"encoding/hex"
	"fmt"

	"github.com/FactomProject/factomd/state"
)

func Identities(copyDS state.DisplayState) string {
	prt := ""
	prt = prt + fmt.Sprintf("=== Identity List ===   Total: %d Displaying: All\n", len(copyDS.Identities))
	for c, i := range copyDS.Identities {
		num := fmt.Sprintf("%d", c)
		prt = prt + "------------------------------------" + num + "---------------------------------------\n"
		stat := returnStatString(i.Status)
		prt = prt + fmt.Sprint("Server Status: ", stat, "\n")
		prt = prt + fmt.Sprint("Identity Chain: ", i.IdentityChainID, "\n")
		prt = prt + fmt.Sprint("Management Chain: ", i.ManagementChainID, "\n")
		prt = prt + fmt.Sprint("Matryoshka Hash: ", i.MatryoshkaHash, "\n")
		prt = prt + fmt.Sprint("Key 1: ", i.Keys[0], "\n")
		prt = prt + fmt.Sprint("Key 2: ", i.Keys[1], "\n")
		prt = prt + fmt.Sprint("Key 3: ", i.Keys[2], "\n")
		prt = prt + fmt.Sprint("Key 4: ", i.Keys[3], "\n")
		prt = prt + fmt.Sprint("Signing Key: ", i.SigningKey, "\n")
		for _, a := range i.AnchorKeys {
			prt = prt + fmt.Sprintf("Anchor Key: {'%s' L%x T%x K:%x}\n", a.BlockChain, a.KeyLevel, a.KeyType, a.SigningKey)
		}
	}
	return prt
}

func Authorities(copyDS state.DisplayState) string {
	prt := ""
	prt = prt + fmt.Sprintf("=== Authority List ===   Total: %d Displaying: All\n", len(copyDS.Authorities))
	for c, i := range copyDS.Authorities {
		num := fmt.Sprintf("%d", c)
		prt = prt + "------------------------------------" + num + "---------------------------------------\n"
		stat := returnStatString(i.Status)
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

func MyNodeInfo(copyDS state.DisplayState) string {
	prt := ""
	prt = prt + fmt.Sprintf("My Node: %s\n", copyDS.NodeName)
	if copyDS.IdentityChainID == nil {
		prt = prt + fmt.Sprint("Identity ChainID: \n")
	} else {
		prt = prt + fmt.Sprint("Identity ChainID: ", copyDS.IdentityChainID, "\n")

	}
	pub := copyDS.PublicKey
	if data, err := pub.MarshalBinary(); err != nil {
		prt = prt + fmt.Sprintf("Signing Key: \n")
	} else {
		prt = prt + fmt.Sprintf("Signing Key: %s\n", hex.EncodeToString(data))

	}
	return prt
}

func returnStatString(i uint8) string {
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
		stat = "Skeleton Identity"
	}
	return stat
}
