#!/usr/bin/env bash

echo Restoring a database that was saved previously with cpdb.sh
echo routines: db-wipe.sh db-copy.sh db-restore.sh
rm -r ~/.factom/m2/local-database
cp -r ~/.factom/m2/hld_db ~/.factom/m2/local-database
