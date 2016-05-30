fct newaddress fct b1
fct newaddress fct b2
fct newaddress fct b3
fct newaddress fct b4
fct newaddress fct b5
fct newaddress fct b6
fct newaddress ec e1
fct newaddress ec e2 
fct newaddress ec e3 
fct newaddress ec e4 
fct newaddress ec e5 
fct newaddress ec e6


fct newtransaction t
fct addinput t factoid-wallet-address-name01 186
fct addoutput t b1 30
fct addoutput t b2 30
fct addoutput t b3 30
fct addoutput t b4 30
fct addoutput t b5 30
fct addoutput t b6 30
fct addecoutput t e1 1
fct addecoutput t e2 1
fct addecoutput t e3 1
fct addecoutput t e4 1
fct addecoutput t e5 1
fct addecoutput t e6 1
fct addfee t factoid-wallet-address-name01
fct sign t
fct submit t

