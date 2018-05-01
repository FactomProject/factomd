#/bin/sh
pattern="$1"
shift
echo "grep -E \"$pattern\" $@ | awk -f msgOrder.awk | sort -n | grep -E \"$pattern\" --color='always' | less -R"
grep -H -E "$pattern" "$@" | awk -f msgOrder.awk | sort -n | grep -E "$pattern" --color='always' | less -R
