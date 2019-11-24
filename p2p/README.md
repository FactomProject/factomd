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

## Protocol

### CAT Peering

The CAT (Cyclic Auto Truncate) is a cyclic peering strategy to prevent a rigid network structure. It's based on rounds, during which random peers are indiscriminately dropped to make room for new connections. If a node is full it will reject incoming connections and provide a list of alternative nodes to try and peer with. There are three components, **Rounds**, **Replenish**, and **Listen**.

### Round

Rounds run once every 15 minutes (config: `RoundTime`). It gets a list of peers and if there are more than 30 peers (config: `Drop`), it randomly selects non-special peers to drop to reach 30 peers.

### Replenish

The goal of Replenish is to reach 32 (config: `Target`) active connections. If there are 32 or more connections, Replenish waits. Otherwise, it runs once a second.

The first step is to pick a list of peers to connect to:
If there are fewer than 10 (config: `MinReseed`) connections, replenish retrieves the peers from the seed file to connect to. Otherwise, it sends a Peer-Request message to a *random* peer in the connection pool. If it receives a response within 5 seconds, it will select **one** random peer from the provided list.

The second step is to dial the peers in the list. If a peer in the list rejects the connection with alternatives, the alternatives are added to the list. It dials to the list sequentially until either 32 connections are reached, the list is empty, or 4 connection attempts (working or failed) have been made.


### Listen

When a new TCP connection arrives, the node checks if the IP is banned, if there are more than 36 (config: `Incoming`) connections, or (if `conf.PeerIPLimitIncoming` > 0) there are more than conf.PeerIPLimitIncoming connections from that specific IP. If any of those are true, the connection is **rejected**. Otherwise, it continues with a **Handshake**.

Peers that are rejected are given a list of 3 (conf: `PeerShareAmount`) random peers the node is connected to in a Reject-Alternative message.

### Handshake

The handshake starts with an already established TCP connection.

1. A deadline of 10 seconds (conf: `HandshakeTimeout`) is set for both reading and writing
2. Generate a Handshake containing our preferred version (conf: `ProtocolVersion`), listen port (conf: `ListenPort`), network id (conf: `Network`), and node id (conf: `NodeID`) and send it across the wire
3. Blocking read the first message
4. Verify that we are in the same network
5. Calculate the minimum of both our and their version
6. Check if we can handle that version (conf: `ProtocolVersionMinimum`) and initialize the protocol adapter
7. If it's an outgoing connection, check if the Handshake is of type RejectAlternative, in which case we parse the list of alternate endpoints

If any step fails, the handshake will fail. 

For backward compatibility, the Handshake message is in the same format as protocol v9 requests but it uses the type "Handshake". Nodes running the old software will just drop the invalid message without affecting the node's status in any way.

### 9

Protocol 9 is the legacy (Factomd v6.4 and lower) protocol with the ability to split messages into parts disabled. V9 has the disadvantage of sending unwanted overhead with every message, namely Network, Version, Length, Address, Part info, NodeID, Address, Port. In the old p2p system this was used to post-load information but now has been shifted to the handshake.

Data is serialized via Golang's gob.

### 10

Protocol 10 is the slimmed down version of V9, containing only the Type, CRC32 of the payload, and the payload itself. Data is also serialized via Golang's gob.