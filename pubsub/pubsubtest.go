package pubsub

func Publish_Generic(name string, object interface{}) {

}

func Subscribe_Generic(name string) interface{} {
	var x interface{}

	return x
}

func dummy() {
	var x int = 1
	var foo = Publish_int("foo", x)

	var bar = Subscribe_int("foo")

}
