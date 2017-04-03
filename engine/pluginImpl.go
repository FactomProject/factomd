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
	BlockHeight uint32
	MinuteNum   int
	Msg         []byte // interfaces.IMsg
}

func (g *IEtcdPluginRPC) SendIntoEtcd(blockHeight uint32, minuteNum int, msg []byte) error {
	var resp error
	args := SendIntoEtcdArgs{
		BlockHeight: blockHeight,
		MinuteNum:   minuteNum,
		Msg:         msg,
	}
	err := g.client.Call("Plugin.SendIntoEtcd", &args, &resp)
	if err != nil {
		return err
	}

	return resp
}

type GetMinuteData struct {
	BlockHeight uint32
	MinuteNum   int
}

func (g *IEtcdPluginRPC) GetMinuteData(blockHeight uint32, minuteNum int) [][]byte {
	var resp [][]byte
	args := SendIntoEtcdArgs{
		BlockHeight: blockHeight,
		MinuteNum:   minuteNum,
	}
	err := g.client.Call("Plugin.GetMinuteData", &args, &resp)
	if err != nil {
		return nil
	}

	return resp
}

func (g *IEtcdPluginRPC) GetBlockData(blockHeight uint32) [][]byte {
	var resp [][]byte
	err := g.client.Call("Plugin.GetBlockData", blockHeight, &resp)
	if err != nil {
		return nil
	}

	return resp
}

// Here is the RPC server that IEtcdPluginRPC talks to, conforming to
// the requirements of net/rpc
type IEtcdPluginRPCServer struct {
	// This is the real implementation
	Impl interfaces.IEtcdManager
}

func (s *IEtcdPluginRPCServer) SendIntoEtcd(args *SendIntoEtcdArgs, resp *error) error {
	*resp = s.Impl.SendIntoEtcd(args.BlockHeight, args.MinuteNum, args.Msg)
	return *resp
}

func (s *IEtcdPluginRPCServer) GetMinuteData(args *GetMinuteData, resp *[][]byte) error {
	*resp = s.Impl.GetMinuteData(args.BlockHeight, args.MinuteNum)
	return nil
}

func (s *IEtcdPluginRPCServer) GetBlockData(blockheight uint32, resp *[][]byte) error {
	*resp = s.Impl.GetBlockData(blockheight)
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
