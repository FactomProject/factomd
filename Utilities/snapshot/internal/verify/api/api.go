package api

import (
	"context"

	"github.com/FactomProject/factom"
)

type APIFetcher interface {
	FactoidBalance(ctx context.Context, addr string) (int64, error)
	EntryCreditBalance(ctx context.Context, addr string) (int64, error)
	Height(ctx context.Context) (int64, error)

	Eblock(ctx context.Context, keyMR string) (*factom.EBlock, error)
	ChainHead(ctx context.Context, cid string) (string, error)
}
