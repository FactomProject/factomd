#!/bin/bash

# Compiles static files into binary
# and runs go install
staticfiles -o files/files.go Web/
cd $GOPATH/src/github.com/FactomProject/factomd
go install -ldflags "-X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD`" || cerr=1