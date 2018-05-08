# Identity Parsing

The process the code goes through to parse an identity. (Some notes by Emyrk)

## Factomd Parsing Identities

In factomd, `SyncIdentities()` handles parsing all queued entries. `AddNewIdentityEblocks()` queues up new eblocks to be parsed.


### How changes to an identity make it into admin block

In `FixupLinks()` `SyncIdentities()` is called. This function manages handling syncing all identities in the identity list. If the change is new, it has the power to also add that change to the admin block.

In `ProcessBlocks()`, the new entryblocks/entries are added to the identity list and update identity syncing appropriately. Each identity keeps a record of the entry blocks it needs to parse to be "up to date". The next FixupLinks call will parse these new blocks. This means all changes are on a 1 block delay:


Entry added at block 10 --> Admin block entry at 11 --> Effect at block 12

Example: An identity updates his efficiency at block 23. The authority shows that update at block 25, and the coinbase has the new efficiency.


### Garbage Collection

`UpdateAuthSigningKeys()` Should be called per block to remove old signing keys
`IdentityManager.CancelManager.GC()` should be called per block to remove invalid descriptor cancels.


## Identity Manager (Can be used outside factomd)

All identity entries are processes by the identity manager, and all admin block entries. It maintains the authority and identity set, and has the ability to write to the admin block.

### Processing Entries

in `identityManagerEntryBlock.go`, the function `ProcessIdentityEntryWithABlockUpdate()` is used within factomd. If processing entries outside factomd, use `ProcessIdentityEntry()`.

If the admin block is not nil, that means any identity changes can be written the admin block. If the admin block is nil, that means you are syncing an entry from not the present, and therefore cannot write to the admin block (the admin block has already been written to).