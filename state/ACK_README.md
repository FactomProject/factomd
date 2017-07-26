# Ack

The ack API call will return information found about a given hash and ChainID. The API parameters are `hash` and `chainid`, and the request looks like so:

```json
{  
   "jsonrpc":"2.0",
   "id":0,
   "method":"ack",
   "params":{  
      "hash":"048b7a08636c3e94c0565e06cd451654834d64a73eb6023ebc5133ecd29c2313",
      "chainid":"000000000000000000000000000000000000000000000000000000000000000c"
   }
}
```

The ChainID can be either:

1. `000000000000000000000000000000000000000000000000000000000000000c` (or '`c`' for short)  
   * This signals the hash is a commit txid
2. `000000000000000000000000000000000000000000000000000000000000000f` (or '`f`' for short)
   * This signals the hash is a factoid txid
3. Anything else (given it is 64 hex characters)
   * This signals the hash is an entry hash
4. `000000000000000000000000000000000000000000000000000000000000000a` Is the only hash to be rejected, as AdminBlock is not allowed


## Status List

The various statuses are as follows

```
   "Unknown"         : Not found anywhere
   "NotConfirmed"    : Found on local node, but not in network (Holding Map)
   "TransactionACK"  : Found in network (ProcessList)
   "DBlockConfirmed" : Found in Blockchain
```

## Responses

### For Commit TxIDs (Entry Credit Chain)

Responses can vary depending on the status of the commit, which will always be found in `commitdata`. Commits may not enter the network if the entry has already been payed for within a 2hr window, because of this it is recommended to always search entries by their **entryhash**, rather than the corrosponding commit. Use the entry-credit chain, `000000000000000000000000000000000000000000000000000000000000000c` or `c` for short, for the request chainid

```json
{  
   "jsonrpc":"2.0",
   "id":0,
   "result":{  
      "committxid":"d2c0244cd51b2d7e9f62fd45995d7d7d04e478a752714cbc7e6bd602e8221214",
      "entryhash":"057b149bdaf681a193cf2857797e6c16242aa29a773bd3fd9729f6c20883780d",
      "commitdata":{  
         "status":"DBlockConfirmed"
      },
      "entrydata":{  
         "status":"DBlockConfirmed"
      }
   }
}
```

### For Entry by EntryHash (Recommended for anything entry related)

The ChainID in the Request will be the ChainID the entry is located within. Responses can vary depending on the status of the entry, but if `TransactionACK` or above, the commit txid will also be given. If `DBlockConfirmed`, the blockdate in unix time will also be given. The status of the entry can always be found under `entrydata`.

```json
{  
   "jsonrpc":"2.0",
   "id":0,
   "result":{  
      "committxid":"d2c0244cd51b2d7e9f62fd45995d7d7d04e478a752714cbc7e6bd602e8221214",
      "entryhash":"057b149bdaf681a193cf2857797e6c16242aa29a773bd3fd9729f6c20883780d",
      "commitdata":{  
         "status":"DBlockConfirmed"
      },
      "entrydata":{  
         "blockdate":1499440740,
         "blockdatestring":"2017-07-07 10:19:00",
         "status":"DBlockConfirmed"
      }
   }
}
```

### For Factoid Transaction by txid 

The ChainID in the Request will be the facotoid chain, `000000000000000000000000000000000000000000000000000000000000000f` or `f` for short. Responses are the same as the `factoid-ack` call.

```json
{  
   "jsonrpc":"2.0",
   "id":0,
   "result":{  
      "txid":"f1d9919829fa71ce18caf1bd8659cce8a06c0026d3f3fffc61054ebb25ebeaa0",
      "transactiondate":1441138021975,
      "transactiondatestring":"2015-09-01 15:07:01",
      "blockdate":1441137600000,
      "blockdatestring":"2015-09-01 15:00:00",
      "status":"DBlockConfirmed"
   }
}
```


## Technical Details


### Search Entry Commit by Commit TxID:

`state.GetEntryCommitAckByTXID(txid)`

1. Check DB --> If found, exit with `DBlockConfirmed`
2. Check the topmost PL --> If found, exit with `TransactionACK`
3. If at min 0, check the topmost PL-1 --> If found,exit with `TransactionACK`
4. Check the Holding Map --> If found, exit with `NotConfirmed`
5. Exit as `Unknown`

### Search Entry Commit by EntryHash

`state.GetEntryCommitAckByEntryHash(entryhash)`

1. Check topmost PL --> If found, set status to `TransactionACK` and exit
2. If at min 0, check topmost PL-1 --> If found, set status to `TransactionACK` and exit
3. Check state commitmap --> If found set status to `DBlockConfirmed`, but do not exit
4. Check Holding Map ---> If found, exit with `NotConfirmed`
5. Exit as `Unknown`

### Search Entry Reveal by EntryHash

`state.GetEntryRevealAckByEntryHash(entryhash)`

1. Check Database --> If found, exit with `DBlockConfirmed`
2. Check replay filter with mask --> If found, exit with `TransactionACK`
3. Check Holding Map ---> If found, exit with `NotConfirmed`
4. Exit as `Unknown`