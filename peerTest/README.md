# factomd/peerTest

This folder contains tests that must be run in parallel (2 tests at a time).

These tests are useful for testing features between builds
by running 1 of each pair from previous/current builds.

Tests in this folder will *not run* on circle.ci


## BrainSwap

Run these two tests simultaneously to observe
an identy swap between go processes.

These two tests are configured to be peers.

```
nohup go test -v BrainSwapFollower_test.go &
go test -v BrainSwapNetwork_test.go
```
