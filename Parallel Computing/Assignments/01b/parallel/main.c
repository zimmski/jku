#include <omp.h>
#include <stdbool.h>
#include <stdio.h>
#include <string.h>
#include <stdlib.h>

// Let's use global variables for our state, no need to make this any prettier.
int **a;
int numberOfNodes;

typedef struct PathStruct {
	int Length;
	bool *Visited;
	int *Order;
	int OrderLength;
} Path;

Path *newPath() {
	Path *p = (Path *)calloc(1, sizeof(Path));

	p->Length = 0;
	p->Visited = (bool*)calloc(numberOfNodes, sizeof(bool));
	p->Order = (int*)calloc(numberOfNodes, sizeof(int));
	p->OrderLength = 0;

	return p;
}

void copyPath(Path *from, Path *to) {
	to->Length = from->Length;
	memcpy(to->Visited, from->Visited, numberOfNodes * sizeof(bool));
	memcpy(to->Order, from->Order, numberOfNodes * sizeof(int));
	to->OrderLength = from->OrderLength;
}

void printPath(Path *p) {
	printf("%d=", p->Length);
	for (int i = 0; i < p->OrderLength && i < numberOfNodes; i++) {
		printf("%d->", p->Order[i]);
	}
	if (p->OrderLength > numberOfNodes) {
		printf("%d\n", p->Order[0]);
	} else {
		printf("\n");
	}
}

typedef struct StackStruct {
	Path **Items;
	int Length;
} Stack;

// pushStack takes the given path and adds it to the stack.
void pushStack(Stack *stack, Path *path) {
	copyPath(path, stack->Items[stack->Length]);
	stack->Length++;
}

// popStack removes and returns the current path of the stack.
void popStack(Stack *stack, Path *path) {
	copyPath(stack->Items[stack->Length-1], path);
	stack->Length--;
}

// pathExists takes the given node index and returns if the node can be added to the given path.
bool pathExists(Path *path, int node) {
	// If the path is empty, we can add the node right away.
	if (path->OrderLength == 0) {
		return true;
	}

	int edgeLength = a[path->Order[path->OrderLength-1]][node];

	// If the edge is of length zero, there is no path.
	if (edgeLength == 0) {
		return false;
	}

	if (path->Visited[node]) {
		// Exit if we do not have all nodes in our path.
		if (path->OrderLength != numberOfNodes) {
			return false;
		}
		// Exit if the start node is not equal to the end node.
		if (path->Order[0] != node) {
			return false;
		}
	}

	return true;
}

// addNode takes the given node index and adds the node to the given path.
void addNode(Path *path, int node) {
	// If the path is empty, we can add the node right away.
	if (path->OrderLength == 0) {
		path->Visited[node] = true;
		path->Order[0] = node;
		path->OrderLength++;

		return;
	}

	int edgeLength = a[path->Order[path->OrderLength-1]][node];

	// Do not record the last edge.
	if (!path->Visited[node]) {
		path->Visited[node] = true;
		path->Order[path->OrderLength] = node;
	}
	path->OrderLength++;

	path->Length += edgeLength;
}

// removeLastNode removes the last inserted node
void removeLastNode(Path *path) {
	if (path->OrderLength == 0) {
		// This should never happen.

		return;
	}

	int node = path->Order[path->OrderLength-1];

	path->Visited[node] = false;
	path->Length -= a[path->Order[path->OrderLength-2]][node];
	path->Order[path->OrderLength-1] = 0; // This is unnecessary.
	path->OrderLength--;
}

// addNodeIfPathExist takes the given node index and tries to add the node to the given path.
// Returns -1 if the node cannot be added to the path.
int addNodeIfPathExist(Path *path, int node) {
	if (!pathExists(path, node)) {
		return -1;
	}

	addNode(path, node);

	return path->Length;
}

// queue holds a queue structure with a fixed preallocated item length.
typedef struct QueueStruct {
	Path **Items;
	int Size;
	int Current;
	int Free;
} Queue;

Queue *queue;
Path *sharedWinner;

Queue *newQueue(int size) {
	Queue *q = (Queue*)calloc(1, sizeof(Queue));

	q->Items = (Path**)calloc(size, sizeof(Path*));
	q->Size = size;
	q->Current = -1;
	q->Free = 0;

	for (int i = 0; i < size; i++) {
		q->Items[i] = newPath();
	}

	return q;
}

