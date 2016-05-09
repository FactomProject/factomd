while true; do
	fct balances
	sleep 6
	echo "Factoid Transaction"
	./flight.sh
	sleep 5
	./eclight.sh
	sleep 1
done
