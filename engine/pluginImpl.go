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

func (s *IManagerPluginRPCServer) RetrieveDBStateByHeight(height uint32, resp *error) error {
	*resp = s.Impl.RetrieveDBStateByHeight(height)
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
*		interfaces.IEtcdManager		**
*										**
******************************************/

// Here is an implementation that talks over RPC
type IEtcdPluginRPC struct{ client *rpc.Client }

func (g *IEtcdPluginRPC) RetrieveDBStateByHeight(height uint32) error {
	var resp error
	err := g.client.Call("Plugin.RetrieveDBStateByHeight", height, &resp)
	if err != nil {
		return err
	}

	return resp
}

type SendIntoEtcdArgs struct {
	Msg []byte // interfaces.IMsg
}

type SendIntoEtcdData struct {
	Error error
}

func (g *IEtcdPluginRPC) SendIntoEtcd(msg []byte) error {
	var resp SendIntoEtcdData
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
	var resp SendIntoEtcdData

	err := g.client.Call("Plugin.Reinitiate", new(interface{}), &resp)
	if err != nil {
		g.client.Close()
		return err
	}

	return nil
}

type GetFromEtcdData struct {
	Bytes    []byte
	NewIndex int64
}

func (g *IEtcdPluginRPC) GetData(oldIndex int64) ([]byte, int64) {
	var resp GetFromEtcdData
	err := g.client.Call("Plugin.GetData", oldIndex, &resp)
	if err != nil {
		return nil, oldIndex
	}

	return resp.Bytes, resp.NewIndex
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

func (s *IEtcdPluginRPCServer) SendIntoEtcd(args *SendIntoEtcdArgs, resp *SendIntoEtcdData) error {
	err := s.Impl.SendIntoEtcd(args.Msg)
	resp.Error = err
	return nil
}

func (s *IEtcdPluginRPCServer) Reinitiate(args interface{}, resp *SendIntoEtcdData) error {
	return s.Impl.Reinitiate()
}

func (s *IEtcdPluginRPCServer) GetData(arg int64, resp *GetFromEtcdData) error {
	dataBytes, newIndex := s.Impl.GetData(arg)
	resp.Bytes = dataBytes
	resp.NewIndex = newIndex
	return nil
}

func (s *IEtcdPluginRPCServer) Ready(args interface{}, resp *ReadyArgs) error {
	ready, err := s.Impl.Ready()
	resp.Err = err
	resp.Ready = ready
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
