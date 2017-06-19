// Interface that allows factomd to offload the dbstate fetching to this
// plugin. If offloaded, factomd will need to drain the buffer by launching
// a drain go routine
package engine

import (
	"net/rpc"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/hashicorp/go-plugin"
)

/*****************************************
 *										**
 *				Torrents				**
 *		interfaces.IManagerPlugin		**
 *										**
 *****************************************/

// Here is an implementation that talks over RPC
type IManagerPluginRPC struct{ client *rpc.Client }

func (g *IManagerPluginRPC) RetrieveDBStateByHeight(height uint32) error {
	var resp error
	err := g.client.Call("Plugin.RetrieveDBStateByHeight", height, &resp)
	if err != nil {
		return err
	}

	return resp
}

func (g *IManagerPluginRPC) CompletedHeightTo(height uint32) error {
	var resp error
	err := g.client.Call("Plugin.CompletedHeightTo", height, &resp)
	if err != nil {
		return err
	}

	return resp
}

func (g *IManagerPluginRPC) UploadIfOnDisk(height uint32) bool {
	var resp bool
	err := g.client.Call("Plugin.UploadIfOnDisk", height, &resp)
	if err != nil {
		return false
	}
	return resp
}

func (g *IManagerPluginRPC) Alive() error {
	var resp error
	err := g.client.Call("Plugin.Alive", new(interface{}), &resp)
	return err
}

type UploadDBStateArgs struct {
	Data []byte
	Sign bool
}

func (g *IManagerPluginRPC) UploadDBStateBytes(data []byte, sign bool) error {
	var resp error
	args := UploadDBStateArgs{
		Data: data,
		Sign: sign,
	}
	err := g.client.Call("Plugin.UploadDBStateBytes", &args, &resp)
	if err != nil {
		return err
	}

	return resp
}

func (g *IManagerPluginRPC) SetSigningKey(sec []byte) error {
	var resp error
	err := g.client.Call("Plugin.SetSigningKey", sec, &resp)
	if err != nil {
		return err
	}

	return resp
}

func (g *IManagerPluginRPC) IsBufferEmpty() bool {
	var resp bool
	err := g.client.Call("Plugin.IsBufferEmpty", new(interface{}), &resp)
	if err != nil {
		return false
	}

	return resp
}

func (g *IManagerPluginRPC) FetchFromBuffer() []byte {
	var resp []byte
	err := g.client.Call("Plugin.FetchFromBuffer", new(interface{}), &resp)
	if err != nil {
		return []byte{0x00}
	}

	return resp
}

// Here is the RPC server that IManagerPluginRPC talks to, conforming to
// the requirements of net/rpc
type IManagerPluginRPCServer struct {
	// This is the real implementation
	Impl interfaces.IManagerController
}

func (s *IManagerPluginRPCServer) UploadIfOnDisk(height uint32, resp *bool) error {
	*resp = s.Impl.UploadIfOnDisk(height)
	return nil
}
func (s *IManagerPluginRPCServer) Alive(args interface{}, resp *error) error {
	*resp = s.Impl.Alive()
	return *resp
}

func (s *IManagerPluginRPCServer) RetrieveDBStateByHeight(height uint32, resp *error) error {
	*resp = s.Impl.RetrieveDBStateByHeight(height)
	return *resp
}

func (s *IManagerPluginRPCServer) CompletedHeightTo(height uint32, resp *error) error {
	*resp = s.Impl.CompletedHeightTo(height)
	return *resp
}

func (s *IManagerPluginRPCServer) UploadDBStateBytes(args *UploadDBStateArgs, resp *error) error {
	*resp = s.Impl.UploadDBStateBytes(args.Data, args.Sign)
	return *resp
}

func (s *IManagerPluginRPCServer) IsBufferEmpty(args interface{}, resp *bool) error {
	*resp = s.Impl.IsBufferEmpty()
	return nil
}

func (s *IManagerPluginRPCServer) SetSigningKey(key []byte, resp *error) error {
	*resp = s.Impl.SetSigningKey(key)
	return *resp
}

func (s *IManagerPluginRPCServer) FetchFromBuffer(args interface{}, resp *[]byte) error {
	*resp = s.Impl.FetchFromBuffer()
	return nil
}

// This is the implementation of plugin.Plugin so we can serve/consume this
//
// This has two methods: Server must return an RPC server for this plugin
// type. We construct a IManagerPluginRPCServer for this.
//
// Client must return an implementation of our interface that communicates
// over an RPC client. We return IManagerPluginRPC for this.
//
// Ignore MuxBroker. That is used to create more multiplexed streams on our
// plugin connection and is a more advanced use case.
type IManagerPlugin struct {
	// Impl Injection
	Impl interfaces.IManagerController
}

