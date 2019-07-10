package engine

// Build sets the factomd build id using git's SHA
// Version sets the semantic version number of the build
// $ go install -ldflags "-X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD` -X github.com/FactomProject/factomd/engine.=`cat VERSION`"
// It also seems to need to have the previous binary deleted if recompiling to have this message show up if no code has changed.
// Since we are tracking code changes, then there is no need to delete the binary to use the latest message
var Build string
var FactomdVersion string = "BuiltWithoutVersion"
