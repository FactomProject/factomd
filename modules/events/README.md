# Live Feed events service

FactomD now supports a new Live Feed API which, as the name suggests, it can post real-time information 
about what’s happening on the Factom network as well as in the factomd node itself.

The API consists of a event emitter client which the is inside factomd, and the second one is 
an event receiver server. The (tcp/udp) client inside factomd dumps all the data into a network pipe, 
with limited filtering capabilities (just a static setting whether to include the content and external id’s in the datastream or not.)

## How to enable live-feed
To enable and configure the first layer Live Feed in factomd there is a new property section in factomd.conf:
```
; ------------------------------------------------------------------------------
; Configuration for live feed API
; ------------------------------------------------------------------------------
[LiveFeedAPI]
EnableLiveFeedAPI                     = true
EventReceiverProtocol                 = tcp
EventReceiverHost                     = 127.0.0.1
EventReceiverPort                     = 8040
EventSenderPort                       = 8041
EventFormat                           = protobuf
EventSendStateChange                  = false
EventBroadcastContent                 = OnRegistration 
EventReplayDuringStartup              = false
PersistentReconnect                   = false
```

Here is an overview of these options:

| Property                          | Description                                                                         | Values      |
| --------------------------------- | ----------------------------------------------------------------------------------- | ----------- |
|  EnableLiveFeedAPI                | Turn the Live Feed API on or off                                            | true &#124; false
|  EventReceiverProtocol            | The network protocol that is used to send event messages over the network.     | tcp &#124; udp |
|  EventReceiverHost                | The receiver endpoint host.                                                | DNS name &#124; IP address |
|  EventReceiverPort                | The receiver endpoint port.                                                  | port number |
|  EventSenderPort                  | The client or sender port.                                                   | port number |
|  EventFormat                      | The output format in which the event sent.                                      | protobuf &#124; json |
|  EventSendStateChange             | It’s possible to choose whether the chain and entry commit registrations should only be sent once, followed by state change events vs resending them for every state change. The first option reduces overhead & network traffic, but requires the implementer to track which state changes belong to which chain or entry.| true &#124; false |
|  EventBroadcastContent            | This option will determine whether the external ID’s and content will be included in the event stream. There are three level settings for this. Please note that the combination of EventSendStateChange = false and EventBroadcastContent=always, will resend all data on every state change. The maximum content size per entry is only 10KB, however with a large number of transactions per second this may add up to an undesirable amount of data. | always &#124; once &#124; never |
|  EventReplayDuringStartup         | At startup factomd can replay all the events that were stored since that last fastboot snapshot. Use this property to turn that on/off.   | true &#124; false |

The same properties can be overridden by command line parameters which are the same as above but lowercase.
The retry mechanism of the first layer is pretty strict. When a receiver is down or for some reason unresponsive it will retry to connect 3 times. If a receiver is not up by then, it will keep retrying to restore the connection every 5 minutes, but in the meantime it will start dropping the events until the receiver is back up. For mission critical use-cases there are prometheus counters in place:
* **factomd_livefeed_not_send_counter** - the number of events that should be send, but couldn't be delivered to the receiver.
* **factomd_livefeed_dropped_from_queue**_counter - the number of events that couldn't be send, because the queue is full.

Along with the block height inside the events that are emitted, these are the tools with which the receiver can detect if the feed is complete. It’s the responsibility of the receiver to request missing entries/blocks when required.