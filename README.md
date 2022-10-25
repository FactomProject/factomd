# Factom

**The Factom Era of the protocol ended on 30 Oct 2022. The the protocol continues under Accumulate Era as of 31 Oct 2022**

**The project repositories for the Protocol have moved to https://GitLab.com/AccumulateNetwork**


Official Golang implementation of the Factom Protocol.

[![CircleCI](https://circleci.com/gh/FactomProject/factomd/tree/master.svg?style=shield)](https://circleci.com/gh/FactomProject/factomd/tree/master)

Factom is an open-source blockchain project designed to store structured data immutably at a fixed cost. Data integrity verification is one of the many uses cases the protocol excels at. The Factom Protocol represents opportunities for businesses, governments, and other entities to record and preserve their data in an immutable and cost-efficient way. 
[Visit the Factom Protocol website for more information](https://www.factomprotocol.org/)

## Community

Connect with the community through one or more of the following:

[Website](https://www.factomprotocol.org/) ◾ [Discord](https://discordapp.com/invite/YYM9w2V) ◾ [Twitter](https://twitter.com/factomprotocol) ◾ [Reddit](https://www.reddit.com/r/factom/) ◾ [Youtube](https://www.youtube.com/channel/UCmxp39JZjPaHHRObW3R3Stg)

## Getting Started

### Running a node with docker

Factomd releases include docker images that can be pulled from the official repository: https://hub.docker.com/r/factominc/factomd The alpine version is strongly recommended.

For instance to run the latest master build for mainnet:
```bash
docker run -d --name "factomd" -p "8088:8088" -p "8090:8090" -p "8108:8108" factominc/factomd:master-alpine
```

### Build/install factomd

You will need the following software installed on your machine:
- `git`
- A recent version of Go (supporting Go modules)
- `make`

```bash
git clone https://github.com/FactomProject/factomd
cd factomd
make
# or `make install` to install factomd into GOBIN folder
```

Also `make all` cross compiles factomd for various platforms.

### Running factomd

Factomd can be run with default configuration with this simple command:

```
$ factomd
```

The default factomd folder is located at `~/.factom/`. The node database is stored by default at `~/.factom/m2/<network type>-database`.

#### Configuration

Factomd will run with a set of default configuration. Those can be altered in two ways:
- By passing flags when running factomd. Run `factomd --help` to get the list of supported parameters.
- By editing a config file. See `factomd.conf` for an example of such file. Note that this config file is also used to configure `factom-walletd`.

```bash
mkdir -p ~/.factom/m2/
cp factomd.conf ~/.factom/m2/
```

#### Default ports

| Service       | Port         |
| ------------- | ------------ |
| JSON-RPC API  | 8088         |
| Web UI        | 8090         |
| P2P           | 8108 (MAIN net), 8109 (TEST net), 8110  (LOCAL net)|

### Running factomd for local development

To get a local development node running:

```
$ factomd --network=LOCAL --blktime=60
```

Note that the blocktime here is set to 60s (instead of the regulat 600s) for the convenience of development. Head to `http://localhost:8090` to see the Web UI of your local node.

On a LOCAL network, the address `FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q` comes pre-loaded with Factoids that can be used for your testing as the associated private key is known: `Fs3E9gV6DXsYzf7Fqx1fVBQPQXV695eP3k5XbmHEZVRLkMdD9qCK`.
