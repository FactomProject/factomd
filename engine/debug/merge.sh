#/bin/sh
#grep -E \"\" $@ | awk -f msgOrder.awk | sort -n | less -R
echo $@
grep -E . "$@"  | awk -f msgOrder.awk | sort -n | less -R

