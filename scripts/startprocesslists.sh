#!/usr/bin/env bash
reset
tail -f out.txt | gawk -f scripts/ProcessList.awk