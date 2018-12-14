#/bin/sh
echo -n "Total requestes "
grep "MissingMsg " $@ | grep Send  |  wc -l
echo -n "Total unique messages requested "
grep "MissingMsg " $@ | grep Send | grep -Eo "\[[^]]* \]" | grep -Eo "[0-9]+/[0-9]/[0-9]+" | sort -u | wc -l
echo -n "Total messages requested "
grep "MissingMsg " $@ | grep Send | grep -Eo "\[[^]]* \]" | grep -Eo "[0-9]+/[0-9]/[0-9]+" |  wc -l
