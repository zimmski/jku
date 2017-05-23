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

Path **stack;
int stackLength;

// pushPath takes the given path and adds it to our working stack.
void pushPath(Path *path) {
	copyPath(path, stack[stackLength]);
	stackLength++;
}

// popPath removes and returns the current path of the stack.
void popPath(Path *path) {
	copyPath(stack[stackLength-1], path);
	stackLength--;
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

// solve tries to find the shortest cyclic path visiting all nodes in the currently loaded graph.
Path *solve() {
	// Preallocate the stack.
	int maxPaths = numberOfNodes * (numberOfNodes - 1) / 2;
	stack = (Path**)calloc(maxPaths, sizeof(Path*));
	for (int i = 0; i < maxPaths; i++) {
		stack[i] = newPath();
	}
	stackLength = 0;

	// Init the stack by adding the first path.
	Path *p = newPath();
	addNode(p, 0);
	pushPath(p);

	Path *winner = newPath();

	double start = omp_get_wtime();

	while (stackLength != 0) {
		popPath(p);

		for (int i = 0; i < numberOfNodes; i++) {
			if (!pathExists(p, i)) {
				continue;
			}

			addNode(p, i);

			// If the path is at the last node
			if (p->OrderLength == numberOfNodes) {
				// We know that the last edge of a path is always the same node, so do this last step right now and therefore drop all other paths for this node.

				if (pathExists(p, p->Order[0])) {
					addNode(p, p->Order[0]);

					// Record if the current path is the best one.
					if (winner->Length == 0 || p->Length < winner->Length) {
						/*printf("Found winner candidate: ");
						printPath(p);*/

						copyPath(p, winner);
					} /*else {
						printf("Completed path is too long: ");
						printPath(p);
					}*/
				}

				break;
			}

			// If the path is not done, put the path back on the stack but only proceed with paths that are shorter than the current best path.
			if (winner->Length == 0 || p->Length < winner->Length) {
				/*printf("Pushed incomplete path: ");
				printPath(p);*/

				pushPath(p);
			} /*else {
				printf("Pruned incomplete path: ");
				printPath(p);
			}*/

			removeLastNode(p);
		}
	}

	double executionTime = omp_get_wtime() - start;

	printf("Execution took %0.7f seconds\n", executionTime);

	if (winner->Length == 0) {
		// There is no winner
		return NULL;
	}

	return winner;
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
	if (argc != 2) {
		printf("Program must be called with <filepath to graph file> as argument.\n");

		return 1;
	}

	char *err = readGraph(argv[1]);
	if (err != NULL) {
		printf("ERROR: %s\n", err);

		return 1;
	}

	Path *winner = solve();
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
