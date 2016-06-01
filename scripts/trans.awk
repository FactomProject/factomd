/Factoid Addresses/ { print "\nBalances" }
/Transactions/ { print $0 "\n"}
/input:/ {print}
/[^c]output:/ {print}
/ecoutput:/ {print}
/Creating Chain:/ {print}
/Command Failed:  Server Error/ { print "\n******* No Factom System Available ******" }
/    b.  / {print}
/    e.  / {print}


