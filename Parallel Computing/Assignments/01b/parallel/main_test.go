package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddNodeIfPathExist(t *testing.T) {
	assert.NoError(t, readGraph("../graphs/01-original.graph"))

	p := newPath()

	// Add the initial node 0.
	assert.Equal(t, 0, addNodeIfPathExist(p, 0))
	assert.Equal(t, 0, p.Length)
	assert.Equal(t, []bool{true, false, false, false}, p.Visited)
	assert.Equal(t, []int{0, 0, 0, 0}, p.Order)
	assert.Equal(t, 1, p.OrderLength)

	// Add node 0 again.
	assert.Equal(t, -1, addNodeIfPathExist(p, 0))
	assert.Equal(t, 0, p.Length)
	assert.Equal(t, []bool{true, false, false, false}, p.Visited)
	assert.Equal(t, []int{0, 0, 0, 0}, p.Order)
	assert.Equal(t, 1, p.OrderLength)

	// Add node 1.
	assert.Equal(t, 1, addNodeIfPathExist(p, 1))
	assert.Equal(t, 1, p.Length)
	assert.Equal(t, []bool{true, true, false, false}, p.Visited)
	assert.Equal(t, []int{0, 1, 0, 0}, p.Order)
	assert.Equal(t, 2, p.OrderLength)

	// Add node 0 again.
	assert.Equal(t, -1, addNodeIfPathExist(p, 0))

	// Add node 3.
	assert.Equal(t, 7, addNodeIfPathExist(p, 3))
	assert.Equal(t, 7, p.Length)
	assert.Equal(t, []bool{true, true, false, true}, p.Visited)
	assert.Equal(t, []int{0, 1, 3, 0}, p.Order)
	assert.Equal(t, 3, p.OrderLength)

	// Add node 0 again.
	assert.Equal(t, -1, addNodeIfPathExist(p, 0))

	// Add node 2.
	assert.Equal(t, 19, addNodeIfPathExist(p, 2))
	assert.Equal(t, 19, p.Length)
	assert.Equal(t, []bool{true, true, true, true}, p.Visited)
	assert.Equal(t, []int{0, 1, 3, 2}, p.Order)
	assert.Equal(t, 4, p.OrderLength)

	// Finish with adding node 0.
	assert.Equal(t, 20, addNodeIfPathExist(p, 0))
	assert.Equal(t, 20, p.Length)
	assert.Equal(t, []bool{true, true, true, true}, p.Visited)
	assert.Equal(t, []int{0, 1, 3, 2}, p.Order)
	assert.Equal(t, 5, p.OrderLength)
}

func TestRemoveLastNode(t *testing.T) {
	assert.NoError(t, readGraph("../graphs/01-original.graph"))

	p := newPath()

	// Add the initial node 0.
	assert.Equal(t, 0, addNodeIfPathExist(p, 0))
	assert.Equal(t, 0, p.Length)
	assert.Equal(t, []bool{true, false, false, false}, p.Visited)
	assert.Equal(t, []int{0, 0, 0, 0}, p.Order)
	assert.Equal(t, 1, p.OrderLength)

	// Add node 1.
	assert.Equal(t, 1, addNodeIfPathExist(p, 1))
	assert.Equal(t, 1, p.Length)
	assert.Equal(t, []bool{true, true, false, false}, p.Visited)
	assert.Equal(t, []int{0, 1, 0, 0}, p.Order)
	assert.Equal(t, 2, p.OrderLength)

	// Remove node 1.
	removeLastNode(p)
	assert.Equal(t, 0, p.Length)
	assert.Equal(t, []bool{true, false, false, false}, p.Visited)
	assert.Equal(t, []int{0, 0, 0, 0}, p.Order)
	assert.Equal(t, 1, p.OrderLength)
}

func TestCopyPath(t *testing.T) {
	assert.NoError(t, readGraph("../graphs/01-original.graph"))

	p0 := newPath()

	assert.NotEqual(t, -1, addNodeIfPathExist(p0, 0))

	p1 := newPath()
	copyPath(p0, p1)
	assert.NotEqual(t, -1, addNodeIfPathExist(p1, 1))

	assert.Equal(
		t,
		&Path{
			Length:      0,
			Visited:     []bool{true, false, false, false},
			Order:       []int{0, 0, 0, 0},
			OrderLength: 1,
		},
		p0,
	)
	assert.Equal(
		t,
		&Path{
			Length:      1,
			Visited:     []bool{true, true, false, false},
			Order:       []int{0, 1, 0, 0},
			OrderLength: 2,
		},
		p1,
	)
}

func TestSolve(t *testing.T) {
	assert.NoError(t, readGraph("../graphs/01-original.graph"))

	p := solve()
	assert.Equal(t, 15, p.Length)
	assert.Equal(t, []int{0, 3, 1, 2}, p.Order)
}
