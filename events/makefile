generate:
	protoc \
	    --gofast_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mopsee/protobuf/opsee.proto=github.com/opsee/protobuf/opseeproto,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:./eventmessages/generated \
	    --proto_path=$(GOPATH)/src:. eventmessages/*.proto
clean:
	rm -f eventmessages/*.pb.go
	rm -f eventmessages/*pb_test.go
.PHONY:
	clean
	generate
