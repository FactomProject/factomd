# factomd/longTest

This folder contains simulation tests that take a very long time to run.
These tests are *not run* on circle.ci and are meant for manual testing.

### LoadWith1pctDrop_test.go

Basic loadtest meant to stress a simulated network while under load.

1st part sets up  the initial network.

```
go test -v ./longTest/... -run TestSetupLoadWith1pctDrop
```

2nd part can be run repeatedly and tests booting up while under load.

```
go test -v ./longTest/... -run TestLoadWith1pctDrop
```
