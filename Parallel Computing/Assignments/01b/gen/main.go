package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Program must be called with <number of nodes> <fraction of the matrix that is not zero in percent> as arguments.")

		os.Exit(1)
	}

	// Read in and validate program arguments.
	numberOfNodes, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Number of nodes argument is not a number")

		os.Exit(1)
	}
	if numberOfNodes < 2 {
		fmt.Println("Number of nodes must be greater than one")

		os.Exit(1)
	}
	fraction, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Fraction argument is not a number")

		os.Exit(1)
	}
	if fraction < 0 || fraction > 100 {
		fmt.Println("Fraction argument is not a percentage")

		os.Exit(1)
	}

	// Initialize the matrix.
	a := make([][]int, numberOfNodes)
	for y := 0; y < len(a); y++ {
		a[y] = make([]int, numberOfNodes)
	}

	// Our random generator which is initialized with the OS's time in nanoseconds.
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// The maximum length for an edge.
	maxLength := 100

	// Set a fraction of the matrix to none-zero lengths excluding the edges that point to the same node.
	for i := 0; i < ((numberOfNodes*numberOfNodes)-numberOfNodes)*fraction/100; i++ {
		var x, y int

		// Find an edge in the matrix which is not an edge to the same not and is unused.
		for x == y || a[y][x] != 0 {
			e := r.Intn(numberOfNodes * numberOfNodes)
			y = e / numberOfNodes
			x = e % numberOfNodes
		}

		a[y][x] = r.Intn(maxLength) + 1
	}

	// Print out the matrix.
	fmt.Println(numberOfNodes)
	for y := 0; y < numberOfNodes; y++ {
		for x := 0; x < numberOfNodes; x++ {
			fmt.Print(a[y][x])
			if x != numberOfNodes-1 {
				fmt.Print("\t")
			}
		}
		fmt.Println()
	}
}
