package cmd

import (
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/api"
)

func TestFindNodePath(t *testing.T) {
	// Build a simple tree:
	// root (id: "0")
	//   child1 (id: "1")
	//     grandchild1 (id: "2")
	//   child2 (id: "3")
	root := &api.Node{
		ID:   "0",
		Name: "Root",
		Type: "DOCUMENT",
		Children: []api.Node{
			{
				ID:   "1",
				Name: "Child1",
				Type: "FRAME",
				Children: []api.Node{
					{
						ID:   "2",
						Name: "Grandchild1",
						Type: "RECTANGLE",
					},
				},
			},
			{
				ID:   "3",
				Name: "Child2",
				Type: "FRAME",
			},
		},
	}

	// Test finding root
	path, node := findNodePath(root, "0")
	if node == nil || node.ID != "0" {
		t.Errorf("Expected to find root node")
	}
	if len(path) != 1 || path[0].ID != "0" {
		t.Errorf("Expected path length 1, got %d", len(path))
	}

	// Test finding child1
	path, node = findNodePath(root, "1")
	if node == nil || node.ID != "1" {
		t.Errorf("Expected to find child1 node")
	}
	if len(path) != 2 || path[0].ID != "0" || path[1].ID != "1" {
		t.Errorf("Expected path [0,1], got %v", path)
	}

	// Test finding grandchild1
	path, node = findNodePath(root, "2")
	if node == nil || node.ID != "2" {
		t.Errorf("Expected to find grandchild1 node")
	}
	if len(path) != 3 || path[0].ID != "0" || path[1].ID != "1" || path[2].ID != "2" {
		t.Errorf("Expected path [0,1,2], got %v", path)
	}

	// Test non-existent node
	path, node = findNodePath(root, "99")
	if node != nil || path != nil {
		t.Errorf("Expected nil for non-existent node")
	}
}

func TestCollectSiblings(t *testing.T) {
	// Build a tree with parent and three children
	parent := &api.Node{
		ID:   "p",
		Name: "Parent",
		Type: "FRAME",
		Children: []api.Node{
			{ID: "c1", Name: "Child1", Type: "RECTANGLE"},
			{ID: "c2", Name: "Child2", Type: "RECTANGLE"},
			{ID: "c3", Name: "Child3", Type: "RECTANGLE"},
		},
	}
	// Path where target is child2
	path := []*api.Node{
		{ID: "0", Name: "Root"}, // dummy root
		parent,
		&parent.Children[1], // target child2
	}
	siblings := collectSiblings(path)
	if len(siblings) != 3 {
		t.Errorf("Expected 3 siblings, got %d", len(siblings))
	}
	// Ensure sibling list includes all children (including target?)
	// Our implementation returns all children of parent; target is included.
	// That's fine.
	foundC1 := false
	for _, sib := range siblings {
		if sib.ID == "c1" {
			foundC1 = true
		}
	}
	if !foundC1 {
		t.Errorf("Sibling c1 not found")
	}
}

// We can also test printTreeNode and printTree but they output to cmd,
// which is harder to test. We'll rely on integration tests later.
