# factomd/peerTest

This folder contains tests that must be run in parallel (2 tests at a time).

These tests are useful for testing features between builds
by running 1 of each pair from previous/current builds.


## Add a test to circle.ci

Tests in this folder *will* be run on circle.ci if you add
add the filename to `ci_whitelist` file in this directory

## Naming Convention

Peer tests are expected to be named in A/B pairs

```
*A_test.go
*B_test.go
```

The `B` test will run in the background while the `A` test executes in the foreground.
see ./test.sh in the root of this repo for more details

Ideally `B` test only contains followers (at least initially)

This allow for some forgivness in timing when triggering these two tests manually in order to leverage a debugger.
 

## BrainSwap

Run these two tests simultaneously to observe
an identy swap between go processes.

These two tests are configured to be peers.

```
nohup go test -v BrainSwapA_test.go &
go test -v BrainSwapB_test.go
```
