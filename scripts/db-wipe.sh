#!/usr/bin/env bash

echo Nuking the current database, so we can start over.
echo routines: db-wipe.sh db-copy.sh db-restore.sh
rm -r ~/.factom/m2/local-database

