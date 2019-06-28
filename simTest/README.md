# factomd/simTest

This folder contains simulation tests that can run alone in isolation.

## Add a test to circle.ci

add the filename to `ci_whitelist` file in this directory


### Run a test

NOTE: each `_test.go` file in this folder should be able to be run by itself

EX:
```
 go test -v ./simTest/BrainSwap_test.go
```

This is in contrast to testing by module (as we do with other types of unit tests)
```
go test -v ./engine/...
``
