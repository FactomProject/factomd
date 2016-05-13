
# Factom P2P Networking




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
