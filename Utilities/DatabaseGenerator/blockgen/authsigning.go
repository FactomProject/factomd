package blockgen

import (
	"github.com/FactomProject/factomd/common/adminBlock"
	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
)

// IAuthSigner is an interface that is able to return dbsigs given a previous block.
// 	It keeps track of authority sets and keys
type IAuthSigner interface {
	SignBlock(prev *state.DBState) interfaces.IAdminBlock
}

// DefaultAuthSigner only signs for 38bab
type DefaultAuthSigner struct {
}

func (DefaultAuthSigner) SignBlock(prev *state.DBState) interfaces.IAdminBlock {
	ab := adminBlock.NewAdminBlock(prev.AdminBlock)
	hash, err := prev.AdminBlock.BackReferenceHash()
	if err != nil {
		panic(err.Error())
	}
	ab.GetHeader().SetPrevBackRefHash(hash)
	h, _ := primitives.HexToHash("38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9")
	pkey, _ := primitives.NewPrivateKeyFromHex("cc1985cdfae4e32b5a454dfda8ce5e1361558482684f3367649c3ad852c8e31a")
	data, _ := prev.DirectoryBlock.GetHeader().MarshalBinary()
	sig := pkey.Sign(data)
	ab.AddDBSig(h, sig)

	ab.MarshalBinary()
	return ab
}
