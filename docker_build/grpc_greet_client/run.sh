#!/bin/sh


n=1

/server

# continue until $n equals 5
while [ $n ]
do
	echo "Welcome $n times."
	sleep $n*10
	n=$(( n+1 ))	 # increments $n
done

