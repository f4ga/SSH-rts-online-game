package research

import (
	"sort"
)

// TechTree represents a directed acyclic graph of technologies.
type TechTree struct {
	Nodes map[string]*TechNode
	Edges map[string][]string // from prerequisite -> dependent
}

// NewTechTree creates a TechTree from a slice of nodes.
func NewTechTree(nodes []*TechNode) *TechTree {
	tree := &TechTree{
		Nodes: make(map[string]*TechNode),
		Edges: make(map[string][]string),
	}
	for _, node := range nodes {
		tree.Nodes[node.ID] = node
		for _, req := range node.Requirements {
			tree.Edges[req] = append(tree.Edges[req], node.ID)
		}
	}
	return tree
}

// AvailableTechnologies returns IDs of technologies that a player can research
// (prerequisites met, not yet researched).
func (t *TechTree) AvailableTechnologies(researched map[string]bool) []string {
	var available []string
	for id, node := range t.Nodes {
		if researched[id] {
			continue
		}
		ok := true
		for _, req := range node.Requirements {
			if !researched[req] {
				ok = false
				break
			}
		}
		if ok {
			available = append(available, id)
		}
	}
	sort.Strings(available)
	return available
}

// IsDAG checks if the technology graph is acyclic (should be true by design).
func (t *TechTree) IsDAG() bool {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var dfs func(string) bool
	dfs = func(id string) bool {
		visited[id] = true
		recStack[id] = true

		for _, neighbor := range t.Edges[id] {
			if !visited[neighbor] {
				if dfs(neighbor) {
					return true
				}
			} else if recStack[neighbor] {
				return true
			}
		}
		recStack[id] = false
		return false
	}

	for id := range t.Nodes {
		if !visited[id] {
			if dfs(id) {
				return false
			}
		}
	}
	return true
}

// TopologicalOrder returns a topological ordering of technologies.
func (t *TechTree) TopologicalOrder() ([]string, error) {
	if !t.IsDAG() {
		return nil, ErrCycleDetected
	}

	inDegree := make(map[string]int)
	for id := range t.Nodes {
		inDegree[id] = 0
	}
	for _, deps := range t.Edges {
		for _, dep := range deps {
			inDegree[dep]++
		}
	}

	queue := make([]string, 0)
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var order []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		for _, neighbor := range t.Edges[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(order) != len(t.Nodes) {
		return nil, ErrCycleDetected
	}
	return order, nil
}

// PrerequisiteChain returns the chain of prerequisites leading to a technology.
func (t *TechTree) PrerequisiteChain(techID string) []string {
	var chain []string
	visited := make(map[string]bool)

	var collect func(string)
	collect = func(id string) {
		if visited[id] {
			return
		}
		visited[id] = true
		node, ok := t.Nodes[id]
		if !ok {
			return
		}
		for _, req := range node.Requirements {
			collect(req)
		}
		chain = append(chain, id)
	}
	collect(techID)
	return chain
}

// Errors
var (
	ErrCycleDetected = newTreeError("cycle detected in technology graph")
)

type treeError string

func newTreeError(msg string) error {
	return treeError(msg)
}

func (e treeError) Error() string {
	return string(e)
}