func (p *IManagerPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &IManagerPluginRPCServer{Impl: p.Impl}, nil
}

func (IManagerPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &IManagerPluginRPC{client: c}, nil
}

/*****************************************
*										**
*				Etcd					**
*		interfaces.IEtcdManager			**
*										**
******************************************/

// Here is an implementation that talks over RPC
type IEtcdPluginRPC struct{ client *rpc.Client }

type SendIntoEtcdArgs struct {
	Msg []byte // interfaces.IMsg
}

type PickUpFromHashArgs struct {
	MessageHash string
}

type NewBlockLeaseArgs struct {
	Height uint32
}

type ErrorData struct {
	Error error
}

func (g *IEtcdPluginRPC) SendIntoEtcd(msg []byte) error {
	var resp ErrorData
	args := SendIntoEtcdArgs{
		Msg: msg,
	}
	err := g.client.Call("Plugin.SendIntoEtcd", &args, &resp)
	if err != nil {
		return err
	}

	//log.Println(resp.NewIndex)
	return resp.Error
}

func (g *IEtcdPluginRPC) Reinitiate() error {
	var resp ErrorData

	err := g.client.Call("Plugin.Reinitiate", new(interface{}), &resp)
	if err != nil {
		g.client.Close()
		return err
	}

	return nil
}

func (g *IEtcdPluginRPC) NewBlockLease(blockHeight uint32) error {
	var resp ErrorData
	args := NewBlockLeaseArgs{
		Height: blockHeight,
	}
	err := g.client.Call("Plugin.NewBlockLease", &args, &resp)
	if err != nil {
		g.client.Close()
		return err
	}

	return nil
}

func (g *IEtcdPluginRPC) PickUpFromHash(messageHash string) error {
	var resp ErrorData
	args := PickUpFromHashArgs{
		MessageHash: messageHash,
	}
	err := g.client.Call("Plugin.PickUpFromHash", &args, &resp)
	if err != nil {
		return err
	}
	return resp.Error
}

type GetFromEtcdData struct {
	Bytes []byte
}

func (g *IEtcdPluginRPC) GetData() []byte {
	var resp GetFromEtcdData
	err := g.client.Call("Plugin.GetData", new(interface{}), &resp)
	if err != nil {
		return nil
	}

	return resp.Bytes
}

type ReadyArgs struct {
	Ready bool
	Err   error
}

func (g *IEtcdPluginRPC) Ready() (bool, error) {
	var resp ReadyArgs
	err := g.client.Call("Plugin.Ready", new(interface{}), &resp)
	if err != nil {
		return false, err
	}
	return resp.Ready, resp.Err
}

// Here is the RPC server that IEtcdPluginRPC talks to, conforming to
// the requirements of net/rpc
type IEtcdPluginRPCServer struct {
	// This is the real implementation
	Impl interfaces.IEtcdManager
}

func (s *IEtcdPluginRPCServer) SendIntoEtcd(args *SendIntoEtcdArgs, resp *ErrorData) error {
	err := s.Impl.SendIntoEtcd(args.Msg)
	if err != nil {
		return err
	}
	resp.Error = err
	return nil
}

func (s *IEtcdPluginRPCServer) Reinitiate(args interface{}, resp *ErrorData) error {
	return s.Impl.Reinitiate()
}

func (s *IEtcdPluginRPCServer) NewBlockLease(args *NewBlockLeaseArgs, resp *ErrorData) error {
	err := s.Impl.NewBlockLease(args.Height)
	resp.Error = err
	return nil
}

func (s *IEtcdPluginRPCServer) GetData(args interface{}, resp *GetFromEtcdData) error {
	dataBytes := s.Impl.GetData()
	resp.Bytes = dataBytes
	return nil
}

func (s *IEtcdPluginRPCServer) Ready(args interface{}, resp *ReadyArgs) error {
	ready, err := s.Impl.Ready()
	resp.Err = err
	resp.Ready = ready
	return nil
}

func (s *IEtcdPluginRPCServer) PickUpFromHash(args *PickUpFromHashArgs, resp *ErrorData) error {
	err := s.Impl.PickUpFromHash(args.MessageHash)
	resp.Error = err
	return nil
}

type IEtcdPlugin struct {
	// Impl Injection
	Impl interfaces.IEtcdManager
}

func (p *IEtcdPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &IEtcdPluginRPCServer{Impl: p.Impl}, nil
}

func (IEtcdPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &IEtcdPluginRPC{client: c}, nil
}
