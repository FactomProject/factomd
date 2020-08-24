LDFLAGS = "-s -w -X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD` -X github.com/FactomProject/factomd/engine.FactomdVersion=`cat VERSION`"

build:
	go build -trimpath -ldflags $(LDFLAGS) -v
install:
	go install -trimpath -ldflags $(LDFLAGS) -v

.PHONY: clean

clean:
	rm -f factomd
