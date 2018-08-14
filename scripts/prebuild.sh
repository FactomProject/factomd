#!/bin/bash
set -x
echo package engine                                 > engine/version.go 
echo var Build string = \"`git rev-parse HEAD`\" >> engine/version.go  
echo var FactomdVersion string = \"`cat VERSION`\"  >> engine/version.go  

