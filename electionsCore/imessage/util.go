package imessage

// MakeMessageArray returns some arbitrary amount of messages as an array
func MakeMessageArray(messages ...IMessage) []IMessage {
	return messages
}

// MakeMessageArray returns some arbitrary amount of messages and an array as an array
func MakeMessageArrayFromArray(array []IMessage, messages ...IMessage) []IMessage {
	return append(array, messages...)
}
