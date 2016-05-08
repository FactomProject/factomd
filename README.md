# Factom

Factom is an Open-Source project that provides a way to build applications on the Bitcoin blockchain. 

Factom began by providing proof of existence services, but then move on to provide proof of existence of transforms.  A list of such entries can be thought of as a Factom Chain.  Factom can be used to implement private tokens, smart contracts, smart properties, and more.

Factom leverages the Bitcoin Blockchain, but in a way that minimizes the amount of data actually inserted in the Blockchain.  Thus it provides a mechanism for creating Bitcoin 2.0 services for the trading of assets, securities, commodities, or other complex applications without increasing blockchain "pollution".

Factom is designed to be a distributed, autonomous protocol, like Bitcoin.  Factom has a token that is used to provide a way for users to pay for the use of the protocol, and give the Federated Servers (required to run the protocol) incentives to run and promote the protocol.  

## State of Development

We are very much at an Alpha level of development.  This is the M2 codebase, which is a direct implementation of the Factom Whitepaper and the Factom Consensus Protocol.

## Getting Started

You need to set up Go environment with golang 1.5 or 1.6. You also need to install the latest version of git, and it doesn't hurt to set up a github account.

###Install the m2 repository

Get the M2 database, with the following command

	go get github.com/FactomProject/factomd

You should now be ready to execute factomd.  The following sections go into detail about how to compile and run factomd.

###Testing M2

The test team is working on the master branch, while the developers are working on the m2 branch.  But because of shared repositories (some of which have some m2 changes), moving between milestone 1 code, m2s, and m2 is a bit complicated.  The all.sh script is your friend.  Follow these steps to get M2 setup and running for test:

	cd ~/go/src/github.com/FactomProject/factomd	# <However or wherever you put it>
	./all.sh m2					# This is going to put you into the m2 branch
	git checkout master				# Gets you back to master
	go install					# Recompiles factomd with master code
	
You are now compiled.  We still need to setup the configuration for running factomd.
	
### Running the M2 Simulator for the first time

Create a ~/.factom/m2 directory

cd to the factomd directory created with the go get.

On Linux and OS X execute:

	cp factomd.conf ~/.factom/m2/

Now you are ready to execute factomd.

	go install
	factomd

