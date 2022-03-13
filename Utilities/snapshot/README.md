# Snapshot CLI Tool

The snapshot CLI tool is used to download all factoid balances and entry data into a easy to parse format. The tool will read from a factom database, defaulting to the main database path `$HOME/.factom/m2/main-database/ldb/MAIN/factoid_level.db`. **By default entry data is not snapshotted**. Use `-e true` to include entry data. 

All output will be written to a directory, default `./snapshot`.

## CLI

Use `snapshot new` to take a new snapshot of a given database. To delete the snapshot, you can use `snapshot clean`.

You can run against an API of a running node:
```
snapshot new --db-type=api --db=http://localhost:8088
```

```bash
Usage:
  snapshot [flags]
  snapshot [command]

Available Commands:
  clean       Deletes the snapshotted data.
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  new         Take a new snapshot of a factom database
  verify      verifies the snapshotted data against factom

Flags:
  -h, --help         help for snapshot
      --log string   set the log level (default "debug")

Use "snapshot [command] --help" for more information about a command.

```

# Data Formats

## Balances

The `balances` file will have `height <block_height>` for the height the snapshot was taken at. The `block_height`'s factoid block is included in the snapshot. All factoid balances are in factoshis.

```
height 246710
FA2FKpXhyWPGpQE9yQ7dc419n4eBW9NrPGcFXhuUQYn8BJ6wYU6H: 0
FA1zS79XhiyLRrvGGE6aADACt9Uh5oiBoyf9nYgN22HwGkYcAokP: 87975143
EC2nGxgm8LMaTY6zTE2xzvkoba61hkHy5Uxmn8Zc3udoqwuampaF: 450
EC3FzWegxBoX7Hk7ZHvSjpcpKeeN1KDPscwBf2C1ejmybSWMQJCr: 42
```

## Entries

Entries will be written to a `snapshot/entries` directory in flat files for each chain. Each chain file will have its entries written in order with a newline for each entry. Entry lines are prefixed with `et:` followed by the marshaled entry. Eblocks are prefixed with `eb:` and contain their height and keymr. Minute markers are omitted.

```
eb:4340 176d5e0e0cecfce6887268cb5753615fd10b3011f34c2654a5c8cdad9eb08a19
et:<entry_binary in base64>
```

# Verify snapshot

You can verify the snapshot against a running node. Keep in mind, if a node is on a running network, then the balances will mismatch. Ideally you check against a node that loaded a db and is not syncing any network.

```
snapshot verify balances
snapshot verify chains
```

# TODO

It could be optimized with go routines, and there might be a file limit for chains. So I might need to rotate the file cache and only have N number of file descriptors open. Eg use something like https://github.com/hashicorp/golang-lru

Currently `verify chains` does not prove **all** chains were recorded. We can do this by tallying up the chains in the system, and compare to the number of chains we recorded.