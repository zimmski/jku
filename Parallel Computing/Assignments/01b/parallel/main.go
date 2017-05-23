package main

// Let's do C style here and do not use any Go features.

import (
	"fmt"
	// "github.com/pkg/profile"
	"os"
	"runtime"
	"sync"
)

// Let's use global variables for our state, no need to make this any prettier.
var a [][]int
var numberOfNodes int
var queue Queue
var sharedWinner SharedPath

// queue holds a queue structure with a fixed preallocated item length.
type Queue struct {
	sync.Mutex

	Items   []*Path
	Size    int
	Current int
	Free    int
}

func newQueue(size int) Queue {
	q := Queue{
		Items:   make([]*Path, size),
		Size:    size,
		Current: -1,
		Free:    0,
	}
	for i := 0; i < size; i++ {
		q.Items[i] = newPath()
	}

	return q
}

// addQueue takes the given path and adds it to the queue.
func addQueue(path *Path) {
	if queue.Free == queue.Current {
		panic("This should never happen")
	}

	copyPath(path, queue.Items[queue.Free])
	if queue.Current == -1 {
		queue.Current = queue.Free
	}

	queue.Free++
	queue.Free %= queue.Size // Take care of the overrun.
}

// removeQueue removes and returns the current path of the queue.
func removeQueue(path *Path) {
	// Is the queue empty?
	if queue.Current == -1 {
		panic("This should never happen")
	}

	copyPath(queue.Items[queue.Current], path)
	queue.Current++
	queue.Current %= queue.Size // Take care of the overrun.

	if queue.Current == queue.Free {
		// If the queue is empty, just reset it.

		queue.Current = -1
		queue.Free = 0
	}
}

type SharedPath struct {
	sync.Mutex
	Path
}

func newSharedPath() SharedPath {
	return SharedPath{
		Path: *newPath(),
	}
}

type Stack struct {
	Items  []*Path
	Length int
}

// pushStack takes the given path and adds it to the stack.
func pushStack(stack *Stack, path *Path) {
	copyPath(path, stack.Items[stack.Length])
	stack.Length++
}

// popStack removes and returns the current path of the stack.
func popStack(stack *Stack, path *Path) {
	copyPath(stack.Items[stack.Length-1], path)
	stack.Length--
}

type Worker struct {
	Stack  *Stack
	Winner *Path
	Path   *Path
}

func newWorker() *Worker {
	maxPaths := numberOfNodes * (numberOfNodes - 1) / 2

	w := &Worker{
		Stack: &Stack{
			Items:  make([]*Path, maxPaths),
			Length: 0,
		},
		Winner: newPath(),
		Path:   newPath(),
	}
	for i := 0; i < maxPaths; i++ {
		w.Stack.Items[i] = newPath()
	}

	return w
}

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
	queue = newQueue(numberOfNodes)
	sharedWinner = newSharedPath()

	// Init the queue by adding the first path.
	p := newPath()
	addNode(p, 0)
	addQueue(p)

	// Preallocate worker's data.
	workerLength := runtime.GOMAXPROCS(-1)
	workers := make([]*Worker, workerLength)
	for i := 0; i < workerLength; i++ {
		workers[i] = newWorker()
	}

	var wg sync.WaitGroup
	wg.Add(workerLength)

	for i := 0; i < workerLength; i++ {
		go func(i int) {
			solveWorker(workers[i])

			wg.Done()
		}(i)
	}

	wg.Wait()

	return &sharedWinner.Path
}

// expandQueue tries to expand the queue, if it succeeds it returns true and w.Path holds the next path for the worker.
func expandQueue(w *Worker) bool {
	for {
		// If the queue is empty we are done.
		if queue.Current == -1 {
			return false
		}

		removeQueue(w.Path)

		// If we had more than one path in the queue, we continue with the removed one for the local stack.
		if queue.Current != -1 {
			return true
		}

		// Expand the current path and push everything on the queue if needed.
		for i := 0; i < numberOfNodes; i++ {
			if !pathExists(w.Path, i) {
				continue
			}

			addNode(w.Path, i)

			if checkAndEvaluateCompletedPath(w) {
				break
			}

			// If the path is not done, put the path back on the queue but only proceed with paths that are shorter than the current best path.
			if w.Winner.Length == 0 || w.Path.Length < w.Winner.Length {
				addQueue(w.Path)
			}

			removeLastNode(w.Path)
		}

		// Check again, if we should exit or if we should expand.
	}
}

// solve tries to find the shortest cyclic path visiting all nodes in the currently loaded graph using a worker, while updating the shared winner.
func solveWorker(w *Worker) {
	for {
		queue.Lock()
		expanded := expandQueue(w)
		queue.Unlock()

		if !expanded {
			// There are no more paths in the queue to process.
			break
		}

		pushStack(w.Stack, w.Path)

		for w.Stack.Length != 0 {
			popStack(w.Stack, w.Path)

			// Expand the current path and push everything on the stack if needed.
			for i := 0; i < numberOfNodes; i++ {
				if !pathExists(w.Path, i) {
					continue
				}

				addNode(w.Path, i)

				if checkAndEvaluateCompletedPath(w) {
					break
				}

				// If the path is not done, put the path back on the stack but only proceed with paths that are shorter than the current best path.
				if w.Winner.Length == 0 || w.Path.Length < w.Winner.Length {
					pushStack(w.Stack, w.Path)
				}

				removeLastNode(w.Path)
			}
		}
	}
}

// checkAndEvaluateCompletedPath checks if the current path can be completed, and evaluates if it is the new winner.
// Returns true if the path has been completed.
func checkAndEvaluateCompletedPath(w *Worker) bool {
	// If the path is at the last node
	if w.Path.OrderLength == numberOfNodes {
		// We know that the last edge of a path is always the same node, so do this last step right now and therefore drop all other paths for this node.

		if pathExists(w.Path, w.Path.Order[0]) {
			addNode(w.Path, w.Path.Order[0])

			// Record if the current path is the best one.
			if w.Winner.Length == 0 || w.Path.Length < w.Winner.Length {
				sharedWinner.Lock()

				if sharedWinner.Length == 0 || w.Path.Length < sharedWinner.Length {
					copyPath(w.Path, &sharedWinner.Path)
				}

				copyPath(&sharedWinner.Path, w.Winner)

				sharedWinner.Unlock()
			}
		}

		return true
	}

	return false
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

// pathExists takes the given node index and returns if the node can be added to the given path.
func pathExists(path *Path, node int) bool {
	// If the path is empty, we can add the node right away
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
	// If the path is empty, we can add the node right away
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
