package api

import "context"

type APIFetcher interface {
	FactoidBalance(ctx context.Context, addr string) (int64, error)
	EntryCreditBalance(ctx context.Context, addr string) (int64, error)
	Height(ctx context.Context) (int64, error)
}
