#!/usr/bin/env bash
#Count the MMR traffic ina aset of log files
#countMMR.sh fnode0_network*
echo -n "Total requests "
grep "MissingMsg " $@ | grep Send  |  wc -l
echo -n "Total unique messages requested "
grep "MissingMsg " $@ | grep Send | grep -Eo "\[[^]]* \]" | grep -Eo "[0-9]+/[0-9]/[0-9]+" | sort -u | wc -l
echo -n "Total messages requested "
grep "MissingMsg " $@ | grep Send | grep -Eo "\[[^]]* \]" | grep -Eo "[0-9]+/[0-9]/[0-9]+" |  wc -l
