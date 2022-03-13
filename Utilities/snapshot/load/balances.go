package load

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Balances struct {
	Height        int
	FCTAddressMap map[string]int64
	ECAddressMap  map[string]int64
}

func LoadBalances(file io.Reader) (bal *Balances, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic, file format invalid :%v", r)
		}
	}()

	bal = new(Balances)
	bal.FCTAddressMap = make(map[string]int64)
	bal.ECAddressMap = make(map[string]int64)

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, fmt.Errorf("no lines in the file")
	}
	first := scanner.Text()
	parts := strings.Split(first, " ")
	height, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("parse height: %w", err)
	}
	bal.Height = height

	for scanner.Scan() {
		line := scanner.Text()
		if len(line) == 0 {
			continue
		}
		parts := strings.Split(line, ":")
		addr := parts[0]
		balPart := strings.TrimSpace(parts[1])
		balance, err := strconv.ParseInt(balPart, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("parse balance (%s) %s: %w", addr, balPart, err)
		}

		switch addr[:2] {
		case "FA":
			bal.FCTAddressMap[addr] = balance
		case "EC":
			bal.ECAddressMap[addr] = balance
		default:
			return nil, fmt.Errorf("addr %s not reconigzed", addr)
		}
	}
	return bal, nil
}
