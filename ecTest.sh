#/bin/sh
cmd="factom-cli -w localhost:8089 -s localhost:8088 importaddress Fs2DNirmGDtnAZGXqca3XHkukTNMxoMGFFQxJA3bAjJnKzzsZBMH"
echo $cmd
eval $cmd
cmd="ecAddr=$(factom-cli -w localhost:8089 -s localhost:8088 newecaddress)"
echo $cmd
eval $cmd
cmd="factom-cli -w localhost:8089 -s localhost:8088  buyec FA3EPZYqodgyEGXNMbiZKE5TS2x2J9wF8J9MvPZb52iGR78xMgCb $ecAddr 1"
echo $cmd
eval $cmd
cmd="factom-cli -w localhost:8089 -s localhost:8088  balance $ecAddr"
echo $cmd
eval $cmd

