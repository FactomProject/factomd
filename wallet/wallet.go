package wallet

import (
	"log"
	"os"
	//"fmt"
	"code.google.com/p/gcfg"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/util"
)

var (
	walletFile      = "wallet.dat"
	walletStorePath = "/tmp/wallet"

	//defaultPrivKey PrivateKey
	keyManager KeyManager
)

func init() {
	util.Trace()
	loadConfigurations()
	loadKeys()
}

func loadKeys() {
	err := keyManager.InitKeyManager(walletStorePath, walletFile)
	if err != nil {
		panic(err)
	}

}

func loadConfigurations() {
	cfg := struct {
		Wallet struct {
			WalletStorePath string
		}
	}{}

	var sf = "wallet.conf"
	wd, err := os.Getwd()
	if err != nil {
		log.Println(err)
	} else {
		sf = wd + "/" + sf
	}

	err = gcfg.ReadFileInto(&cfg, sf)
	if err != nil {
		log.Println(err)
		log.Println("Wallet using default settings...")
	} else {
		log.Println("Walet using settings from: " + sf)
		log.Println(cfg)

		walletStorePath = cfg.Wallet.WalletStorePath
	}

}

func SignData(data []byte) Signature {
	return keyManager.keyPair.Sign(data)
}

//impliment Signer
func Sign(d []byte) Signature { return SignData(d) }

func ClientPublicKey() PublicKey {
	return keyManager.keyPair.Pub
}

func MarshalSign(msg interfaces.BinaryMarshallable) Signature {
	return keyManager.keyPair.MarshalSign(msg)
}

/*
func DetachMarshalSign(msg interfaces.BinaryMarshallable) *DetachedSignature {
	sig := MarshalSign(msg)
	return sig.DetachSig()
}*/

func ClientPublicKeyStr() string {
	return ClientPublicKey().String()
}

/*
func FactoidAddress() string {
	netid := byte('\x07')
	util.Trace("NOT IMPLEMENTED !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!") // FIXME
	return factoid.AddressFromPubKey(ClientPublicKey().Key, netid)
}

func GetMyBalance() (bal int64) {
	//	bal =  factoid.GetBalance(FactoidAddress())
	util.Trace("NOT IMPLEMENTED !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!") // FIXME
	return 0
}
*/
