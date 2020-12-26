# Simulator 

factomd can run a simulated network via the commandline.  This allows testing of much more complicated networks than would be possible otherwise.   The  simulator is very extensible, so new features will be added as we go along.

## Databases and Journaling

When you run factomd, if no database directory exists in the m2 directory, one is created.  All databases are created under the ~/.factom/m2/database directory.  These databases are named "bolt" for the main factom node, and "blotSimNNN" where NNN is the node number for Simulation nodes cloned from the main factom node.  Inside each of these directories, a network folder is created.  LOCAL (for test networks built on your own machine), MAIN for the main network, and TEST for the test network.   This allows swapping between networks without concern for corupting the databases.

As M2 runs, journal files are created in the database directory. All messages are journaled for all nodes in the simulator.  This gives the ability to "rerun" a message sequence to debug observed issues. When factomd is restarted, all journal files for those nodes are reset.

Below is a discription of how to run journal files.

## Flags to control the simulator

To get the current list of flags, type the command:

	factomd -h

Which will get you something like:
	
    Command Line Arguments:
        -h
    Usage of factomd:
    -balancehash
        If false, then don't pass around balance hashes (default true)
    -blktime int
        Seconds per block.  Production is 600.
    -broadcastnum int
        Number of peers to broadcast to in the peer to peer networking (default 16)
    -checkheads
        Enables checking chain heads on boot (default true)
    -clonedb string
        Override the main node and use this database for the clones in a Network.
    -config string
        Override the config file location (factomd.conf)
    -controlpanelport int
        Port for control panel webserver;  Default 8090
    -controlpanelsetting string
        Can set to 'disabled', 'readonly', or 'readwrite' to overwrite config file
    -count int
        The number of nodes to generate (default 1)
    -customnet string
        This string specifies a custom blockchain network ID.
    -db string
        Override the Database in the Config file and use this Database implementation. Options Map, LDB, or Bolt
    -deadline int
        Timeout Delay in milliseconds used on Reads and Writes to the network comm (default 1000)
    -debugconsole string
        Enable DebugConsole on port. localhost:8093 open 8093 and spawns a telnet console, remotehost:8093 open 8093
    -debuglog string
        regex to pick which logs to save
    -drop int
        Number of messages to drop out of every thousand
    -enablenet
        Enable or disable networking (default true)
    -exclusive
        If true, we only dial out to special/trusted peers.
    -exclusive_in
        If true, we only dial out to special/trusted peers and no incoming connections are accepted.
    -exposeprofiler
        Setting this exposes the profiling port to outside localhost.
    -factomhome string
        Set the Factom home directory. The .factom folder will be placed here if set, otherwise it will default to $HOME
    -fast
        If true, Factomd will fast-boot from a file. (default true)
    -fastlocation string
        Directory to put the Fast-boot file in.
    -fastsaverate int
        Save a fastboot file every so many blocks. Should be > 1000 for live systems. (default 1000)
    -faulttimeout int
        Seconds before considering Federated servers at-fault. Default is 120. (default 120)
    -fixheads
        If --checkheads is enabled, then this will also correct any errors reported (default true)
    -fnet string
        Read the given file to build the network connections
    -follower
        If true, force node to be a follower.  Only used when replaying a journal.
    -fullhasheslog
        true create a log of all unique hashes seen during processing
    -journal string
        Rerun a Journal of messages
    -journaling
        Write a journal of all messages received. Default is off.
    -keepmismatch
        If true, do not discard DBStates even when a majority of DBSignatures have a different hash
    -leader
        If true, force node to be a leader.  Only used when replaying a journal. (default true)
    -logPort string
        Port for pprof logging (default "6060")
    -logjson
        Use to set logging to use a json formatting
    -loglvl string
        Set log level to either: none, debug, info, warning, error, fatal or panic (default "none")
    -logstash
        If true, use Logstash
    -logurl string
        Endpoint URL for Logstash (default "localhost:8345")
    -mpr int
        Set the Memory Profile Rate to update profiling per X bytes allocated. Default 512K, set to 1 to profile everything, 0 to disable. (default 524288)
    -net string
        The default algorithm to build the network connections (default "tree")
    -network string
        Network to join: MAIN, TEST or LOCAL
    -networkport int
        Port for p2p network; default 8110
    -node int
        Node Number the simulator will set as the focus
    -nodename string
        Assign a name to the node
    -peers string
        Array of peer addresses. 
    -plugin string
        Input the path to any plugin binaries
    -port int
        Port where we serve WSAPI;  default 8088
    -prefix string
        Prefix the Factom Node Names with this value; used to create leaderless networks.
    -reparseanchorchains
        If true, reparse bitcoin and ethereum anchor chains in the database
    -rotate
        If true, responsibility is owned by one leader, and Rotated over the leaders.
    -roundtimeout int
        Seconds before audit servers will increment rounds and volunteer. (default 30)
    -rpcpass string
        Password to protect factomd local API. Ignored if rpcuser is blank
    -rpcuser string
        Username to protect factomd local API with simple HTTP authentication
    -runtimeLog
        If true, maintain runtime logs of messages passed.
    -selfaddr string
        comma separated IPAddresses and DNS names of this factomd to use when creating a cert file
    -sim_stdin
        If true, sim control reads from stdin. (default true)
    -startdelay int
        Delay to start processing messages, in seconds (default 10)
    -stderrlog string
        Log stderr to a file, optionally the same file as stdout
    -stdoutlog string
        Log stdout to a file
    -sync2 int
        Set the initial blockheight for the second Sync pass. Used to force a total sync, or skip unnecessary syncing of entries. (default -1)
    -test.bench regexp
        run only benchmarks matching regexp
    -test.benchmem
        print memory allocations for benchmarks
    -test.benchtime d
        run each benchmark for duration d (default 1s)
    -test.blockprofile file
        write a goroutine blocking profile to file
    -test.blockprofilerate rate
        set blocking profile rate (see runtime.SetBlockProfileRate) (default 1)
    -test.count n
        run tests and benchmarks n times (default 1)
    -test.coverprofile file
        write a coverage profile to file
    -test.cpu list
        comma-separated list of cpu counts to run each test with
    -test.cpuprofile file
        write a cpu profile to file
    -test.failfast
        do not start new tests after the first test failure
    -test.list regexp
        list tests, examples, and benchmarks matching regexp then exit
    -test.memprofile file
        write an allocation profile to file
    -test.memprofilerate rate
        set memory allocation profiling rate (see runtime.MemProfileRate)
    -test.mutexprofile string
        write a mutex contention profile to the named file after execution
    -test.mutexprofilefraction int
        if >= 0, calls runtime.SetMutexProfileFraction() (default 1)
    -test.outputdir dir
        write profiles to dir
    -test.parallel n
        run at most n tests in parallel (default 8)
    -test.run regexp
        run only tests and examples matching regexp
    -test.short
        run smaller test suite to save time
    -test.testlogfile file
        write test action log to file (for use only by cmd/go)
    -test.timeout d
        panic test binary after duration d (default 0, timeout disabled)
    -test.trace file
        write an execution trace to file
    -test.v
        verbose: print additional output
    -timedelta int
        Maximum timeDelta in milliseconds to offset each node.  Simulates deltas in system clocks over a network.
    -tls
        Set to true to require encrypted connections to factomd API and Control Panel
    -tormanage
        Use torrent dbstate manager. Must have plugin binary installed and in $PATH
    -torupload
        Be a torrent uploader
    -waitentries
        Wait for Entries to be validated prior to execution of messages
    -wrproc
        Write processed blocks to temporary debug file (default true)


