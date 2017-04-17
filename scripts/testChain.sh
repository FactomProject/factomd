EC=$(factom-cli newecaddress)

myrand=$RANDOM

echo test test test $myrand | factom-cli addchain -f -n one -n $myrand $EC

echo entry $myrand | factom-cli addentry -f -n one -n $myrand $EC

echo Entry Credit Address: $EC

sleep 3

big=FA2jK2HcLnRdS94dEcU27rF3meoJfpUcZPSinpb7AwQvPRY6RL1Q

factom-cli buyec $big $EC 100

sleep 3

factom-cli balance $EC

echo test test test $myrand | factom-cli addchain -f -n two -n $myrand $EC

sleep 3

factom-cli balance $EC
