# Internal PubSub

This pubsub package is used for internal communication between modules. In the future, this pubsub can be abstracted to work over a tcp connection. The basic idea is to create a framework for communication between modules. This communication is 1 way, where we have Publishers writing data, and Subscribers reading data.

The original idea was a simple channel based pub/sub model that all modules would use, however, this simple idea was quickly proved to be too general/simple. The second solution was to generalise the interface between publishers and subscribers, but not to generalise the interface to **use** the publisher or subscriber.

Because of this, we have multiple implementations of publishers and subscribers.


# Usage

The `examples` directory has example uses of the pubsub. The basic usage is to use the `PubFactory` and `SubFactory`. All publishers use a global registry and directory style path naming. To make a new base publisher, and subscribe to it:
```golang
    // establish a publisher
    //                             Type             Path
    publisher := pubsub.PubFactory.Base().Publish("/source")

    // establish a subscriber to read the value
    subscriber := pubsub.SubFactory.Value().Subscribe("/source")
    
    publisher.Write(1)
    fmt.Println(subscriber.Read()) // 1
```

Using wrappers is not that difficult either. If you want more than 1 writer to a publisher, you can use the multi wrap. There are a few types of wrappers, and their usage varies, but they will always be given as optional parameters to the initialization of the pub or sub.
```golang
	// establish a publisher
	//                             Type             Path
	publisher1 := pubsub.PubFactory.Base().Publish("/source", pubsub.PubMultiWrap())
	publisher2 := pubsub.PubFactory.Base().Publish("/source", pubsub.PubMultiWrap())

	// establish a subscriber to read the value
	subscriber := pubsub.SubFactory.Value().Subscribe("/source")

	publisher1.Write(1)
	fmt.Println(subscriber.Read()) // 1

	publisher2.Write(2)
	fmt.Println(subscriber.Read()) // 1
```

It should be noted threaded publishers have a `Start()` function on them. This should be called when making the publisher in it's own GoRoutine. The Multi wrapper ensures the start function is always called once, so if you pass the Multi wrapper, it makes `Start()` idempotent.

## Publishers

Publishers are called by the caller to write data to all subscribers.

- Base: A simple base publisher that uses mutexs to control threadsafe writes. All writes are executed in the thread context of the caller and go to all subscribers.
- Threaded: A publisher that executes the writes in a thread context specific to the publisher. All write calls are buffered in a channel, and go to all subscribers.
- RoundRobin: A threaded publisher that writes each piece of data to 1 subscriber on a round robin basis.
- MsgSplit: A threaded publisher that only sends the writes to 1 subscriber based on the msghash of the data. This shards the data such that the same msg will always go to the same subscriber.

## Subscribers

Subscribers are written to by publishers, and read by the caller.

- Base: Has the base functionality for all subscribers, but all writes are no-ops, meaning it does not handle incoming data.
- Channel: Buffers incoming writes to a channel.
- Counter: Counts the number of writes, drops the data.
- Value: Stores the last write.


# Wrappers

There is additional functionality that can be applied to any publisher/subscriber. This is optional functionality. When publishing or subscribing, these are optional parameters

## Publisher Wrappers

- Multi: Adding a multi wrapper allows multiple threads to publish to the same publisher.

## Subscriber Wrappers

- Callback: Add a callback on writes, and optionally ignore messages if the callback returns an error. This can be used to make a primitive filter.
- Context: Allows an external context to be bound to the lifecycle of the subscriber.
- MsgFilter: Only accept IMsgs of a certain type.
