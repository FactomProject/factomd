while true; do
	fct balances
	sleep 1
	echo "Factoid Transaction"
	./flight.sh
	sleep 1
	./eclight.sh
	sleep 1
done
