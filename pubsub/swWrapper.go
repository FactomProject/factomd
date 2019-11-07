package pubsub

func SubWrap(wrapper ISubscriberWrapper, sub IPubSubscriber) IPubSubscriber {
	return wrapper.Wrap(sub)
}
