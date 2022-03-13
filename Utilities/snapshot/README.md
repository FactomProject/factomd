# Snapshot CLI Tool

The snapshot CLI tool is used to download all factoid balances and entry data into a easy to parse format. The tool will read from a factom database, defaulting to the main database path `$HOME/.factom/m2/main-database/ldb/MAIN/factoid_level.db`.

All output will be written to a directory, default `./snapshot`.

## CLI Options

```bash
Usage:
  snapshot [flags]

Flags:
      --db string                   the location of the database to use (default "$HOME/.factom/m2/main-database/ldb/MAIN/factoid_level.db")
      --db-type string              optionally change the type to 'bolt' (default "level")
      --debug string                set the log level (default "debug")
      --debug-heights uint32Array   heights to print diagnostic information at (default &cmd.uint32ArrayFlags(nil))
  -d, --dump-dir string             where to dump snapshot data. empty means do not dump (default "snapshot")
  -h, --help                        help for snapshot
  -s, --stop-height int             height to stop the snapshot at (default -1)
```

# Data Formats

## Balances

The `balances` file will have `height <block_height>` for the height the height the snapshot was taken at. The `block_height`'s factoid block is included in the snapshot. All factoid balances are in factoshis.

```
height 246710
FA2FKpXhyWPGpQE9yQ7dc419n4eBW9NrPGcFXhuUQYn8BJ6wYU6H: 0
FA1zS79XhiyLRrvGGE6aADACt9Uh5oiBoyf9nYgN22HwGkYcAokP: 87975143
EC2nGxgm8LMaTY6zTE2xzvkoba61hkHy5Uxmn8Zc3udoqwuampaF: 450
EC3FzWegxBoX7Hk7ZHvSjpcpKeeN1KDPscwBf2C1ejmybSWMQJCr: 42
```