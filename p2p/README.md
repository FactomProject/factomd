# Standalone Implementation of Factom's P2P Network

# Summary
This package implements a partial gossip network, limited to peer connectivity and delivery of messages. The application is responsible for triggering fanout, rounds, and message repetition detection. 

This is a complete rework of the code that aims to get rid of uncertainty of connections as well as polling strategies and also add support for handling multiple protocol versions. This will open up opportunities to improve network flow through generational increments without breaking backward compatibility.

Goals:
* Fully configurable, independent, isolated instances
* Add handshake process to connections to distinguish and identify nodes
* Ability to handle multiple protocol schemes starting from version 9
* Reduced network packet overhead in scheme version 10
* Simple integration
* Easier to read code with comprehensive documentation

# Motivation

* Peers are not identifiable beyond ip address. Multiple connections from the same node are not differentiated and assigned random numbers
* Peers have inconsistent information (some data is only available after the first transmission)
* Nodes depend on the seed file to join the network even across reboots; peers are not persisted
* Connections are polled in a single loop to check for new data causing unnecessary delays
* Connections have a complicated state machine and correlation with a Peer object
* Complicated program flow with a lot of mixed channels, some of which are polled

Tackling some of these would require significant overhaul of the existing code to the point where it seemed like it would be easier to start from scratch. 

# Specification

