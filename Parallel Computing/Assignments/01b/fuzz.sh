#!/bin/bash

set -x

while true
do
	NAME=fuzz-$(date +%s).graph

	# ./bin/gen 25 $((RANDOM%100 + 1)) > $NAME
	./bin/gen 18 90 > $NAME
	taskset -c 0 time ./bin/sequential $NAME
done
