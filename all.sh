# Copyright 2015 Factom Foundation
# Use of this source code is governed by the MIT
# license that can be found in the LICENSE file.

# To be run from within the FactomCode project.
#
# Factom has a pile of dependencies, and development requries that these be
# kept in sync with each other.  This script allows you to check out a
# particular branch in many repositories, while specifying a default branch.
#
# So for example, if you want to check development, and default to master:
#
#  ./all.sh development master
#
# Or if you have your own branch TerribleBug, building off development:
#
#  ./all.sh TerribleBug development
#
# Any repository that doesn't have a development branch in this last case
# is going to default to master.
#
cd ..

if [[ -z $1 ]]; then
echo "
*********************************************************
*       Defaulting... Checking out Master
*
*       ./all.sh <branch> <default>
*
*       Will try to check out <branch>, will default
*       to <default>, and if neither exists, or are
*       missing, will checkout the master branch.
*
*********************************************************"
branch=master
default=master
else
echo "
*********************************************************
*       Checking out the" $1 "branch
*
*       ./all.sh <branch> <default>
*
*       Will try to check out <branch>, will default
*       to <default>, and if neither exists, or are
*       missing, will checkout the master branch.
*
*********************************************************"
branch=$1
if [[ -z $2 ]]; then
default=master
else
default=$2
fi
fi
checkout() {
    current=`pwd`
    cd $1 $2 > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo $1 | awk "{printf(\"%15s\",\"$1\")}"
        git fetch -q
        git checkout -q $2 > /dev/null 2>&1
        if [ $? -eq 0 ]; then
            echo -e -n " now on" $2    # checkout did not fail
        else
            git checkout -q $3 > /dev/null 2>&1
            if [ $? -ne 0 ]; then
                git checkout -q master > /dev/null 2>&1
                if [ $? -ne 0 ]; then
                   echo -e -n " ****checkout failed!!!"
                else
                   echo -n " defaulting to master"
                fi
            else
                echo -n " defaulting to" $3
            fi
        fi
        git status | awk '/^Your branch is [ab]/ {$1="";$2="";FS=" ";printf(" and%s",$0)}; /Your branch and/{printf("\n\t\t%s",$0)}'
        echo
        git pull 2>&1 | awk '$1=="error:" {print "\t\t"$0};/\|/{print("\t\t"$0)}'
        git status | awk '/:/ {if(!a[$0]){print"\t\t"$0}; a[$0]=1}'
        git status | awk '/^Untracked files.*/ {g=1}; /^\t.*/ { if(g) print"\t\tUntracked:  "$1 }'


        cd $current
   else
        echo $1 | awk "{printf(\"%15s\",\"$1\")}"
        echo " not found"
   fi
}

compile() {
    cerr=0
    current=`pwd`
    cd $1
    echo "Compiling: " $1
    go clean
    go install || cerr=1
    cd $current
    return $cerr
}

compileFactomdGitHash() {
    cerr=0
    current=`pwd`
    cd $1
    echo "Compiling: " $1
    go clean
    #rm $GOPATH/bin/$1
    go install -ldflags "-X github.com/FactomProject/factomd/engine.Build=`git rev-parse HEAD`" || cerr=1
    cd $current
    return $cerr
}


checkout factomd      $branch $default
checkout factom       $branch $default
checkout web          $branch $default
checkout go-spew      $branch $default
checkout go-bip39     $branch $default
checkout go-bip32     $branch $default
checkout go-bip44     $branch $default
checkout bolt         $branch $default
checkout Testing      $branch $default
checkout btcutil      $branch $default
checkout goleveldb    $branch $default
checkout FactomDocs   $branch $default
checkout gocoding     $branch $default
checkout btclog       $branch $default
checkout dynrsrc      $branch $default
checkout ed25519      $branch $default
checkout fastsha256   $branch $default
checkout go-flags     $branch $default
checkout go-socks     $branch $default
checkout seelog       $branch $default
checkout snappy-go    $branch $default
checkout websocket    $branch $default
checkout factom-cli   $branch $default
checkout factom-walletd $branch $default
checkout serveridentity $branch $default

echo "
********************************************************
*     Compiling factom-walletd, the cli, and factomd
********************************************************
"
compileFactomdGitHash factomd || exit 1
compile factom-cli            || exit 1
compile factom-walletd        || exit 1

echo ""
echo "
*******************************************************
*     Running Unit Tests    Now safe to Ctrl+C
*
*  YOU MUST KILL factomd FOR TESTS TO RUN PROPERLY!
*
*  Please run and pass all unit tests before pushing
*  to development or master!  Protect your code with
*  unit tests!  If you can!
*
*******************************************************
"
echo "
+================+
|  factomd   |
+================+
"
go test -short  ./factomd/...
