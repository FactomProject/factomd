#!/usr/bin/env bash
tail -f trans.out | gawk -f scripts/trans.awk