The flags that begin with "test." are supplied by the profiling package installed.  The flags that relate to running factomd and the simulator are the following, with a little more explaination.  That follows below.

## Simulator Commands

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

## Simulator Commands Continued -- Identity
M2 requires servers to have identities if they wish to have the ability to become a federated or audit server. To create an identity, entries must be entered into the blockchain, so controls were added to the simulator to assist in the creation and attachment of identities. Identites take about a minute to generate on a (macbook pro laptop) to meet the proper requirements, so a stack of identities are pregenerated to make testing easier.

How the simulator controls work. First every instance of factomd will share the same stack of identites. Each instance will also have a local pool of identities they can use and attach to their nodes. To load their local pool of identities, they can pop identities off the shared stack, then attach the next open identity in their local identity pool to the current node:

* gN -- Moves N identities from the shared stack to local identity pool
  * Be mindful everyone shares the stack and it can run out.
* t -- Attaches the next identity in the local pool that has not been taken to the current node

<i>The 'gN' command will load entry credits into the zeros entry credit wallet to fund all identity sim controls if the wallet is low on funds. </i>

## Launching Factomd
 
Personally I open two consoles.  I run factomd redirected to out.txt, and in another console I run tail -f out.txt.

So in one console I run a command like:

	factomd -count=10 -net=tree > out.txt
	
And in another console I run:

	tail -f out.txt
	
Then I type commands in the first console as described above, and see the output in the second.  Also messages and errors will show up in the first console (leaving the second console with simple output from factomd).

## -count

The command:
	
	factomd -count=10

Will run the simulator with the configuration found in ~/.factom/m2/factom.conf with 10 nodes, named fnode0 to fnode9.  fnode0 will be the leader.  When you hit enter to get the status, you will see an "L" next to leader nodes.

## -db

You can override the database implementation used to run the simulator.  The options are "Bolt", "LDB", and "Map".  "Bolt" will give you a bolt database, "LDB" will get you a LevelDB database, and "Map" will get you a hashtable based database (i.e. nothing stored to disk).  The most common use of -db is to specify a "Map" database for tests that can be easily rerun without concern for past runs of the tests, or of messing up the database state.

Keep in mind that Map will still overwrite any journals.  For example, you can run a 10 node Factom network in memory with the following command:

	factoid -count=10 -db=Map
	
## -follower

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
	
## -net

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
	
## -node

The simulator always keeps the focus on one node or another.  Some commands are node sensitive (printing directory blocks, etc.), and the walletapp and factom-cli talk to the node in focus.  This allows you to set the node in focus from the beginning.  Mostly a developer thing.

	factom -count=30 -node=15 -net=tree
	
Would start up with node 15 as the focus.

## -p2pAddress

This opens up a TCP port and listens for new connections.

Usage:
	-p2pAddress="tcp://:8108"

## -peers

This connects to a remote computer and passes messages and blocks between them.

	-peers="tcp://192.168.1.69:8108" 

## -prefix

This makes all the simnodes in this process followers.  It prefixes the text provided to the Node names (and the generated file names) of all the Factom instances created.   So without a prefix, you would get nodes named FNode0, FNode1, etc.  With the a_ prefix described below, you would get a_FNode0, a_FNode1, etc.

FNode0 is currently a "magic name", and the node with that name becomes the first default Leader when building Factom up from scratch.  (Of course, if you are loading an existing network, it will come up with the last set of leaders).   In any event, adding the prefix avoids having FNode0 as a name, and as a result all the nodes will be followers.

	-prefix=a_

Multi-computer example:
Computer Leader (ip x.69) `factomd -count=2 -p2pAddress="tcp://:8108" -peers="tcp://192.168.1.72:8108"`
Computer Follower (ip x.72) `factomd -count=5 -p2pAddress="tcp://:8108" -peers="tcp://192.168.1.69:8108" -follower=true -prefix=a_`



