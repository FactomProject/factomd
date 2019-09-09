protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf \
--uml_out=./ \
eventmessages/generalTypes.proto \
eventmessages/factoidBlock.proto \
eventmessages/adminBlock.proto \
eventmessages/entryCredit.proto \
eventmessages/factomEvents.proto