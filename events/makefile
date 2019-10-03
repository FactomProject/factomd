GOCMD:=$(shell which go)
GOLIST:=$(GOCMD) list
GOGET:=$(GOCMD) get
PACKAGES:=$(shell $(GOLIST) ./...)

generate:
	protoc \
	    --gofast_out=Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:./eventmessages/generated \
	    --proto_path=$(GOPATH)/src:. eventmessages/*.proto
clean:
	rm -f eventmessages/generated/eventmessages/*.pb.go
	rm -f eventmessages/generated/eventmessages/*pb_test.go
.PHONY:
	clean
	generate

.PHONY: test
test:
	echo "mode: count" > coverage.out
	for PKG in $(PACKAGES); do \
		$(GOCMD) test -v -covermode=count -coverprofile=profile.out $$PKG > tmp.out; \
		cat tmp.out; \
		if grep -q "^--- FAIL" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		elif grep -q "build failed" tmp.out; then \
			rm tmp.out; \
			exit; \
		fi; \
		if [ -f profile.out ]; then \
			cat profile.out | grep -v "mode:" >> coverage.out; \
			rm profile.out; \
		fi; \
	done; \
	rm tmp.out || true;

.PHONY: view-covered
view-covered: test
	$(GOCMD) tool cover -html=coverage.out

.PHONY: dev-deps
dev-deps:
	GO111MODULE=off $(GOGET) -v -u \
		github.com/golang/protobuf/protoc-gen-go	\
		github.com/gogo/protobuf/protoc-gen-gofast