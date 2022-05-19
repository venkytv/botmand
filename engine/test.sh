#!/bin/bash

read N

I=1
while (( I <= N )); do
	echo "$I"
	I=$(( I + 1 ))
done
