export ROOT_DIR := $(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
$(eval $(ARGS):;@:) # turn arguments into do-nothing targets
export ARGS

all: build-parallel build-sequential
.PHONY: all

bench: build-sequential build-parallel
	./bench.sh
.PHONY: bench

build-gen: dir
	go build -o ./bin/gen ./gen/main.go
.PHONY: build-gen

build-parallel: dir
	gcc -Wall -O3 -fopenmp -std=c99 -o ./bin/parallel ./parallel/main.c
.PHONY: build-parallel

build-sequential: dir
	gcc -Wall -O3 -fopenmp -std=c99 -o ./bin/sequential ./sequential/main.c
.PHONY: build-sequential

dir:
	mkdir -p bin
.PHONY: dir

fuzz: build-gen build-sequential
	./fuzz.sh
.PHONY: fuzz
