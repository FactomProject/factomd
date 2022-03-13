package api

import (
	"context"

	"github.com/FactomProject/factom"
)

type FactomFetcher struct {
}

func NewFactomFetcher(factomServer, walletServer string) *FactomFetcher {
	if factomServer != "" {
		factom.SetFactomdServer(factomServer)
	}
	if walletServer != "" {
		factom.SetWalletServer(walletServer)
	}
	return &FactomFetcher{}
}

func (fa *FactomFetcher) FactoidBalance(_ context.Context, addr string) (int64, error) {
	return factom.GetFactoidBalance(addr)
}
func (fa *FactomFetcher) EntryCreditBalance(_ context.Context, addr string) (int64, error) {
	return factom.GetECBalance(addr)
}

func (fa *FactomFetcher) Height(_ context.Context) (int64, error) {
	heights, err := factom.GetHeights()
	if err != nil {
		return -1, err
	}
	return heights.DirectoryBlockHeight, nil
}

func (fa *FactomFetcher) Eblock(_ context.Context, keyMR string) (*factom.EBlock, error) {
	eblock, err := factom.GetEBlock(keyMR)
	if err != nil {
		return nil, err
	}
	return eblock, err
}

func (fa *FactomFetcher) ChainHead(_ context.Context, cid string) (string, error) {
	head, _, err := factom.GetChainHead(cid)
	if err != nil {
		return "", err
	}
	return head, err
}
