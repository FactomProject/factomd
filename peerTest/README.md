# factomd/peerTest

This folder contains tests that must be run in parallel (2 tests at a time).

These tests are useful for testing features between builds
by running 1 of each pair from previous/current builds.

Tests in this folder *will* be run on circle.ci.

## Naming Convention

  Peer tests are expected to be named in Follower/Network pairs

  ```
  *Follower_test.go
  *Network_test.go
  ```

  The network test will run in the background while the follower test executes in the foreground.
  see ./test.sh in the root of this repo for more details

## BrainSwap

Run these two tests simultaneously to observe
an identy swap between go processes.

These two tests are configured to be peers.

```
nohup go test -v BrainSwapFollower_test.go &
go test -v BrainSwapNetwork_test.go
```