// addQueue takes the given path and adds it to the queue.
void addQueue(Path *path) {
	if (queue->Free == queue->Current) {
		printf("ERROR: This should never happen\n");

		exit(1);
	}

	copyPath(path, queue->Items[queue->Free]);
	if (queue->Current == -1) {
		queue->Current = queue->Free;
	}

	queue->Free++;
	queue->Free %= queue->Size; // Take care of the overrun.
}

// removeQueue removes and returns the current path of the queue.
void removeQueue(Path *path) {
	// Is the queue empty?
	if (queue->Current == -1) {
		printf("ERROR: This should never happen\n");

		exit(1);
	}

	copyPath(queue->Items[queue->Current], path);
	queue->Current++;
	queue->Current %= queue->Size; // Take care of the overrun.

	if (queue->Current == queue->Free) {
		// If the queue is empty, just reset it.

		queue->Current = -1;
		queue->Free = 0;
	}
}

typedef struct WorkerStruct {
	Stack *Stack;
	Path *Winner;
	Path *Path;
} Worker;

Worker *newWorker() {
	int maxPaths = numberOfNodes * (numberOfNodes - 1) / 2;

	Worker *w = (Worker *)calloc(1, sizeof(Worker));

	w->Stack = (Stack*)calloc(1, sizeof(Stack*));
	w->Stack->Items = (Path**)calloc(maxPaths, sizeof(Path*));
	w->Stack->Length = 0;
	w->Winner = newPath();
	w->Path = newPath();

	for (int i = 0; i < maxPaths; i++) {
		w->Stack->Items[i] = newPath();
	}

	return w;
}

// checkAndEvaluateCompletedPath checks if the current path can be completed, and evaluates if it is the new winner.
// Returns true if the path has been completed.
bool checkAndEvaluateCompletedPath(Worker *w) {
	// If the path is at the last node
	if (w->Path->OrderLength == numberOfNodes) {
		// We know that the last edge of a path is always the same node, so do this last step right now and therefore drop all other paths for this node.

		if (pathExists(w->Path, w->Path->Order[0])) {
			addNode(w->Path, w->Path->Order[0]);

			// Record if the current path is the best one.
			if (w->Winner->Length == 0 || w->Path->Length < w->Winner->Length) {
				#pragma omp critical(mutex_sharedWinner)
				{
					if (sharedWinner->Length == 0 || w->Path->Length < sharedWinner->Length) {
						/*printf("Thread %d: Found winner candidate: ", omp_get_thread_num());
						printPath(w->Path);*/

						copyPath(w->Path, sharedWinner);
					} /*else {
						printf("Thread %d: Discarded local winner: ", omp_get_thread_num());
						printPath(w->Path);
					}*/

					copyPath(sharedWinner, w->Winner);
				}
			} /*else {
				printf("Thread %d: Completed path is too long: ", omp_get_thread_num());
				printPath(w->Path);
			}*/
		}

		return true;
	}

	return false;
}

// expandQueue tries to expand the queue, if it succeeds it returns true and w->Path holds the next path for the worker.
bool expandQueue(Worker *w) {
	while (1) {
		// If the queue is empty we are done.
		if (queue->Current == -1) {
			return false;
		}

		removeQueue(w->Path);

		// If we had more than one path in the queue, we continue with the removed one for the local stack.
		if (queue->Current != -1) {
			return true;
		}

		// Expand the current path and push everything on the queue if needed.
		for (int i = 0; i < numberOfNodes; i++) {
			if (!pathExists(w->Path, i)) {
				continue;
			}

			addNode(w->Path, i);

			if (checkAndEvaluateCompletedPath(w)) {
				break;
			}

			// If the path is not done, put the path back on the queue but only proceed with paths that are shorter than the current best path.
			if (w->Winner->Length == 0 || w->Path->Length < w->Winner->Length) {
				addQueue(w->Path);
			}

			removeLastNode(w->Path);
		}

		// Check again, if we should exit or if we should expand.
	}
}

