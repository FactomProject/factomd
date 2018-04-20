package dataDumpFormatting

import (
	"encoding/hex"
	"fmt"

	"github.com/FactomProject/factomd/common/constants"
	"github.com/FactomProject/factomd/state"
)

func Identities(copyDS state.DisplayState) string {
	prt := ""
	prt = prt + fmt.Sprintf("=== Identity List ===   Total: %d Displaying: All\n", len(copyDS.Identities))
	for c, i := range copyDS.Identities {
		num := fmt.Sprintf("%d", c)
		prt = prt + "------------------------------------" + num + "---------------------------------------\n"
		stat := constants.IdentityStatusString(i.Status)
		prt = prt + fmt.Sprint("Server Status: ", stat, "\n")
		prt = prt + fmt.Sprintf("Synced Status: ID[%t] MG[%t]\n", i.IdentityChainSync.Synced(), i.ManagementChainSync.Synced())
		prt = prt + fmt.Sprintf("Identity Chain: %s (C:%d R:%d)\n", i.IdentityChainID.String(), i.IdentityCreated, i.IdentityRegistered)
		prt = prt + fmt.Sprintf("Management Chain: %s (C:%d R:%d)\n", i.ManagementChainID.String(), i.ManagementCreated, i.ManagementRegistered)
		prt = prt + fmt.Sprint("Matryoshka Hash: ", i.MatryoshkaHash, "\n")
		prt = prt + fmt.Sprint("Key 1: ", i.Keys[0], "\n")
		prt = prt + fmt.Sprint("Key 2: ", i.Keys[1], "\n")
		prt = prt + fmt.Sprint("Key 3: ", i.Keys[2], "\n")
		prt = prt + fmt.Sprint("Key 4: ", i.Keys[3], "\n")
		prt = prt + fmt.Sprint("Signing Key: ", i.SigningKey, "\n")
		prt = prt + fmt.Sprint("Coinbase Address: ", i.GetCoinbaseHumanReadable(), "\n")
		for _, a := range i.AnchorKeys {
			prt = prt + fmt.Sprintf("Anchor Key: {'%s' L%x T%x K:%x}\n", a.BlockChain, a.KeyLevel, a.KeyType, a.SigningKey)
		}
		prt += fmt.Sprintf("ID Eblock Syncing: Current: %d  Target: %d\n", i.IdentityChainSync.Current.DBHeight, i.IdentityChainSync.Target.DBHeight)
		prt += fmt.Sprintf("MG Eblock Syncing: Current: %d  Target: %d\n", i.ManagementChainSync.Current.DBHeight, i.ManagementChainSync.Target.DBHeight)
	}
	return prt
}

func Authorities(copyDS state.DisplayState) string {
	prt := ""
	prt = prt + fmt.Sprintf("=== Authority List ===   Total: %d Displaying: All\n", len(copyDS.Authorities))
	for c, i := range copyDS.Authorities {
		num := fmt.Sprintf("%d", c)
		prt = prt + "------------------------------------" + num + "---------------------------------------\n"
		stat := constants.IdentityStatusString(i.Status)
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
