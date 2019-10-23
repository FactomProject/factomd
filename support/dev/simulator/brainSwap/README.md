# BrainSwap Simulator Test

Run this test to manually verify we can identity swap between versions of factomd.

1. Copy FactomD binaries to compare into ./v0 and ./v1
2. Start `./v0/run.sh` in a separate terminal
3. Manually use the simulator to construct an authority set
4. Run `./v0/swap.sh` to change identities
5. Start `./v1/run.sh` in a separate terminal
6. Run `./v1/swap.sh` to change identities
7. Observe the test run - swap should occur at block 10 min 0
8. Wait to verify network is healthy then Kill nodes.


## Alternatively: Using Unit test

`./test/run.sh` runs a go unit test that will create an authority set

use 'test' in place of 'v0' when running the steps above.

