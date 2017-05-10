package engine_test

import (
	"bytes"
	"testing"

	"github.com/FactomProject/factomd/common/interfaces"
	. "github.com/FactomProject/factomd/engine"
	"github.com/hashicorp/go-plugin"
)

// FakeEtcdInstance is a fake etcd plugin
type FakeEtcdInstance struct{}

func (f *FakeEtcdInstance) SendIntoEtcd(msg []byte) error          { return nil }
func (f *FakeEtcdInstance) GetData() []byte                        { return nil }
func (f *FakeEtcdInstance) Reinitiate() error                      { return nil }
func (f *FakeEtcdInstance) NewBlockLease(blockHeight uint32) error { return nil }
func (f *FakeEtcdInstance) Ready() (bool, error)                   { return true, nil }

// FakeTorrent is a fake torrent plugin
type FakeTorrent struct{}

func (FakeTorrent) RetrieveDBStateByHeight(height uint32) error     { return nil }
func (FakeTorrent) UploadDBStateBytes(data []byte, sign bool) error { return nil }
func (FakeTorrent) RequestMoreUploads() int                         { return 0 }
func (FakeTorrent) CompletedHeightTo(height uint32) error           { return nil }
func (FakeTorrent) IsBufferEmpty() bool                             { return true }
func (FakeTorrent) FetchFromBuffer() []byte                         { return nil }
func (FakeTorrent) SetSigningKey(sec []byte) error                  { return nil }
func (FakeTorrent) Alive() error                                    { return nil }

// PluginRPCConn returns a plugin RPC client and server that are connected
// together and configured.
func PluginRPCConn(t *testing.T, ps map[string]plugin.Plugin) (*plugin.RPCClient, *plugin.RPCServer) {
	// Create two net.Conns we can use to shuttle our control connection
	clientConn, serverConn := plugin.TestConn(t)

	// Start up the server
	server := &plugin.RPCServer{Plugins: ps, Stdout: new(bytes.Buffer), Stderr: new(bytes.Buffer)}
	go server.ServeConn(serverConn)

	// Connect the client to the server
	client, err := plugin.NewRPCClient(clientConn, ps)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	return client, server
}

// TestTorrentImpl just checks the plugin implementation of the interface
func TestTorrentImpl(t *testing.T) {
	x := new(IManagerPlugin)
	x.Impl = new(FakeTorrent)
	client, _ := PluginRPCConn(t, map[string]plugin.Plugin{
		"manager": x,
	})

	raw, err := client.Dispense("manager")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	mc := raw.(interfaces.IManagerController)

	err = mc.Alive()
	if err != nil {
		t.Error(err)
	}

	err = mc.RetrieveDBStateByHeight(0)
	if err != nil {
		t.Error(err)
	}

	err = mc.UploadDBStateBytes(nil, true)
	if err != nil {
		t.Error(err)
	}

	v := mc.RequestMoreUploads()
	if v != 0 {
		t.Error("Should be 0")
	}

	err = mc.CompletedHeightTo(0)
	if err != nil {
		t.Error(err)
	}

	b := mc.FetchFromBuffer()
	if b != nil {
		t.Error("Should be nil")
	}

	err = mc.SetSigningKey(nil)
	if err != nil {
		t.Error(err)
	}
}

// TestEtcdImpl just checks the plugin implementation of the interface
func TestEtcdImpl(t *testing.T) {
	x := new(IEtcdPlugin)
	x.Impl = new(FakeEtcdInstance)
	client, _ := PluginRPCConn(t, map[string]plugin.Plugin{
		"etcd": x,
	})

	raw, err := client.Dispense("etcd")
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	mc := raw.(interfaces.IEtcdManager)

	err = mc.SendIntoEtcd(nil)
	if err != nil {
		t.Error(err)
	}

	v := mc.GetData()
	if v != nil {
		t.Error("Should be nil")
	}

	err = mc.Reinitiate()
	if err != nil {
		t.Error(err)
	}

	err = mc.NewBlockLease(0)
	if err != nil {
		t.Error(err)
	}

	b, err := mc.Ready()
	if err != nil {
		t.Error(err)
	}
	if !b {
		t.Error("Should be true")
	}
}
