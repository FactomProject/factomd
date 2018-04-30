# Factom

[![Build Status](https://travis-ci.org/FactomProject/factomd.svg?branch=develop)](https://travis-ci.org/FactomProject/factomd)
[![CircleCI](https://circleci.com/gh/FactomProject/factomd/tree/develop.svg?style=shield)](https://circleci.com/gh/FactomProject/factomd/tree/develop)

Factom is an Open-Source project that provides a way to build applications on the Bitcoin blockchain. 

Factom began by providing proof of existence services, but then move on to provide proof of existence of transforms. A list of such entries can be thought of as a Factom Chain.  Factom can be used to implement private tokens, smart contracts, smart properties, and more.

Factom leverages the Bitcoin Blockchain, but in a way that minimizes the amount of data actually inserted in the Blockchain. Thus it provides a mechanism for creating Bitcoin 2.0 services for the trading of assets, securities, commodities, or other complex applications without increasing blockchain "pollution".


## Getting Started

You need to set up Go environment with golang 1.10 or higher. You also need git.  See the [Install from source](https://github.com/FactomProject/FactomDocs/blob/master/installFromSourceDirections.md) directions for more details and wallet installation instructions.

### Install the dependency namagement program

First check if golang 1.10 or higher is installed.  some operationg systems install older versions.

`go version` should return something like
`go version go1.10.1 linux/amd64`

Next install Glide, which gets the dependencies for factomd and places them in the `$GOPATH/src/github/FactomProject/factomd/vendor` directory.

`go get -u github.com/Masterminds/glide`

### Install factomd Full Node

```
cd $GOPATH/src/github.com/FactomProject/
git clone https://github.com/FactomProject/factomd $GOPATH/src/github.com/FactomProject/factomd

# clear the glide cache since it has been known to cause errors
# deleting the $GOPATH/src/github.com/FactomProject/factomd/vendor  directory may be useful too
glide cc
cd $GOPATH/src/github.com/FactomProject/factomd
# this command download the dependencies and sets them to the right version
glide install
# install factomd with either the install.sh script or:
go install -ldflags "-X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD` -X github.com/FactomProject/factomd/engine.FactomdVersion=`cat VERSION`" -v
# you can optionally use a config file to run in a non-standard mode
# mkdir -p ~/.factom/m2/
# cp $GOPATH/src/github.com/FactomProject/factomd/factomd.conf ~/.factom/m2/

```



## M2 Simulator 

factomd can run a simulated network via the commandline.  This allows testing of much more complicated networks than would be possible otherwise.   The  simulator is very extensible, so new features will be added as we go along.

### Databases and Journaling

When you run factomd, if no database directory exists in the m2 directory, one is created.  All databases are created under the ~/.factom/m2/database directory.  These databases are named "bolt" for the main factom node, and "blotSimNNN" where NNN is the node number for Simulation nodes cloned from the main factom node.  Inside each of these directories, a network folder is created.  LOCAL (for test networks built on your own machine), MAIN for the main network, and TEST for the test network.   This allows swapping between networks without concern for corupting the databases.

As M2 runs, journal files are created in the database directory. All messages are journaled for all nodes in the simulator.  This gives the ability to "rerun" a message sequence to debug observed issues. When factomd is restarted, all journal files for those nodes are reset.

Below is a discription of how to run journal files.

### Flags to control the simulator

To get the current list of flags, type the command:

	factomd -h

Which will get you something like:
	
    //////////////////////// Copyright 2017 Factom Foundation
    //////////////////////// Use of this source code is governed by the MIT
    //////////////////////// license that can be found in the LICENSE file.
    Go compiler version: go1.6.2
    Using build: 
    len(Args) 2
    Usage of factomd:
    -blktime int
            Seconds per block.  Production is 600.
    -clonedb string
            Override the main node and use this database for the clones in a Network.
    -count int
            The number of nodes to generate (default 1)
    -db string
            Override the Database in the Config file and use this Database implementation
    -drop int
            Number of messages to drop out of every thousand
    -exclusive
            If true, we only dial out to special/trusted peers.
    -folder string
            Directory in .factom to store nodes. (eg: multiple nodes on one filesystem support)
    -follower
            If true, force node to be a follower.  Only used when replaying a journal.
    -journal string
            Rerun a Journal of messages
    -leader
            If true, force node to be a leader.  Only used when replaying a journal. (default true)
    -net string
            The default algorithm to build the network connections (default "tree")
    -node int
            Node Number the simulator will set as the focus
    -p2pPort string
            Port to listen for peers on. (default "8108")
    -peers string
            Array of peer addresses. 
    -port int
            Address to serve WSAPI on
    -prefix string
            Prefix the Factom Node Names with this value; used to create leaderless networks.
    -profile string
            If true, turn on the go Profiler to profile execution of Factomd
    -rotate
            If true, responsiblity is owned by one leader, and rotated over the leaders.
    -runtimeLog
            If true, maintain runtime logs of messages passed.
    -test.bench string
            regular expression to select benchmarks to run
    -test.benchmem
            print memory allocations for benchmarks
    -test.benchtime duration
            approximate run time for each benchmark (default 1s)
    -test.blockprofile string
            write a goroutine blocking profile to the named file after execution
    -test.blockprofilerate int
            if >= 0, calls runtime.SetBlockProfileRate() (default 1)
    -test.count n
            run tests and benchmarks n times (default 1)
    -test.coverprofile string
            write a coverage profile to the named file after execution
    -test.cpu string
            comma-separated list of number of CPUs to use for each test
    -test.cpuprofile string
            write a cpu profile to the named file during execution
    -test.memprofile string
            write a memory profile to the named file after execution
    -test.memprofilerate int
            if >=0, sets runtime.MemProfileRate
    -test.outputdir string
            directory in which to write profiles
    -test.parallel int
            maximum test parallelism (default 1)
    -test.run string
            regular expression to select tests and examples to run
    -test.short
            run smaller test suite to save time
    -test.timeout duration
            if positive, sets an aggregate time limit for all tests
    -test.trace string
            write an execution trace to the named file after execution
    -test.v
            verbose: print additional output
    -timedelta int
            Maximum timeDelta in milliseconds to offset each node.  Simulates deltas in system clocks over a network.

The flags that begin with "test." are supplied by the profiling package installed.  The flags that relate to running factomd and the simulator are the following, with a little more explaination.  That follows below.

### Simulator Commands

While the simulator is running, you can perform a number of commands to poke at, and examine the state of factomd:

* aN -- Dump the Admin block at directory block height N. So like a1 or a213. Currently just dumps the JSON. 
* fN -- Dump the Factoid block at directory block height N.  So like f5 or f1203.  Pretty prints.
* dN -- Dump the Directory block at directory block height N.  d4 or d21230
* <enter> -- gives the state of all nodes in the simulated network.
* D -- Dumps all the messages in the system to standard out, including the directory blocks and the process lists.
* l -- Attempt to make this server a leader (must have a valid identity to become one) 
* o -- Attempt to make this server an auditor (must have a valid identity to become one) 
* s -- Show the state of all nodes as their state changes in the simulator.
* i -- Shows the current identities being monitored for changes
* u -- shows the current authorities (federated/audit servers)
* N -- typing a node number shifts focus.  You now are talking to said node from the CLI or wallet

### Simulator Commands Continued -- Identity
M2 requires servers to have identities if they wish to have the ability to become a federated or audit server. To create an identity, entries must be entered into the blockchain, so controls were added to the simulator to assist in the creation and attachment of identities. Identites take about a minute to generate on a (macbook pro laptop) to meet the proper requirements, so a stack of identities are pregenerated to make testing easier.

How the simulator controls work. First every instance of factomd will share the same stack of identites. Each instance will also have a local pool of identities they can use and attach to their nodes. To load their local pool of identities, they can pop identities off the shared stack, then attach the next open identity in their local identity pool to the current node:

* gN -- Moves N identities from the shared stack to local identity pool
  * Be mindful everyone shares the stack and it can run out.
* t -- Attaches the next identity in the local pool that has not been taken to the current node

<i>The 'gN' command will load entry credits into the zeros entry credit wallet to fund all identity sim controls if the wallet is low on funds. </i>

### Launching Factomd
 
Personally I open two consoles.  I run factomd redirected to out.txt, and in another console I run tail -f out.txt.

So in one console I run a command like:

	factomd -count=10 -net=tree > out.txt
	
And in another console I run:

	tail -f out.txt
	
Then I type commands in the first console as described above, and see the output in the second.  Also messages and errors will show up in the first console (leaving the second console with simple output from factomd).

### -count

The command:
	
	factomd -count=10

Will run the simulator with the configuration found in ~/.factom/m2/factom.conf with 10 nodes, named fnode0 to fnode9.  fnode0 will be the leader.  When you hit enter to get the status, you will see an "L" next to leader nodes.

### -db

You can override the database implementation used to run the simulator.  The options are "Bolt", "LDB", and "Map".  "Bolt" will give you a bolt database, "LDB" will get you a LevelDB database, and "Map" will get you a hashtable based database (i.e. nothing stored to disk).  The most common use of -db is to specify a "Map" database for tests that can be easily rerun without concern for past runs of the tests, or of messing up the database state.

Keep in mind that Map will still overwrite any journals.  For example, you can run a 10 node Factom network in memory with the following command:

	factoid -count=10 -db=Map
	
### -follower

At times it is nice to force factomd to launch a follower rather than a leader (or the other way around).  Especially when playing back a journal of messages to investigate why a server got into a particular state.  So suppose we have a leader journal leader.log.  We could execute that log with this command:

	factoid -journal=leader.log -follower=false -db=Map
	
Or if we had a follower log, follower.log, we could execute it with:

	factoid -journal=follower.log -follower=true -db=Map

##3 -journal

Running factomd creates a journal file for every node in the ~/.factom/m2/database/ directory, of the form journalNNN.log where NNN is the node number.   So if there is a failure or a desire to rerun the same message stream as a test, this can be done by copying the journalNNN.log files, then running them.  For example, suppose we ran a 10 node network and did some testing:

	factomd -count=10 
	<testing done>
	
Now we can kill factomd, then copy the leader log and a follower log:

	cp ~/.factom/m2/database/journal0.log ./leader.log
	cp ~/.factom/m2/database/journal3.log ./follower.log
	
We can then replay these messages in factomd:

	factoid -journal=leader.log -follower=false -db=Map
	factoid -journal=follower.log -follower=true -db=Map

Keep in mind, after the state has been replayed, the simulator continues to run.  So you can easily examine the resulting state, and (in the case of a leader) run more transactions and such.  And this is also journaled, so there is an ability to modify and rerun the modified states.

The journal file can also be edited.  Only messages (lines that begin with 'MsgHex:' and the following hex) are interpreted.  So you can move these lines about, or even copy and paste from other files.
	
### -net

The network is constructed using one of a number of algorithms.  tree is the default, and looks like this, where 0 is connected to 1 and 2, 1 is connected to 3 and 4, 2 is connected to 4 and 5, etc.

	                      0
	                   1     2
	                3     4     5
	             6     7     8     9
	             ...

circles is a bit hard to map out but it is a series of loops of seven nodes, where each loop is connected to about a 1/3 down from the previous loop, 2/3 down the loop prior to that one (if it exists) and 1/2 the circle 2 prior (if it exists). The goal is to maximize the number of alternate routes to later nodes to test timing of messages through the network.

long is just one long chain of nodes

loops creates a long chain of nodes that have short cuts added here and there.

Running a particular network configuration looks like:

	factomd -count=30 -net=tree
	factomd -count=30 -net=circles
	factomd -count=30 -net=long
	factomd -count=30 -net=loops
	
### -node

The simulator always keeps the focus on one node or another.  Some commands are node sensitive (printing directory blocks, etc.), and the walletapp and factom-cli talk to the node in focus.  This allows you to set the node in focus from the beginning.  Mostly a developer thing.

	factom -count=30 -node=15 -net=tree
	
Would start up with node 15 as the focus.

### -p2pAddress

This opens up a TCP port and listens for new connections.

Usage:
	-p2pAddress="tcp://:8108"

### -peers

This connects to a remote computer and passes messages and blocks between them.

	-peers="tcp://192.168.1.69:8108" 

### -prefix

This makes all the simnodes in this process followers.  It prefixes the text provided to the Node names (and the generated file names) of all the Factom instances created.   So without a prefix, you would get nodes named FNode0, FNode1, etc.  With the a_ prefix described below, you would get a_FNode0, a_FNode1, etc.

FNode0 is currently a "magic name", and the node with that name becomes the first default Leader when building Factom up from scratch.  (Of course, if you are loading an existing network, it will come up with the last set of leaders).   In any event, adding the prefix avoids having FNode0 as a name, and as a result all the nodes will be followers.

	-prefix=a_

Multi-computer example:
Computer Leader (ip x.69) `factomd -count=2 -p2pAddress="tcp://:8108" -peers="tcp://192.168.1.72:8108"`
Computer Follower (ip x.72) `factomd -count=5 -p2pAddress="tcp://:8108" -peers="tcp://192.168.1.69:8108" -follower=true -prefix=a_`



