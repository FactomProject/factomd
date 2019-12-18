package state

import "fmt"

// ThreadSafeBalanceMap are all the thread safe operations
// More exist for testing.
type ThreadSafeBalanceMap interface {
	GetBalance(address [32]byte) int64
	GetBalances(addresses [][32]byte) map[[32]byte]int64
	UpdateBalances(addresses [][32]byte, deltas []int64) error
	UpdateBalancesFromDeltas(deltas []Delta) error

	Serve()
	Close()
	Closed() bool
}

// BalanceMap enables CRUD of balances in a threadsafe manner.
type BalanceMap struct {
	balances map[[32]byte]int64

	queries chan *BalanceQuery  // Read operations on an address
	updates chan *BalanceUpdate // Read operations on an address

	quit chan struct{}
}

func NewBalanceMap() *BalanceMap {
	b := new(BalanceMap)
	b.balances = make(map[[32]byte]int64)
	b.queries = make(chan *BalanceQuery, 100)
	b.updates = make(chan *BalanceUpdate, 100)
	b.quit = make(chan struct{})

	return b
}

func (bm *BalanceMap) ThreadSafe() ThreadSafeBalanceMap {
	return bm
}

func (bm *BalanceMap) Close() {
	close(bm.quit)
}

// Closed returns true if the BalanceMap is closed and no longer
// servicing returns
func (bm *BalanceMap) Closed() bool {
	select {
	case _, open := <-bm.quit:
		return !open
	default:
		return false
	}
}

// Serve will enable the serving
func (bm *BalanceMap) Serve() {
	for {
		select {
		case <-bm.quit: // closing
		case q := <-bm.queries: // Read only query
			resp := new(BalanceResponse)
			resp.Balances = bm.getBalances(q.Addresses)
			q.reply(resp)
		case u := <-bm.updates:
			resp := bm.ApplyDeltas(u.Deltas)
			u.reply(resp)
		}
	}
}

// GetBalance is threadsafe
func (bm *BalanceMap) GetBalance(address [32]byte) int64 {
	bals := bm.GetBalances([][32]byte{address})
	return bals[address]
}

// GetBalances is threadsafe
func (bm *BalanceMap) GetBalances(addresses [][32]byte) map[[32]byte]int64 {
	query := NewBalanceQuery(addresses)
	bm.queries <- query
	resp := <-query.ReturnChannel // No errors for balance queries
	return resp.Balances
}

func (bm *BalanceMap) UpdateBalances(addresses [][32]byte, deltas []int64) error {
	update, err := NewBalanceUpdateFromLists(addresses, deltas)
	if err != nil {
		return err
	}
	return bm.updateBalances(update)
}

func (bm *BalanceMap) UpdateBalancesFromDeltas(deltas []Delta) error {
	update := NewBalanceUpdate(deltas)
	return bm.updateBalances(update)
}

func (bm *BalanceMap) updateBalances(update *BalanceUpdate) error {
	bm.updates <- update
	resp := <-update.ReturnChannel
	return resp.Error
}

func (bm *BalanceMap) getBalances(addresses [][32]byte) map[[32]byte]int64 {
	bals := make(map[[32]byte]int64)
	for _, a := range addresses {
		bals[a] = bm.balances[a]
	}
	return bals
}

// ApplyDeltas will run through all the balance deltas and apply them.
//	If a balance goes below 0, then the application will fail, and rollback
//	to the balances prior to the application
// On a successful application, we will return all the new balances for
//	addresses in the deltas.
// We will NOT keep 0 bal addresses
func (bm *BalanceMap) ApplyDeltas(deltas []Delta) *BalanceResponse {
	resp := new(BalanceResponse)
	for i, d := range deltas {
		bm.balances[d.Address] += d.Delta
		bal := bm.balances[d.Address]
		if bal < 0 {
			// This delta failed.
			resp.Error = fmt.Errorf("address %x balance went below 0 (%d)", d.Address, bal)
			bm.Rollback(deltas[:i+1])
			return resp
		} else if bal == 0 { // Delete any 0 bal addresses
			delete(bm.balances, d.Address)
		}
	}
	return resp
}

// AllBalances is primarily for unit tests.
func (bm *BalanceMap) AllBalances() map[[32]byte]int64 {
	n := make(map[[32]byte]int64)
	for k, v := range bm.balances {
		n[k] = v
	}
	return n
}

// Rollback will undo all the delta operations in reverse order.
//	The order shouldn't really matter, but it is doing the opposite of the
//	apply.
func (bm *BalanceMap) Rollback(deltas []Delta) {
	for i := len(deltas) - 1; i >= 0; i-- {
		bm.balances[deltas[i].Address] -= deltas[i].Delta
		bal := bm.balances[deltas[i].Address]
		if bal == 0 { // Delete any 0 bal addresses
			delete(bm.balances, deltas[i].Address)
		}
	}
}

// BalanceQuery is a readonly operation to the balance map
type BalanceQuery struct {
	Addresses [][32]byte // The requested addresses to fetch the balance for

	ReturnChannel chan *BalanceResponse
}

func NewBalanceQuery(addresses [][32]byte) *BalanceQuery {
	q := new(BalanceQuery)
	q.Addresses = addresses
	q.ReturnChannel = make(chan *BalanceResponse, 1)
	return q
}

func (q *BalanceQuery) reply(resp *BalanceResponse) {
	q.ReturnChannel <- resp
}

type BalanceUpdate struct {
	Deltas []Delta

	ReturnChannel chan *BalanceResponse
}

func (q *BalanceUpdate) reply(resp *BalanceResponse) {
	q.ReturnChannel <- resp
}

func NewBalanceUpdateFromLists(addresses [][32]byte, deltas []int64) (*BalanceUpdate, error) {
	if len(addresses) != len(deltas) {
		return nil, fmt.Errorf("must have the same length addresses and deltas")
	}

	ds := make([]Delta, len(addresses))
	for i := range deltas {
		ds[i].Address = addresses[i]
		ds[i].Delta = deltas[i]
	}

	return NewBalanceUpdate(ds), nil
}

func NewBalanceUpdate(deltas []Delta) *BalanceUpdate {
	u := new(BalanceUpdate)
	u.Deltas = deltas
	u.ReturnChannel = make(chan *BalanceResponse, 1)
	return u
}

// Delta is a balance update delta
type Delta struct {
	Address [32]byte
	Delta   int64
}

type BalanceResponse struct {
	Balances map[[32]byte]int64
	Error    error
}
