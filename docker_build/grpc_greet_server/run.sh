#!/bin/sh


n=1

/server

# continue until $n equals 5
while [ 1 ]
do
	echo "Welcome $n times."
        curl http://google.com 	
	curl http://example.com 
	curl https://apache.com
	n=$(( n+1 ))	 # increments $n
done

