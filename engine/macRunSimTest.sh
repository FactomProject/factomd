#/bin/sh
# set -x

if [ -z "$1" ]
  then
    echo excluding long tests
    pattern='.*'
  else
    pattern="$1"
fi
if [ -z "$2" ]
  then
    echo excluding debug tests and long tests
    npattern="TestPass|TestFail|TestRandom|long"
  else
    npattern="$2|TestPass|TestFail|TestRandom|long"
fi



echo preparing to run: -$pattern- -$npattern-
grep -Eo " Test[^( ]+" factomd_test.go | grep -Ev "$npattern" | sort
sleep 3

mkdir -p test
#remove old logs
grep -Eo " Test[^( ]+" factomd_test.go | grep -Ev "$npattern" | sort |  xargs -n 1 -I testname rm -rf test/testname
#compile the tests
go test -c github.com/FactomProject/factomd/engine -o test/factomd_test
#run the tests

grep -Eo " Test[^( ]+" factomd_test.go | grep -Ev "$npattern" | sort | xargs -I TestMakeALeader -n1 bash -c  'echo "Run TestMakeALeader"; mkdir -p test/TestMakeALeader; cd test/TestMakeALeader; ../factomd_test --test.v --test.timeout 30m  --test.run "^TestMakeALeader$" &> testlog.txt; pwd; grep -EH "PASS:|FAIL:|panic|bind| Timeout "  testlog.txt'

echo "Results:"
find . -name testlog.txt | sort | xargs grep -EHm1 "PASS:"
echo ""
find . -name testlog.txt | sort | xargs grep -EHm1 "FAIL:|panic|bind| Timeout "



#(echo git checkout git rev-parse HEAD; find . -name testlog.txt | xargs grep -EH "PASS:|FAIL:|panic") | mail -s "Test results `date`" `whoami`@factom.com

