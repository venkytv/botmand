#!/bin/bash

# PREFIX can be overridden in the engine config file
: ${PREFIX:="-> "}

while read LINE; do
	case "$LINE" in

		*bye*)
			echo "See you again!"
			break
			;;

		*)
			echo "${PREFIX}${LINE}"
			;;

	esac
done
