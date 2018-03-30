package engine

// All plugins we can intiate

import (
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
	"github.com/FactomProject/factomd/common/primitives"
	"github.com/FactomProject/factomd/state"
	"github.com/hashicorp/go-plugin"
)

// How often to check the plugin if it has messages ready
var CHECK_BUFFER time.Duration = 2 * time.Second

var _ log.Logger
var _ = ioutil.Discard

// pluginMap is the map of plugins we can dispense.
var pluginMap = map[string]plugin.Plugin{
	"manager": &IManagerPlugin{},
}

var managerHandshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "Torrent_Manager",
	MagicCookieValue: "factom_torrent",
}

// LaunchDBStateManagePlugin launches the plugin and returns an interface that
// can be interacted with like a usual interface. The client returned must be
// killed before we exit
func LaunchDBStateManagePlugin(path string, inQueue interfaces.IQueue, s *state.State, sigKey *primitives.PrivateKey, memProfileRate int) (interfaces.IManagerController, error) {
	// So we don't get debug logs. Comment this out if you want to keep plugin
	// logs
	log.SetOutput(ioutil.Discard)

	// We're a host! Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: managerHandshakeConfig,
		Plugins:         pluginMap,
		Cmd:             exec.Command(path+"factomd-torrent", fmt.Sprintf("-mpr=%d", memProfileRate)),
	})

	stop := make(chan int, 10)

	// Make sure we close our client on close
	AddInterruptHandler(func() {
		fmt.Println("Manager plugin is now closing...")
		client.Kill()
		stop <- 0
	})

	// Connect via RPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	// Request the plugin
	raw, err := rpcClient.Dispense("manager")
	if err != nil {
		return nil, err
	}

	// This feels like a normal interface
	// implementation but is in fact over an RPC connection.
	manager := raw.(interfaces.IManagerController)

	if sigKey != nil {
		manager.SetSigningKey(sigKey.Key[:32])
	}

	// Upload controller controls how quickly we upload things. Keeping our rate similar to
	// the torrent prevents data being backed up in memory
	s.Uploader = state.NewUploadController(manager)
	AddInterruptHandler(func() {
		fmt.Println("State's Upload controller is now closing...")
		s.Uploader.Close()
	})

	// We need to drain messages returned by the plugin.
	go manageDrain(inQueue, manager, s, stop)
	// RunUploadController controls our uploading to ensure not overloading queues and consuming too
	// much memory
	go s.RunUploadController()
	// StartTorrentSyncing will use torrents to sync past our current height if we are not synced up
	go s.StartTorrentSyncing()

	return manager, nil
}

// manageDrain handles messages being returned by the plugin, since our requests are asynchronous
// When we make a request via a retrieve, this function will pick up the return
func manageDrain(inQueue interfaces.IQueue, man interfaces.IManagerController, s *state.State, quit chan int) {
	cm := NewCompletedHeightManager()
	for {
		select {
		case <-quit:
			return
		default:
			if err := man.Alive(); err != nil {
				log.Fatal("ERROR: Connection lost to torrent plugin")
				return
			}

			// If there is something for us to grab
			if !man.IsBufferEmpty() {
				var data []byte
				// Exit conditions: If empty, quit. If length == 1 and first/only byte it 0x00
				for !(man.IsBufferEmpty() || (len(data) == 1 && data[0] == 0x00)) {
					// If we have too much to process, do not spam inqueue, let the plugin hold it
					if inQueue.Length() > 1000 {
						for inQueue.Length() > 500 {
							time.Sleep(100 * time.Millisecond)
						}
					}
					data = man.FetchFromBuffer()
					dbMsg := new(messages.DBStateMsg)
					err := dbMsg.UnmarshalBinary(data)
					if err != nil {
						log.Println("Error unmarshaling dbstate from plugin: ", err)
						continue
					}

					// Set the highest completed
					if s.HighestCompletedTorrent < dbMsg.DirectoryBlock.GetDatabaseHeight() {
						s.HighestCompletedTorrent = dbMsg.DirectoryBlock.GetDatabaseHeight()
					}

					// Already processed this height completely
					if dbMsg.DirectoryBlock.GetDatabaseHeight() < s.EntryDBHeightComplete {
						continue
					}

					if !cm.CompleteHeight(int(dbMsg.DirectoryBlock.GetDatabaseHeight())) {
						continue
					}
					cm.ClearTo(int(s.EntryDBHeightComplete))

					inQueue.Enqueue(dbMsg)
				}
			}

			time.Sleep(CHECK_BUFFER)
		}
	}
}

// CompletedHeightsManager ensures the same height is not processed many times
type CompletedHeightsManager struct {
	Completed []int64
	Base      int
}

func NewCompletedHeightManager() *CompletedHeightsManager {
	return new(CompletedHeightsManager)
}

// CompleteHeight will signal a height has been completed. It will return a boolean value
// to indicate whether or not to allow this height to be added to the inmsg queue
func (c *CompletedHeightsManager) CompleteHeight(height int) bool {
	if height < c.Base {
		return false
	}

	endHeight := len(c.Completed) + c.Base
	if endHeight <= height {
		needed := (height - endHeight) + 1
		if needed < 500 {
			needed = 500
		}
		c.Completed = append(c.Completed, make([]int64, needed)...)
	}

	now := time.Now().Unix()
	last := c.Completed[height-c.Base]
	if last == 0 {
		c.Completed[height-c.Base] = now
		return true
	} else if now-240 < last {
		// If completed less than 4min ago
		return false
	}
	c.Completed[height-c.Base] = now

	return true
}

// ClearTo will indicate anything below this height is no longer needed
func (c *CompletedHeightsManager) ClearTo(height int) {
	// If it's close, no point in doing anything. Just wait
	if height <= c.Base+200 {
		return
	}
	clear := height - c.Base
	c.Base = height
	if len(c.Completed) < clear {
		c.Completed = make([]int64, 0)
		return
	}
	c.Completed = c.Completed[clear:]
}