## Terminology
* **Peer**/**Node**: A node is any application that uses the same p2p protocol to connect to the network. A peer is a node that is connected to us.
* **Endpoint**: An internet location a node can connect to. Consists of an IP address and port.
* **Peer Hash**: Each peer receives a unique peer hash using the format `[ip]:[port] [hex nodeid]`, where `[ip]` is taken from the TCP connection itself, and `[port]` and `[hex nodeid]` are determined from the peer's handshake.
* **Parcel**: A parcel is a container for network messages that has a type and a payload. Parcels of the application type are delivered to the application, others are used internally.

## Package Structure

The P2P package consists of these major components:

1. **Network** is the publicly visible interface that applications use to interact with the p2p network. It's initialized with a **Configuration** object and other components all have a back reference to the network so they can interact with each other.
2. **Controller** is the heart of the p2p package. It's split over several files, separated by area of responsibility. The controller handles accepting/creating new connections, peer management, and data routing. The controller has a **PeerStore** that holds all active peer connections.
3. **Peer**s are connections to another node. Each peer has an active TCP connection and a **Protocol**, which translates **Parcel**s into a format the peer can understand.

## Overview

![Quick Overview](https://camo.githubusercontent.com/070ef686795dbc8650ba5a29a8237e196047e4e6/68747470733a2f2f692e696d6775722e636f6d2f665137675855712e706e67)

### Lifecycles

#### Peer

The foundation of a peer is a TCP connection, which are created either through an incoming connection, or a dial attempt. The peer object is initialized and the TCP connection is given to the handshake process (see below for more info). If the handshake process is *unsuccessful*, the tcp connection is torn down and the peer object is destroyed. If the handshake process is sucessful, the peer's read/send loops are started and it sends an *online notification* to the *controller's status channel*.

Upon receiving the online notification, the controller will include that peer in the *PeerStore* and make it available for routing. The read loop reads data from the connection and sends it to the controller. The send loops takes data from the controller and sends it to the connection. If an error occurs during the read/write process or the peer's **Stop()** function is called, the peer stops its internal loops and also sends an *offline notification* to the *controller's status channel*.

Upon receiving the offline notification, the controller will remove that peer from the *PeerStore* and destroy the object.

If the controller dials the same node, a new Peer object will be created rather than recycling the old one. If an error during read or write occurs, the Peer will call its own **Stop()** function.

#### Parcel (Application -> Remote Node)

The application creates a new parcel with the *application type*, a *payload*, and a *target*, which may either be a peer's hash, or one of the predefined flags: *Broadcast*, *Full Broadcast*, or *Random Peer*. The parcel is given to the **ToNetwork** channel. 

The controller *routes* all parcels from the ToNetwork channel to individual Peer's *send channels* based on their target:
1. Peer's Hash: parcel is given directly to that peer
2. Random Peer: a random peer is given the parcel
3. Broadcast: 16 peers (config: `Fanout`) are randomly selected from the list of non-special peers. Those 16 peers and all the special peers are given the parcel
4. Full Broadcast: all peers are given the parcel

Each Peer monitors their send channel. If a parcel arrives, it is given to the *Protocol*. The *Protocol* reads the parcel and creates a corresponding *protocol message*, which is then written to the connection in a manner dictated by the protocol. For more information on the protocols, see below.

#### Parcel (Remote Node -> Application)

A Peer's *Protocol* reads a *protocol message* from the connection and turns it into a *Parcel*. The parcel is then given to the controller's *peerData channel*. The controller separates *p2p parcels* from *application parcels*. Application parcels are given to the **FromNetwork** channel without any further processing.


## Protocol

### CAT Peering

The CAT (Cyclic Auto Truncate) is a cyclic peering strategy to prevent a rigid network structure. It's based on rounds, during which random peers are indiscriminately dropped to make room for new connections. If a node is full it will reject incoming connections and provide a list of alternative nodes to try and peer with. There are three components, **Rounds**, **Replenish**, and **Listen**.

### Round

Rounds run once every 15 minutes (config: `RoundTime`) and does the following:

1. Persist current peer endpoints and bans in the peer file (config: `PersistFile`)
2. If there are more than 30 peers (config: `DropTo`), it randomly selects non-special peers to drop to reach 30 peers.

### Replenish

The goal of Replenish is to reach 32 (config: `TargetPeers`) active connections. If there are 32 or more connections, Replenish waits. Otherwise, it runs once a second.

Once on startup, if the peer file is less than an hour old (config: `PersistAge`) Replenish will try to re-establish those connections first. This improves reconnection speeds after rebooting a node.

The first step is to pick a list of peers to connect to:
If there are fewer than 10 (config: `MinReseed`) connections, replenish retrieves the peers from the seed file to connect to. Otherwise, it sends a Peer-Request message to a *random* peer in the connection pool. If it receives a response within 5 seconds, it will select **one** random peer from the provided list.

The second step is to dial the peers in the list. If a peer in the list rejects the connection with alternatives, the alternatives are added to the list. It dials to the list sequentially until either 32 connections are reached, the list is empty, or 4 connection attempts (working or failed) have been made.

### Listen

When a new TCP connection arrives, the node checks if the IP is banned, if there are more than 36 (config: `Incoming`) connections, or (if `conf.PeerIPLimitIncoming` > 0) there are more than conf.PeerIPLimitIncoming connections from that specific IP. If any of those are true, the connection is **rejected**. Otherwise, it continues with a **Handshake**.

Peers that are rejected are given a list of 3 (conf: `PeerShareAmount`) random peers the node is connected to in a Reject-Alternative message.

### Handshake

The handshake follows establishing a TCP connection. The "outgoing" handshake is performed by the node dialing into another node. The format of the Handshake struct is protocol-dependent but it contains the following information:

| Name | Type | Description |
|------|------|-------------|
| Network | NetworkID | The network id of the network (ie, MainNet = `0xfeedbeef`) (conf: `Network`) |
| Version | uint16 | The version of the protocol we want to use. (conf: `ProtocolVersion`) | 
| Type | ParcelType | For V10 and up, this is either type "Handshake" (`0x8`) or "Reject with Alternatives" (`0x9`). For V9, this is "Peer Request" (`0x3`) |
| NodeID | uint32 | An application-defined value that can persist across restarts (conf: `NodeID`) |
| ListenPort | string | The port the node is defined to listen at (conf: `ListenPort`) | 
| Loopback | uint64 | A unique nonce to detect loopback connections | 
| Alternatives | slice of Endpoints | If the connection is rejected, a list of alternative endpoints to connect to | 

#### Outgoing Handshake

1. Set a deadline of 10 seconds (conf: `HandshakeTimeout`)
2. Select the desired protocol
3. Encode the handshake data using the desired protocol
4. Send the handshake
5. Wait for a response
6. Attempt to identify the encoding scheme and decode the first message as handshake
7. Validate the handshake response to see if the network, loopback, and types are desired
8. Use the protocol that matches the reply's encoding and version

If any step fails, the handshake is considered failed. 

#### Incoming Handshake

1. Set a deadline of 10 seconds (conf: `HandshakeTimeout`)
2. Attempt to identify the encoding scheme and decode the first message as handshake
3. Validate the handshake response to see if the network and types are desired
4. (Optional) Propose an alternative protocol
5. Create a handshake, copying the loopback value from 3.
6. Send the handshake

If any step fails, the handshake is considered failed. In most cases, the node should use the same protocol that it received the handshake for. However, p2p1 nodes are unable to understand the newer protocol and will always send a message containing protocol version 9. In this case, nodes are expected to downgrade the protocol to V9. 

### 9

Protocol 9 is the legacy (Factomd v6.6 and lower) protocol with the ability to split messages into parts disabled. V9 has the disadvantage of sending unwanted overhead with every message, namely Network, Version, Length, Address, Part info, NodeID, Address, Port. In the old p2p system this was used to post-load information but now has been shifted to the handshake.

Data is serialized via Golang's gob.

### 10

Protocol 10 is the slimmed down version of V9, containing only the Type, CRC32 of the payload, and the payload itself. Data is also serialized via Golang's gob. The handshake is encoded using V9's format.

### 11

Protocol 11 uses Protobuf ([protocolV11.proto](protocolV11.proto)) to define the Handshake and message. To signal that the connection is used for V11, the 4-byte sequence `0x70627566` (ASCII for "pbuf") is transmitted first. Protobufs are transmitted by sending the size of the marshalled protobuf first, encoded as uint32 in Big Endian format, followed by the protobuf byte sequence itself.

V11 has a maximum parcel size of 128 Mebibytes.

## Usage

### Setting up a Network

In order to set up a network, you need two things: a network id, and a bootstrap file.

The network ID can be generated with `p2p.NewNetworkID(string)`, with your preferred name as input. For example, "myNetwork" results in `0x29cb7175`. There are also predefined networks, like `p2p.MainNet` that are used for Factom specific networks.

The bootstrap seed file contains the addresses of your seed nodes, the ones that every new node will attempt to connect to. Plaintext, one `ip:port` address per line. An example is [Factom's mainnet seed file](https://raw.githubusercontent.com/FactomProject/factomproject.github.io/master/seed/mainseed.txt):
```
52.17.183.121:8108
52.17.153.126:8108
52.19.117.149:8108
52.18.72.212:8108
52.19.44.249:8108
52.214.189.110:8108
34.249.228.82:8108
34.248.202.6:8108
52.19.181.120:8108
34.248.6.133:8108
```


### Connecting to a Network

First, you need to create the configuration:

```go
config := p2p.DefaultP2PConfiguration()
config.Network = p2p.NewNetworkID("myNetwork")
config.SeedURL = "http://url/of/seed/file.txt"
config.PersistFile = "/path/to/peerfile.json"
```

The default values are derived from Factom's network and described in the [Configuration file](configuration.go). The `config.NodeID` is a unique number tied to a node's ip and port. The same node should use the same NodeID between restarts, but two nodes running at the same time and using the same ip and listen port should have different NodeIDs. The latter is the case if you have multiple nodes behind a NAT connecting to a public network.

The `config.PersistFile` setting can be blank to not save peers and bans to disk. Enabling this makes a node able to restart the network faster and re-establish old connections.

### Starting the Network

Once you have the config, the rest is easy.

```go
network, err := p2p.NewNetwork(config)
if err != nil {
    // handle err, typically related to the peer file or unable to bind to a listen port
}

network.Run() // nonblocking, starts its own goroutines
```

You can start reading and writing to the network immediately, though no peers may be connected at first. You can check how many connections are established via `network.Total()`.

### Reading and Writing

To send an application message to the network, you need to create a Parcel with a **target** and a **payload**:

```go
parcel := p2p.NewParcel(p2p.Broadcast, byteSequence)
network.ToNetwork.Send(parcel)
```

The target can be either a peer's hash, or one of the predefined flags of `p2p.RandomPeer`, `p2p.Broadcast`, or `p2p.FullBroadcast`. The functions of these are described in detail in the Lifecycle section "Parcel (Application -> Remote Node)". The p2p package is data agnostic and any interpretation of the byte sequence is left up to the application.

To read incoming Parcels:

```go
for parcel := range network.FromNetwork.Reader() {
    // parcel.Address is the sender's peer hash
    // parcel.Payload is the application data
}
```

If you want to return a message to the sender, use the parcel's Address as the **target** of a new parcel.

