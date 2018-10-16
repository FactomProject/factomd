#/bin/bash
# set -x
# exclude TestPass|TestFail|TestRandom
pattern=$1
shift
mkdir -p test

#remove old logs
grep -Eo " Test[^( ]+" factomd_test.go | grep -P "$pattern" | grep -Ev "TestPass|TestFail|TestRandom" |  xargs -n 1 -I testname rm -rf test/testname

#compile the tests
go test -c github.com/FactomProject/factomd/engine -o test/factomd_test
#run the tests

grep -Eo " Test[^( ]+" factomd_test.go | grep -P "$pattern" | grep -Ev "TestPass|TestFail|TestRandom" | sort
grep -Eo " Test[^( ]+" factomd_test.go | grep -P "$pattern" | grep -Ev "TestPass|TestFail|TestRandom" | sort | xargs -I TestMakeALeader -n1 bash -c  'mkdir -p test/TestMakeALeader; cd test/TestMakeALeader; ../factomd_test --test.v --test.timeout 600s  --test.run TestMakeALeader &> testlog.txt; pwd; grep -EH "PASS:|FAIL:|panic|bind"  testlog.txt'
find . -name testlog.txt | sort | xargs grep -EH "PASS:|FAIL:|panic|bind"



#(echo git checkout git rev-parse HEAD; find . -name testlog.txt | xargs grep -EH "PASS:|FAIL:|panic") | mail -s "Test results `date`" `whoami`@factom.com
