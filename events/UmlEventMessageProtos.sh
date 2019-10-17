export PLANTUML_LIMIT_SIZE=8192
export PATH=/home/sander/DEV/go/src/github.com/protoc-gen-uml/target/universal/stage/bin:$PATH
protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/gogo/protobuf/protobuf \
--uml_out=./ \
eventmessages/sharedTypes.proto \
eventmessages/factoidBlock.proto \
eventmessages/adminBlock.proto \
eventmessages/factomEvents.proto
plantuml ./complete_model.puml