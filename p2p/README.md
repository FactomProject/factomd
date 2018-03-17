
# Factom P2P Networking

### Diagram
![diagram.jpg](https://raw.githubusercontent.com/FactomProject/factomd/m2/p2p/diagram.jpg)

## Conceptual

The P2P Network for factom is a custom library which is mostly independent of hte factomd codebase.  It has almost no external dependencies.  

It is designed to be autonamous and operate without configuration.  When starting up a node will look to the information in the configuration file and the command line to determine who to connect to.  The seedURL is a source of initial peers to connect ot for new nodes.  Each node will attempt to keep at least 8 outgoing connections live and allow a larger number of incomming connections. 

Nodes share peers with each other when they first connect, and periodically thereafter.  Nodes also check the messages they get from other nodes ot verify they are on the same network (eg: production blockchain vs testnet) and are of compatible software versions among other things.  Each connection results in merits or demerits depending on the quality of the connection.  The nodes keep a quality score on a per-IP basis.

Please note that all the networking is IPV4.  IPV6 is not supported.  Additionally, this network will not tunnel thru NAT.

Nodes can be set up to only dial out to a limited set of peers, called "special peers".  Special peers are not shareed with other peers in the network. Additionally, special peers will always be connected to and if there are conectivity problems the connections will remain persistent, and constantly reconnect. Special peers can be determined on the command line or in the configuration file. 

## Operations

#### Command line options
```
  -customnet string
    	This string specifies a custom blockchain network ID.
  -network string
    	Network to join: MAIN, TEST or LOCAL
  -networkPort int
    	Address for p2p network to listen on.
  -peers string
    	Array of peer addresses. 
      These peers are considered "special"
```
#### Config file

An example of the config file is below.  Network determines which network we are participating in.  Main is the production blockchain.  TEST is the Testnet.

For each network type you can change the port, seedURL and peers.   The SeedURL is a static file of known trusted peers to connect to.  The special peers are peers you want to always dial out to.

````
; --------------- Network: MAIN | TEST | LOCAL
Network                               = LOCAL
PeersFile            = "peers.json"
MainNetworkPort      = 8108
MainSeedURL          = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/mainseed.txt"
MainSpecialPeers     = ""
TestNetworkPort      = 8109
TestSeedURL          = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/testseed.txt"
TestSpecialPeers     = ""
LocalNetworkPort     = 8110
LocalSeedURL         = "https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/localseed.txt"
LocalSpecialPeers    = ""
````

Seed file example:
```
1.2.3.4:5678
2.3.4.5:6789
```

## Architecture

App <-> Controller <-> Connection <-> TCP (or UDP in future)

Controller - controller.go
This manages the peers in the network. It keeps connections alive, and routes messages 
from the application to the appropriate peer.  It talks to the application over several
channels. It uses a commandChannel for process isolation, but provides public functions
for all of the commands.  The messages for the network peers go over the ToNetwork and
come in on the FromNetwork channels.

Connection - connection.go
This struct represents an individual connection to another peer. It talks to the 
controller over channels, again providing process/memory isolation. 
