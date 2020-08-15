package engine_test

import (
	"bytes"
	"testing"

	"github.com/PaulSnow/factom2d/common/interfaces"
	. "github.com/PaulSnow/factom2d/engine"
	"github.com/hashicorp/go-plugin"
)

// FakeTorrent is a fake torrent plugin
type FakeTorrent struct{}

func (FakeTorrent) RetrieveDBStateByHeight(height uint32) error     { return nil }
func (FakeTorrent) UploadDBStateBytes(data []byte, sign bool) error { return nil }
func (FakeTorrent) UploadIfOnDisk(height uint32) bool               { return true }
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
	// Client working
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

	v := mc.UploadIfOnDisk(0)
	if v != true {
		t.Error("Should be true")
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

	bo := mc.IsBufferEmpty()
	if !bo {
		t.Error("Should be true")
	}

	// Client not working
	client.Close()
	err = mc.Alive()
	if err == nil {
		t.Error("Stream closed, this should fail")
	}

	err = mc.RetrieveDBStateByHeight(0)
	if err == nil {
		t.Error("Stream closed, this should fail")
	}

	err = mc.UploadDBStateBytes(nil, true)
	if err == nil {
		t.Error("Stream closed, this should fail")
	}

	err = mc.CompletedHeightTo(0)
	if err == nil {
		t.Error("Stream closed, this should fail")
	}

	b = mc.FetchFromBuffer()
	if b == nil || len(b) != 1 {
		t.Error("Should be length 1")
	}

	err = mc.SetSigningKey(nil)
	if err == nil {
		t.Error("Stream closed, this should fail")
	}

	bo = mc.IsBufferEmpty()
	if bo {
		t.Error("Should be false")
	}
}
