#!/bin/bash

# Compiles static files into binary
# and runs go install

staticfiles -o files/statics/statics.go Web/statics
staticfiles -o files/templates/templates.go Web/templates

cd $GOPATH/src/github.com/FactomProject/factomd
go install -ldflags "-X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD` -X github.com/FactomProject/factomd/engine.FactomdVersion=`cat VERSION`" || cerr=1