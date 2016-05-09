while true; do
	fct balances
	sleep 10
	echo "Factoid Transaction"
	./flight.sh
	sleep 10
	./eclight.sh
	sleep 1
done
