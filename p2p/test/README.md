
# P2P/test

This directory contains scripts designed to help manage running tests of factomd
across a cluster of cloud servers.  

It is organized into three classes of scripts and some support files

#### Bringing up a new cluster, EZ mode
1. Add the machines to followers.conf, and the leader to leader.conf
2. Run all_config.sh to do initial configuration
3. Run all_update.sh to install factomd and the scripts
4. Run test_start.sh to start the tests. (Note: as of this writing this
script doesn't work quite right and you have to hit control-c (at least on
the mac) after it starts each machine.


### Config Files

Leader.conf and Follower.conf are both lists of machines in the standard format.  
Each machine has its own line.  Generally there should be one leader and multiple
followers.

#### Standard Format for describing machines. 

Wherever one needs to indicate a machine, whether as a parameter to a script or in
a configuration file, the format used is "user@server.address.or.IP".  For example:
"example@factom.com".   (Locally I have aliases setup in my .ssh config so you will
see things like "m2p2pa" which get resolved to the "user@machine" format.

### EZ Mode Scripts

(run these locally)

all_config.sh - resets *all* of the servers for testing.
     ** Note** this resets the .factom and deletes your database!

all_update.sh - builds and pushes the latest version of factomd, and these scripts
 to the nodes. 

test_start.sh - Starts testing on all the nodes (control-c between each node to
 continue)

test_stop.sh - Stops the tests on all nodes.

(on remote nodes)

r - this is a script that just tail -f on runlog.txt

### Utility Scripts

config_remote_test_box.sh takes as a parameter a remote machine.  This is in the
"standard format" descirbed below.   This is for setting up a single machine that
has just been deployed.  It copies over auth_keys (you need to have the FactomInc
repo in your gopath) and deletes the .factom folder and recreates it, uloading
factomd.conf into the right places.

copy_files_to_test_box.sh - stops factomd, copyies the local factom d (built by
build_Factomd.sh) to the box, copies the leader, follower stop and run_new_header
scripts.

build_factomd.sh - builds a local copy of factom (cross compiles to linux) to
deploy to the nodes.



