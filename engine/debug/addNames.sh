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

/455b7b[^\\(]/ {x+= gsub(/455b7b/,"455b7b(fnode01)");}
/367795[^\\(]/ {x+= gsub(/367795/,"367795(fnode02)");}
/fc37fa[^\\(]/ {x+= gsub(/fc37fa/,"fc37fa(fnode03)");}
/e23849[^\\(]/ {x+= gsub(/e23849/,"e23849(fnode04)");}
/271203[^\\(]/ {x+= gsub(/271203/,"271203(fnode05)");}
/a21d5a[^\\(]/ {x+= gsub(/a21d5a/,"a21d5a(fnode06)");}
/15ac8a[^\\(]/ {x+= gsub(/15ac8a/,"15ac8a(fnode07)");}
/f6e861[^\\(]/ {x+= gsub(/f6e861/,"f6e861(fnode08)");}
/30a663[^\\(]/ {x+= gsub(/30a663/,"30a663(fnode09)");}
/dfa8ac[^\\(]/ {x+= gsub(/dfa8ac/,"dfa8ac(fnode10)");}

/455b / {x+= gsub(/455b/,"455b7b(fnode01)");}
/15ac / {x+= gsub(/15ac/,"367795(fnode02)");}
/2712 / {x+= gsub(/2712/,"fc37fa(fnode03)");}
/3677 / {x+= gsub(/3677/,"e23849(fnode04)");}
/a21d / {x+= gsub(/a21d/,"271203(fnode05)");}
/e238 / {x+= gsub(/e238/,"a21d5a(fnode06)");}
/fc37 / {x+= gsub(/fc37/,"15ac8a(fnode07)");}
/a21d / {x+= gsub(/a21d/,"f6e861(fnode08)");}
/e238 / {x+= gsub(/e238/,"30a663(fnode09)");}
/fc37 / {x+= gsub(/fc37/,"dfa8ac(fnode10)");}

/455b7b$/ {x+= gsub(/455b7b/,"455b7b(fnode01)");}
/15ac8a$/ {x+= gsub(/367795/,"367795(fnode02)");}
/271203$/ {x+= gsub(/fc37fa/,"fc37fa(fnode03)");}
/367795$/ {x+= gsub(/e23849/,"e23849(fnode04)");}
/a21d5a$/ {x+= gsub(/271203/,"271203(fnode05)");}
/e23849$/ {x+= gsub(/a21d5a/,"a21d5a(fnode06)");}
/fc37fa$/ {x+= gsub(/15ac8a/,"15ac8a(fnode07)");}
/a21d5a$/ {x+= gsub(/f6e861/,"f6e861(fnode08)");}
/e23849$/ {x+= gsub(/30a663/,"30a663(fnode09)");}
/fc37fa$/ {x+= gsub(/dfa8ac/,"dfa8ac(fnode10)");}

/34353562/ {x+= gsub(/34353562/,"455b7b(fnode01)");}
/33363737/ {x+= gsub(/33363737/,"367795(fnode02)");} 
/66633337/ {x+= gsub(/66633337/,"fc37fa(fnode03)");}
/65323338/ {x+= gsub(/65323338/,"e23849(fnode04)");}
/32373132/ {x+= gsub(/32373132/,"271203(fnode05)");}
/61323164/ {x+= gsub(/61323164/,"a21d5a(fnode06)");}
/31356163/ {x+= gsub(/31356163/,"15ac8a(fnode07)");}

 {if (x%1024 == 0) {printf("%40s:%d\\r", FILENAME, x)>"/dev/stderr";}}
 {print;}
   
EOF
################################
# End of AWK Scripts           #
################################
awk -i inplace "$scriptVariable" $@
