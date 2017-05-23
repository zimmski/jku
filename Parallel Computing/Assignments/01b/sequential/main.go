package main

// Let's do C style here and do not use any Go features.

// Profiling this program results into the following profile. Which indicates that the last optimizable part would be the memory copy from the stack to the local path variable.
//
// 14.17s of 14.18s total (99.93%)
// Dropped 7 nodes (cum <= 0.07s)
//       flat  flat%   sum%        cum   cum%
//     12.65s 89.21% 89.21%     14.17s 99.93%  main.solve
//      1.52s 10.72% 99.93%      1.52s 10.72%  runtime.memmove
//          0     0% 99.93%     14.18s   100%  main.main
//          0     0% 99.93%     14.18s   100%  runtime.goexit
//          0     0% 99.93%     14.18s   100%  runtime.main
// (pprof)

import (
	"fmt"
	// "github.com/pkg/profile"
	"os"
)

// Let's use global variables for our state, no need to make this any prettier.
var a [][]int
var numberOfNodes int
var stack []*Path
var stackLength int

func main() {
	// defer profile.Start(profile.CPUProfile).Stop()

	if len(os.Args) != 2 {
		fmt.Println("Program must be called with <filepath to graph file> as argument.")

		os.Exit(1)
	}

	var err error
	err = readGraph(os.Args[1])
	if err != nil {
		fmt.Println(err)

		os.Exit(1)
	}

	winner := solve()
	if winner == nil {
		fmt.Println("There is no cyclic path")
	} else {
		fmt.Printf("The shortest path has length %d with the path ", winner.Length)
		for i := 0; i < numberOfNodes; i++ {
			fmt.Printf("%d->", winner.Order[i])
		}
		fmt.Printf("%d\n", winner.Order[0])
	}
}

// solve tries to find the shortest cyclic path visiting all nodes in the currently loaded graph.
func solve() *Path {
	// Preallocate the stack.
	maxPaths := numberOfNodes * (numberOfNodes - 1) / 2
	stack = make([]*Path, maxPaths)
	for i := 0; i < maxPaths; i++ {
		stack[i] = newPath()
	}
	stackLength = 0

	// Init the stack by adding the first path.
	p := newPath()
	addNode(p, 0)
	pushPath(p)

	winner := newPath()

	for stackLength != 0 {
		popPath(p)

		for i := 0; i < numberOfNodes; i++ {
			if !pathExists(p, i) {
				continue
			}

			addNode(p, i)

			// If the path is at the last node
			if p.OrderLength == numberOfNodes {
				// We know that the last edge of a path is always the same node, so do this last step right now and therefore drop all other paths for this node.

				if pathExists(p, p.Order[0]) {
					addNode(p, p.Order[0])

					// Record if the current path is the best one.
					if winner.Length == 0 || p.Length < winner.Length {
						copyPath(p, winner)
					}
				}

				break
			}

			// If the path is not done, put the path back on the stack but only proceed with paths that are shorter than the current best path.
			if winner.Length == 0 || p.Length < winner.Length {
				pushPath(p)
			}

			removeLastNode(p)
		}
	}

	if winner.Length == 0 {
		// There is no winner
		return nil
	}

	return winner
}

func readGraph(filepath string) error {
	f, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Read in the number of nodes.
	_, err = fmt.Fscanln(f, &numberOfNodes)
	if err != nil {
		return err
	}

	// Initialize the matrix.
	a = make([][]int, numberOfNodes)
	for y := 0; y < len(a); y++ {
		a[y] = make([]int, numberOfNodes)
	}

	// Read in the matrix. Do nothing special, it is not the point to optimize this.
	for y := 0; y < len(a); y++ {
		for x := 0; x < len(a); x++ {
			_, err = fmt.Fscanf(f, "%d", &a[y][x])
			if err != nil {
				return err
			}
			// if x != numberOfNodes-1 {
			// 	_, err = fmt.Fscanf(f, "\t")
			// 	if err != nil {
			// 		return nil, err
			// 	}
			// }
		}
		// _, err = fmt.Fscanln(f)
		// if err != nil {
		// 	return nil, err
		// }
	}

	return nil
}

type Path struct {
	Length      int
	Visited     []bool
	Order       []int
	OrderLength int
}

func newPath() *Path {
	return &Path{
		Length:      0,
		Visited:     make([]bool, numberOfNodes),
		Order:       make([]int, numberOfNodes),
		OrderLength: 0,
	}
}

func copyPath(from *Path, to *Path) {
	to.Length = from.Length
	copy(to.Visited, from.Visited)
	copy(to.Order, from.Order)
	to.OrderLength = from.OrderLength
}

func printPath(p *Path) {
	fmt.Printf("%d=", p.Length)
	for i := 0; i < p.OrderLength && i < numberOfNodes; i++ {
		fmt.Printf("%d->", p.Order[i])
	}
	if p.OrderLength > numberOfNodes {
		fmt.Printf("%d\n", p.Order[0])
	} else {
		fmt.Println()
	}
}

// pushPath takes the given path and adds it to our working stack.
func pushPath(path *Path) {
	copyPath(path, stack[stackLength])
	stackLength++
}

// popPath removes and returns the current path of the stack.
func popPath(path *Path) {
	copyPath(stack[stackLength-1], path)
	stackLength--
}

// pathExists takes the given node index and returns if the node can be added to the given path.
func pathExists(path *Path, node int) bool {
	// If the path is empty, we can add the node right away.
	if path.OrderLength == 0 {
		return true
	}

	edgeLength := a[path.Order[path.OrderLength-1]][node]

	// If the edge is of length zero, there is no path.
	if edgeLength == 0 {
		return false
	}

	if path.Visited[node] {
		// Exit if we do not have all nodes in our path.
		if path.OrderLength != numberOfNodes {
			return false
		}
		// Exit if the start node is not equal to the end node.
		if path.Order[0] != node {
			return false
		}
	}

	return true
}

// addNode takes the given node index and adds the node to the given path.
func addNode(path *Path, node int) {
	// If the path is empty, we can add the node right away.
	if path.OrderLength == 0 {
		path.Visited[node] = true
		path.Order[0] = node
		path.OrderLength++

		return
	}

	edgeLength := a[path.Order[path.OrderLength-1]][node]

	// Do not record the last edge.
	if !path.Visited[node] {
		path.Visited[node] = true
		path.Order[path.OrderLength] = node
	}
	path.OrderLength++

	path.Length += edgeLength
}

// removeLastNode removes the last inserted node
func removeLastNode(path *Path) {
	if path.OrderLength == 0 {
		// This should never happen.

		return
	}

	node := path.Order[path.OrderLength-1]

	path.Visited[node] = false
	path.Length -= a[path.Order[path.OrderLength-2]][node]
	path.Order[path.OrderLength-1] = 0 // This is unnecessary.
	path.OrderLength--
}

// addNodeIfPathExist takes the given node index and tries to add the node to the given path.
// Returns -1 if the node cannot be added to the path.
func addNodeIfPathExist(path *Path, node int) int {
	if !pathExists(path, node) {
		return -1
	}

	addNode(path, node)

	return path.Length
}
