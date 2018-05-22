# Database Generator

This tool can generate databases very quickly with various amounts of entries/blocks. A yaml config is able to tailor the generation of the database. It can build a fresh database, or build off a current as long as the CUSTOM genesis block is used (the factoid address used to fund addresses needs to have fct).

### Usage

See (gen.yaml)[gen.yaml] for config options. Options may depend on which EntryGenerator you decide to use.

#### To run for record generation

```
# b being the number of blocks to buil
DatabaseGenerator -config 15mil.yaml -b 50000
```

#### Check integrity of the DB

You can use the DatabaseIntegrityCheck to verify the database

```
DatabaseIntegrityCheck level factoid_level.db/
```


# Extending

Want to extend the tool for your own flavour of databases?

### Data Generation

If you wish to change the way data is generated, use `randomentrygen.go` and `incremententrygen.go` as a template. You must implement the various functions so the generator can retrieve data, and pay for the entries.

If you need state to make entries in past chains, it must be kept in your entrygen object (see `incremententrygen.go`)

`entrygencore.go` has the basics for all entry generation. Embedding that into your structure is the simpliest way to implement all the required functions.


### Mutli Leader

If your database has multiple leaders, all private keys will be needed. Currently only 1 `IAuthSigner` is implemented, being in `authsigning.go`. This interface will create all the dbsigs for a given dbstate. To support mutliple leaders, `IAuthSigner` will have to be implemented with multiple leader keys.

### Complex Blockchains

This tool currently only creates simple blockchains.