// solve tries to find the shortest cyclic path visiting all nodes in the currently loaded graph using a worker, while updating the shared winner.
void solveWorker(Worker *w) {
	while (1) {
		bool expanded;

		#pragma omp critical(mutex_queue)
		{
			expanded = expandQueue(w);
		}

		if (!expanded) {
			// There are no more paths in the queue to process.
			break;
		}

		pushStack(w->Stack, w->Path);

		while (w->Stack->Length != 0) {
			popStack(w->Stack, w->Path);

			// Expand the current path and push everything on the stack if needed.
			for (int i = 0; i < numberOfNodes; i++) {
				if (!pathExists(w->Path, i)) {
					continue;
				}

				addNode(w->Path, i);

				if (checkAndEvaluateCompletedPath(w)) {
					break;
				}

				// If the path is not done, put the path back on the stack but only proceed with paths that are shorter than the current best path.
				if (w->Winner->Length == 0 || w->Path->Length < w->Winner->Length) {
					/*printf("Thread %d: Pushed incomplete path: ", omp_get_thread_num());
					printPath(w->Path);*/

					pushStack(w->Stack, w->Path);
				} /*else {
					printf("Thread %d: Pruned incomplete path: ", omp_get_thread_num());
					printPath(w->Path);
				}*/

				removeLastNode(w->Path);
			}
		}
	}
}

// solve tries to find the shortest cyclic path visiting all nodes in the currently loaded graph.
Path *solve(int workerLength) {
	queue = newQueue(numberOfNodes);
	sharedWinner = newPath();

	// Init the queue by adding the first path.
	Path *p = newPath();
	addNode(p, 0);
	addQueue(p);

	double executionTime = 0.0;

	omp_set_num_threads(workerLength);

	double startExecution = omp_get_wtime();

	#pragma omp parallel for shared(a, numberOfNodes, queue, sharedWinner, executionTime)
	for (int i = 0; i < workerLength; i++) {
		// Everything inside the parallel pragma is automatically private according to the documentation.

		// At least try to not include the initialization for a worker in our execution time.
		double start = omp_get_wtime();

		// Preallocate worker's data.
		Worker *w = newWorker();

		double initTime = omp_get_wtime() - start;

		solveWorker(w);

		#pragma omp critical(mutex_workerTimes)
		{
			executionTime += initTime;
		}
	}

	executionTime += omp_get_wtime() - startExecution;

	printf("Execution took %0.7f seconds\n", executionTime);

	return sharedWinner;
}

char *readGraph(char *filepath) {
	FILE *f = fopen(filepath, "r");
	if (f == NULL) {
		return "Could not open file";
	}

	// Read in the number of nodes.
	int n = fscanf(f, "%d\n", &numberOfNodes);
	if (n == 0) {
		fclose(f);

		return "Could not read number of nodes";
	}

	// Initialize the matrix.
	a = (int**)calloc(numberOfNodes, sizeof(int*));
	for (int y = 0; y < numberOfNodes; y++) {
		a[y] = (int*)calloc(numberOfNodes, sizeof(int));
	}

	// Read in the matrix. Do nothing special, it is not the point to optimize this.
	for (int y = 0; y < numberOfNodes; y++) {
		for (int x = 0; x < numberOfNodes; x++) {
			if (x != numberOfNodes-1) {
				n = fscanf(f, "%d\t", &a[y][x]);
			} else {
				n = fscanf(f, "%d\n", &a[y][x]);
			}
			if (n == 0) {
				fclose(f);

				return "Could not read matrix number";
			}
		}
	}

	fclose(f);

	return NULL;
}
int main (int argc, char **argv) {
	if (argc != 3) {
		printf("Program must be called with <number of threads> <filepath to graph file> as argument.\n");

		return 1;
	}

	int workerLength = atoi(argv[1]);
	if (workerLength < 1 || workerLength > 32) {
		printf("ERROR: number of threads argument is not a number and must be greater in range [1,32]\n");

		return 1;
	}

	char *err = readGraph(argv[2]);
	if (err != NULL) {
		printf("ERROR: %s\n", err);

		return 1;
	}

	Path *winner = solve(workerLength);
	if (winner == NULL) {
		printf("There is no cyclic path\n");
	} else {
		printf("The shortest path has length %d with the path ", winner->Length);
		for (int i = 0; i < numberOfNodes; i++) {
			printf("%d->", winner->Order[i]);
		}
		printf("%d\n", winner->Order[0]);
	}
}
