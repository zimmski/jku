#!/bin/bash

set -x

for file in 01-original.graph 02-20-nodes-fraction-65.graph 03-20-nodes-fraction-80.graph 04-20-nodes-fraction-89.graph 05-25-nodes-fraction-35.graph 06-18-nodes-fraction-100.graph 07-18-nodes-fraction-100.graph 08-18-nodes-fraction-90.graph; do
	for i in 1 2 3 4 5; do
		echo "graphs/$file run $i seq"
		GOMP_CPU_AFFINITY="576-1023:2" time ./bin/sequential graphs/$file
	done

	for p in 1 2 4 8 16 32; do
		for i in 1 2 3 4 5; do
			echo "graphs/$file run $i pal $p"
			GOMP_CPU_AFFINITY="576-1023:2" time ./bin/parallel $p graphs/$file
		done
	done
done
