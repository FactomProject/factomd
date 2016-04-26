
# Factom P2P Networking




## Architecture

App <-> Peer (p2pPeer) <-> Protocol <-> Wire <-> TCP (or UDP in future)

Wire  
- This is the lowest layer. It manages the sending and recieving of messages, ensuring that complete messages are transmitted/recieved.
- Provides some sort of checksum for messages
- Messages are sent in big endian format.
- Over the air message format is:
    - Magic Cookie (buffer over/underflow protection)
    - Message Length 
    - Checksum
    - Payload
- Speaks with higher levels using the Wire struct

Protocol
- One level up, allows for running multiple protocols over a connection. (eg: heartbeat, factom, stats, etc.)
- Protocols are registered with the network, along with a protocol id. Each message indicates which protocol it belongs to.
- The network manager 


Example code:

Read (and other magnos stuff)
https://github.com/go-mangos/mangos/blob/master/conn.go#L176