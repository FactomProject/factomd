1) Make sure you have a correct factomd.conf set up in the default folder
2) If you need to connect to a different factomd instance than the main server (52.18.72.212:8088), change api.go and recompile
3) Porter will handle everything else automatically:

1) It will load up the configuration and figure out what database to use
2) It will connect to the main factomd server defined in api.go (52.18.72.212:8088)
3) It will iterate backwards over all dBlocks in a staged fashion, moving further and further every stage
4) It will fetch all blocks and entries from the main network (TODO: still need to fix the free-floating block)
5) Everything will be saved to a database with a -Import at the end

To run m2 with the new database, rename the -Import database to the appropriate database name.

----

If you need to import only a few blocks for testing, shorten the list in GetDBlockList in porter.go and don't fetch the current DBlock head.