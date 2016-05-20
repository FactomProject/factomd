#!/usr/bin/env bash
./scripts/lightload.sh > trans.out &
tail -f trans.out | gawk -f scripts/trans.awk
