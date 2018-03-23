# nettool

A tool for testing a factomd network in various conditions.

## Features
 * sets up a network of factomd nodes in separate docker containers
 * sets up a seeds server to serve the list of nodes to discover
 * allows creating initial topology of the network
 * allows dynamically changing the topology by adding and removing rules

## Installation

Install the following prerequisites:

 * [docker](https://www.docker.com/community-edition)
 * [python 3.6](https://www.python.org/downloads/)

It is prefered that you create a new virtual environment for this tool.

Install all required Python libraries:

```
cd support/net
pip install -r requirements.txt
```

Verify the installation:

```
./nettool.py
```

This should display the current usage screen for the tool:

```
Usage:
    ./nettool.py (-h | --help)
    ./nettool.py status [-f FILE | --file FILE]
    ./nettool.py up [--build] [-f FILE | --file FILE]
    ./nettool.py down [--destroy] [-f FILE | --file FILE]
    ./nettool.py ins <from> <to> <action> [--one-way] [-f FILE | --file FILE]
    ./nettool.py add <from> <to> <action> [--one-way] [-f FILE | --file FILE]
    ./nettool.py del <from> <to> <action> [--one-way] [-f FILE | --file FILE]

```

Use `./nettool.py -h` to display the help for the usage.

## Configuration

The tool uses a YAML configuration file that describes the details of the
network that will be set up. By default it looks for the `config.yml` file in
the current directory, but other files can be passed using the `-f | --file`
argument.

> Before changing the file, make sure the environment is down, otherwise the
> tool may lose track of the created containers or rules and you'll need to
> clean it up manually.

See [config.yml](config.yml) for an annotated example on how to create and
modify the network configuration. There are also other examples available in
the [examples](examples) directory.


## Details

The tool uses the following components:
 * `factomd` instances built using the code currently checked out in the parent
   directory (this can be overriden in the config file)
 * `seeds_server` - an nginx instance that serves a list of initial nodes, so
   that the nodes can discover each other in the network
 * `gateway` - a helper container started with elevated privileges to allow
   easy manipulation of network rules - this is necessary for the tool to work
   in a Docker for Mac / Docker for Windows environment, where the user does
   not have direct access to the VM hosting the docker engine
 * a custom Docker network to allow connectivity between `factomd` nodes and
   the seeds server.

Based on the provided configuration file, the tool:
 * builds the images for `factomd`, `seeds_server` and `gateway` if necessary
 * assigns static IPs to the `seeds_server` and all the `factomd` instances
 * creates the Docker network and applies the network rules to create the
   initial network topology
 * starts the `factomd` network in the created environment.

### Network rules

The tool directly manipulates `iptables` rules on the machine that hosts the
docker engine (host machine in case of Linux, a VM in case of Mac/Windows), so
the rules defined in the configuration file or added dynamically work in the
similar way to the `iptables` rules. Each rule has a:
  * `source` and `target` - these can be one of:
    * the `factomd` node name defined in the configuration file
    * `*` to denote the whole network
    * any IP or network definition that is valid for `iptables` rules
  * `action`:
    * `allow` - allows network packets to get from the `source` to the `target`
    * `deny` - prevents network packets from getting from the `source` to the
    `target`
  * by default rules are defined both ways (e.g. `deny` from `A` to `B` will
  prevent packets from getting from `A` to `B` and from `B` to `A`), but this
  behavior can be overriden to test assymetric connectivity failures.

Rules are matched in the order of definition, whenever a packet matches the
predicate, processing is stopped and the given `allow` / `deny` action is
applied.

Example:

```
Rules:
  deny: A -> *
  allow: A -> B
```

The allow rule will never have any effect, since the first rule always matches.


```
Rules:
  allow: A -> B
  deny: A -> *
```

This setup will allow `A` to contact `B`, but deny any other connections
originating from `A`.

## Basic usage

### Help and current status

* Display help and all the available options:

```
./nettool.py -h
```
```
./nettool.py --help
```

* Display the status of the current environment (described in `config.yml`):

```
./nettool.py status
```

### Changing the configuration file

* All commands accept the `-f` or `--file` argument that allows specyfing the
  configuration file (if other than the default `config.yml`):

```
./nettool.py -f examples/tree.yml status
```

### Starting / stopping the network

* Start the network:

```
./nettool.py up
```

The command above will check if the images for all containers are available and
build them if necessary. Note that the image is built only once so changes in
the factomd code will not be reflected.

The command attempts to bring back the containers from any state, so e.g. after
the container was stopped or killed, you can use the `up` command to bring
everything back to the initial running state.

* Start the network, but rebuild all the images first (useful e.g. when some
  changes were made to the factomd code and you want to test it):

```
./nettool.py up --build
```

* Stop the environment:

```
./nettool.py down
```

* Stop the environment and remove all built images and other artifacts:

```
./nettool.py down --destroy
```

### Dynamic rules manipulation

* Display the current set of rules:

```
$ ./nettool.py up
...
$ ./nettool.py status
...
Rules:
    deny: node2 -> *
    deny: * -> node2
    deny: node3 -> *
    deny: * -> node3

```

* Insert a rule at the beginning:

```
$ ./nettool.py status
...
Rules:
    deny: node2 -> *
    deny: * -> node2
    deny: node3 -> *
    deny: * -> node3

$ ./nettool.py ins node1 node2 allow
...

$ ./nettool.py status
...
Rules:
    allow: node2 -> node1
    allow: node1 -> node2
    deny: node2 -> *
    deny: * -> node2
    deny: node3 -> *
    deny: * -> node3
```

* Append a rule at the end:

```
$ ./nettool.py status
...
Rules:
    deny: node2 -> *
    deny: * -> node2
    deny: node3 -> *
    deny: * -> node3

$ ./nettool.py ins node1 node2 allow
...

$ ./nettool.py status
...
Rules:
    deny: node2 -> *
    deny: * -> node2
    deny: node3 -> *
    deny: * -> node3
    allow: node2 -> node1
    allow: node1 -> node2
```

* Delete an existing rule:

```
$ ./nettool.py status
...
Rules:
    deny: node2 -> *
    deny: * -> node2
    deny: node3 -> *
    deny: * -> node3
    allow: node2 -> node1
    allow: node1 -> node2

$ ./nettool.py del node1 node2 allow
...

$ ./nettool.py status
...
Rules:
    deny: node2 -> *
    deny: * -> node2
    deny: node3 -> *
    deny: * -> node3

```

* Define a one-way rule (instead of the default two-way):


```
$ ./nettool.py status
...
Rules:
    deny: node2 -> *
    deny: * -> node2
    deny: node3 -> *
    deny: * -> node3

$ ./nettool.py ins node1 node2 allow --one-way
...

$ ./nettool.py status
...
Rules:
    allow: node1 -> node2
    deny: node2 -> *
    deny: * -> node2
    deny: node3 -> *
    deny: * -> node3
```