This is the simplist way to execute factomd with the defaults.  You can hit "enter" to get the status of the factom nodes running in the simulator.  Every so often, the simulator will provide you an update on the status of the nodes you are running.  More about running the Simulator follows...

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
	
	//////////////////////// Copyright 2015 Factom Foundation
	//////////////////////// Use of this source code is governed by the MIT
	//////////////////////// license that can be found in the LICENSE file.
	Go compiler version: go1.6.2
	Using build: 
	len(Args) 2
	Usage of factomd:
	  -blktime int
	    	Seconds per block.  Production is 600.
	  -count int
	    	The number of nodes to generate (default 1)
	  -db string
	    	Override the Database in the Config file and use this Database implementation
	  -drop int
	    	Number of messages to drop out of every thousand
	  -folder string
	    	Directory in .factom to store nodes. (eg: multiple nodes on one filesystem support)
	  -follower
	    	If true, force node to be a follower.  Only used when replaying a journal.
	  -heartbeat
	    	If true, network just sends heartbeats.
	  -journal string
	    	Rerun a Journal of messages
	  -leader
	    	If true, force node to be a leader.  Only used when replaying a journal. (default true)
	  -net string
	    	The default algorithm to build the network connections (default "tree")
	  -netdebug
	    	If true, print detailed network debugging info.
	  -node int
 	   	Node Number the simulator will set as the focus
	  -p2pAddress string
 	   	Address & port to listen for peers on: (eg: tcp://127.0.0.1:40891) (default "tcp://127.0.0.1:34340")
	  -peers string
	    	Array of peer addresses. Defaults to: "tcp://127.0.0.1:34341 tcp://127.0.0.1:34342 tcp://127.0.0.1:34340"
	  -port int
	    	Address to serve WSAPI on
	  -prefix string
	    	Prefix the Factom Node Names with this value; used to create leaderless networks.
	  -runtimeLog
	    	If true, maintain runtime logs of messages passed. (default true)
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
	    	maximum test parallelism (default 8)
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


The flags that begin with "test." are supplied by the profiling package installed.  The flags that relate to running factomd and the simulator are the following, with a little more explaination.  That follows below.

### Simulator Commands

While the simulator is running, you can perform a number of commands to poke at, and examine the state of factomd.  The interface to the Simulator is very unsophisticated.  You type one of the following commands, followed by <enter> 

* aN -- Dump the Admin block at directory block height N. So like a1 or a213. Currently just dumps the JSON. 
* fN -- Dump the Factoid block at directory block height N.  So like f5 or f1203.  Pretty prints.
* dN -- Dump the Directory block at directory block height N.  d4 or d21230
* s  -- gives the state of all nodes in the simulated network. This state will be updated as the simulator runs.  You can turn off the output by entering s again.
* m -- Dumps all the messages in the system to standard out.  This output is updated as the simulation runs.  You can turn this off by hitting 'm' again.
* p -- Dumps Process Lists and Directory Block States every so often.
* N -- Where N is a number from 0 to N, where N is a number one less than the number of nodes in your simulator.  So for a simulation of 5 nodes, you could enter a number between 0-5.  Typing a node number then hitting <enter> shifts the focus of the simulator to the node number specified.  You now are talking to said node from the CLI or wallet
* ' ' -- A space simply moves the focus to the numerically next node.  So starting with 0, a space followed by <enter> moves the focus to Node 1.  If you are on the last node, this wraps back to node 0.
* l -- We now support multiple leaders.  Shifting the focus of the simulator to a node and typing 'l' <enter> will make that node a leader.     
* <enter> -- provides help on these commands.

 
Personally I open two consoles.  I run factomd redirected to out.txt, and in another console I run tail -f out.txt.

So in one console I run a command like:

	factomd -count=10 -net=tree > out.txt
	
And in another console I run:

	tail -f out.txt
	
Then I type commands in the first console as described above, and see the output in the second.  Also messages and errors and feed backk on state changes to the simulator will show up in the first console (leaving the second console with simple output from factomd).

Running the simulator and doing testing requires that we can change the parameters of the simulator.  While we could constantly edit the configuration file, this becomes tedious very quickly.  So we have added a number of flags that allow the user to change quite a number of the configuation settings without modifying the default configuration.

### -count

The command:
	
	factomd -count=10

Will run the simulator with the configuration found in ~/.factom/m2/factom.conf with 10 nodes, named fnode0 to fnode9.  fnode0 will be the leader.  When you hit enter to get the status, you will see an "L" next to leader nodes.

### -db

You can override the database implementation used to run the simulator.  The options are "Bolt", "LDB", and "Map".  "Bolt" will give you a bolt database, "LDB" will get you a LevelDB database, and "Map" will get you a hashtable based database (i.e. nothing stored to disk").  The most common use of -db is to specify a "Map" database for tests that can be easily rerun without concern for past runs of the tests, or of messing up the database state.

Keep in mind that Map will still overwrite any journals.  For example, you can run a 10 node Factom network in memory with the following command:

	factoid -count=10 -db=Map

### -prefix

Adding a prefix to all the node names in the simulator allows running multiple simulations at the same time on one machine.  We default the name FNode0 to be the leader, and adding a prefix means no node in the simulation will be a leader.
	
### -follower

At times it is nice to force factomd to launch a follower rather than a leader (or the other way around).  This only works well when playing back a journal of messages to investigate why a server got into a particular state.  So suppose we have a leader journal leader.log.  We could execute that log with this command:

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

long is just one long chain of nodes.  This is the minimum number of connections a network can have and still be contiguous.

loops creates a long chain of nodes that have short cuts added here and there.

alot creates a network with many connections.

Running a particular network configuration looks like:

	factomd -count=30 -net=tree
	factomd -count=30 -net=circles
	factomd -count=30 -net=long
	factomd -count=30 -net=loops
	factomd -count=30 -net=alot
	
	
### -node

The simulator always keeps the focus on one node or another.  Some commands are node sensitive (printing directory blocks, etc.), and the walletapp and factom-cli talk to the node in focus.  This allows you to set the node in focus from the beginning.  Mostly a developer thing.

	factom -count=30 -node=15 -net=tree
	
Would start up with node 15 as the focus.

### -drop

The simulator can simulate networks with poor connections by dropping messages.  The drop number is the number of messages out of 1000 that will be dropped.  So the command:

	factom -count=30 -drop=11
	
Will drop 1.1% of all messages.

### -p2pAddress

This opens up a TCP port and listens for new connections.

Usage:
	-p2pAddress="tcp://:8108"

### -peers

This connects to a remote computer and passes messages and blocks between them.

	-peers="tcp://192.168.1.69:8108" 

Multi-computer example:
Computer Leader (ip x.69) `factomd -count=2 -p2pAddress="tcp://:8108" -peers="tcp://192.168.1.72:8108"`
Computer Follower (ip x.72) `factomd -count=5 -p2pAddress="tcp://:8108" -peers="tcp://192.168.1.69:8108" -follower=true

### -prefix

This makes all the simnodes in this process followers.  It prefixes the text to the output  

Leaders cannot run with the prefix.

	-prefix=a_

would make all the nodes named a_FNode0, a_FNode1, etc.

### -blktime

Sets the time for block generation.  In production we run 10 minute blocks, and since blktime takes it parameter in seconds, production would be -blktime=600.  For testing we tend to use shorter block times because waiting for blocks is tedious and unproductive.  The default for the simulator is a 6 second block time.




