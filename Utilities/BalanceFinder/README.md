# BalanceFinder

Given a database, this tool will compute all the balances for all addresses and write them to a file
```
BalanceFinder -o out level ~/.factom/m2/main-database/ldb/MAIN/factoid_level.db
```

## Balance hashes

To compute balance hashes, you can add heights for them to be computed at:

E.G: Print balance hashes at dbheight 1000 and 2000
```
BalanceFinder -h 1000 -h 2000 -o out level ~/.factom/m2/main-database/ldb/MAIN/factoid_level.db
```
