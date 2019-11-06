package pubsub

func Wrap(wrapper ISubscriberWrapper, sub IPubSubscriber) IPubSubscriber {
	return wrapper.Wrap(sub)
}
