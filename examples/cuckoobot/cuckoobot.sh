#!/bin/bash

while :; do
	# Say cuckoo every hour
	if ! (( `date +'%M'` % 60 )); then
		echo "cuckoo"
	fi
	sleep 60
done
