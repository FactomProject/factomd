#!/usr/bin/env bash

echo "Copying the current database for later restore using restoredb.sh"
echo routines: db-wipe.sh db-copy.sh db-restore.sh
rm -r ~/.factom/m2/hld_db
cp -r ~/.factom/m2/local-database ~/.factom/m2/hld_db
