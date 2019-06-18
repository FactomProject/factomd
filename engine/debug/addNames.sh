#!/usr/bin/env bash
# replace node chain IDs with names in log files
# addnames.sh <list of files>

################################
# AWK scripts                  #
################################

#node1  '38bab1455b7bd7e5efd15c53c777c79d0c988e9210f1da49a99d95b3a6417be9'
#node2  '888888367795422bb2b15bae1af83396a94efa1cecab8cd171197eabd4b4bf9b'
#node3  '888888fc37fa418395eeccb95ab0a4c64d528b2aeefa0d1632c8a116a0e4f5b1'
#node4  '888888e238492b2d723d81f7122d4304e5405b18bd9c7cb22ca6bcbc1aab8493'
#node5  '888888271203752870ae5e6fa0cf96f93cf14bd052455ad476ab26de1ad2c077
#node6  '888888a21d5ac004defa311a1ea62f11e45a601742bdaf8ef087148943cefead'
#node7  '88888815ac8a1ab6b8f57cee67ba15aad23ab7d8e70ffdca064200738c201f74'
#node8  '888888f6e86156990970b5486c04dd7b2cfaaf0f756595580ca19591478b0a0e'
#node9  '88888830a663c5bb87e7d31558a07c836467ccd59a2534de99f4a1e8c0f3fe5d'
#node10 '888888dfa8acfbe302d390d739ab0afbfed23e865be124a619f069e11afb5835'


read -d '' scriptVariable << 'EOF'

/455b7b[^\\(]/ {x+= gsub(/455b7b[a-f0-9]*/,"455b7b(fnode0)");}
/367795[^\\(]/ {x+= gsub(/367795[a-f0-9]*/,"367795(fnode01)");}
/fc37fa[^\\(]/ {x+= gsub(/fc37fa[a-f0-9]*/,"fc37fa(fnode02)");}
/e23849[^\\(]/ {x+= gsub(/e23849[a-f0-9]*/,"e23849(fnode03)");}
/271203[^\\(]/ {x+= gsub(/271203[a-f0-9]*/,"271203(fnode04)");}
/a21d5a[^\\(]/ {x+= gsub(/a21d5a[a-f0-9]*/,"a21d5a(fnode05)");}
/15ac8a[^\\(]/ {x+= gsub(/15ac8a[a-f0-9]*/,"15ac8a(fnode06)");}
/f6e861[^\\(]/ {x+= gsub(/f6e861[a-f0-9]*/,"f6e861(fnode07)");}
/30a663[^\\(]/ {x+= gsub(/30a663[a-f0-9]*/,"30a663(fnode08)");}
/dfa8ac[^\\(]/ {x+= gsub(/dfa8ac[a-f0-9]*/,"dfa8ac(fnode09)");}
                                                         
 {print;}


 # print warm fuzzy's to stderr
 {if (FNR%1024 == 1) {printf("%40s:%d   \\r", FILENAME, x)>"/dev/stderr";}}

 END{printf("%40s:%d\\n", FILENAME, x)>"/dev/stderr";}
 
EOF
################################
# End of AWK Scripts           #
################################
awk -i inplace "$scriptVariable" $@
