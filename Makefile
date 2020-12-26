REVISION = $(shell git describe --tags)
$(info    Make factomd $(REVISION))

# Strip leading 'v'
VERSION = $(shell echo $(REVISION) | cut -c 2-)
LDFLAGS = "-s -w -X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD` -X github.com/FactomProject/factomd/engine.FactomdVersion=$(VERSION)"

build:
	go build -trimpath -ldflags $(LDFLAGS) -v
install:
	go install -trimpath -ldflags $(LDFLAGS) -v
all: factomd-darwin-amd64 factomd-windows-amd64.exe factomd-windows-386.exe factomd-linux-amd64 factomd-linux-arm64 factomd-linux-arm7

BUILD_FOLDER := build

factomd-darwin-amd64:
	env GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o $(BUILD_FOLDER)/factomd-darwin-amd64-$(REVISION)
factomd-windows-amd64.exe:
	env GOOS=windows GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o $(BUILD_FOLDER)/factomd-windows-amd64-$(REVISION).exe
factomd-windows-386.exe:
	env GOOS=windows GOARCH=386 go build -trimpath -ldflags $(LDFLAGS) -o $(BUILD_FOLDER)/factomd-windows-386-$(REVISION).exe
factomd-linux-amd64:
	env GOOS=linux GOARCH=amd64 go build -trimpath -ldflags $(LDFLAGS) -o $(BUILD_FOLDER)/factomd-linux-amd64-$(REVISION)
factomd-linux-arm64:
	env GOOS=linux GOARCH=arm64 go build -trimpath -ldflags $(LDFLAGS) -o $(BUILD_FOLDER)/factomd-linux-arm64-$(REVISION)
factomd-linux-arm7:
	env GOOS=linux GOARCH=arm GOARM=7 go build -trimpath -ldflags $(LDFLAGS) -o $(BUILD_FOLDER)/factomd-linux-arm7-$(REVISION)

.PHONY: clean

clean:
	rm -f factomd
	rm -rf build
