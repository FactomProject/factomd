#/bin/sh
mkdir -p test
#compile the tests
rm -rf test/*
go test -c github.com/FactomProject/factomd/engine -o test/factomd_test
#run the tests
grep -Eo " Test[^( ]+" factomd_test.go 
grep -Eo " Test[^( ]+" factomd_test.go | grep "$1" | xargs -I TestMakeALeader -n1 sh -c  'mkdir -p test/TestMakeALeader; cd test/TestMakeALeader; ../factomd_test --test.v --test.timeout 600s  --test.run TestMakeALeader  2>&1 | tee testlog.txt'
find . -name testlog.txt | xargs grep -EH "PASS:|FAIL:"
