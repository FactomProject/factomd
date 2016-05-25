#!/usr/bin/env bash
./scripts/factomload.sh > trans.out &
tail -f trans.out | gawk -f scripts/trans.awk
