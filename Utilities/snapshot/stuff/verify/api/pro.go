package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/FactomProject/factom"
)

type FactomProFetcher struct {
	// Cache is a cheeky way to speed things up
	Cache map[string]*proBalance
}

func NewFactomPro(ctx context.Context) (*FactomProFetcher, error) {
	f := &FactomProFetcher{}
	err := f.preloadCache(ctx)
	if err != nil {
		return nil, fmt.Errorf("preload cache: %w", err)
	}
	return f, nil
}

type proBalance struct {
	PubAddr string `json:"pubAddress"`
	PubKey  string `json:"pubKey"`
	Balance int64  `json:"balance"`
	Type    string `json:"type"`
}

func (fa *FactomProFetcher) FactoidBalance(ctx context.Context, addr string) (int64, error) {
	return fa.balance(ctx, addr)
}

func (fa *FactomProFetcher) EntryCreditBalance(ctx context.Context, addr string) (int64, error) {
	return fa.balance(ctx, addr)
}

func (fa *FactomProFetcher) preloadCache(ctx context.Context) error {
	if fa.Cache == nil {
		fa.Cache = make(map[string]*proBalance)
		var ret struct {
			Result []*proBalance `json:"result"`
		}

		_, err := jsonRequest(ctx,
			"https://explorer.factom.pro/explorer/richlist",
			"GET",
			nil,
			&ret)
		if err != nil {
			return err
		}

		for _, addr := range ret.Result {
			fa.Cache[addr.PubAddr] = addr
		}
	}
	return nil
}

func (fa *FactomProFetcher) Height(ctx context.Context) (int64, error) {
	var ret struct {
		Result struct {
			Version          string `json:"version"`
			BlockchainHeight int64  `json:"blockchainHeight"`
			APIHeight        int64  `json:"apiHeight"`
			Name             string `json:"name"`
			AnchorsHeight    struct {
				Btc int `json:"btc"`
				Eth int `json:"eth"`
			} `json:"anchorsHeight"`
		} `json:"result"`
	}

	_, err := jsonRequest(ctx, "https://explorer.factom.pro/explorer", "GET", nil, &ret)
	if err != nil {
		return -1, err
	}
	return ret.Result.BlockchainHeight, nil
}

func (fa *FactomProFetcher) balance(ctx context.Context, addr string) (int64, error) {
	if v, ok := fa.Cache[addr]; ok {
		return v.Balance, nil
	}

	var ret struct {
		Result proBalance `json:"result"`
	}
	_, err := jsonRequest(ctx,
		fmt.Sprintf("https://explorer.factom.pro/explorer/addresses/%s", addr),
		"GET",
		nil,
		&ret)
	if err != nil {
		return -1, err
	}
	return ret.Result.Balance, nil
}

func jsonRequest(ctx context.Context, u string, method string, body interface{}, ret interface{}) (*http.Response, error) {
	bodyBuf := bytes.NewBuffer([]byte{})
	if body != nil {
		err := json.NewEncoder(bodyBuf).Encode(body)
		if err != nil {
			return nil, fmt.Errorf("encode body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyBuf)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	cli := http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := cli.Do(req)
	if err != nil {
		return resp, err
	}
	defer resp.Body.Close()

	if ret != nil {
		err := json.NewDecoder(resp.Body).Decode(&ret)
		if err != nil {
			return resp, fmt.Errorf("decode resp: %w", err)
		}
	}

	return resp, nil
}

func (FactomProFetcher) Eblock(_ context.Context, keyMR string) (*factom.EBlock, error) {
	return nil, fmt.Errorf("not implemented for facotm.pro")
}

func (FactomProFetcher) ChainHead(_ context.Context, cid string) (string, error) {
	return "", fmt.Errorf("not implemented for facotm.pro")
}
