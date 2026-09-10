package document

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMalformedErrorFormatsPathAndField(t *testing.T) {
	err := &MalformedError{Path: "0/2/1", Field: "id", Reason: "missing"}
	assert.EqualError(t, err, "malformed document at 0/2/1: id required (missing)")
}

func TestMalformedErrorWithoutPath(t *testing.T) {
	err := &MalformedError{Field: "name", Reason: "wrong type: expected string"}
	assert.EqualError(t, err, "malformed document: name required (wrong type: expected string)")
}

func TestNodeEmptyChildrenIsNonNil(t *testing.T) {
	n := &Node{ID: "1:1", Name: "Root", Type: "FRAME", Children: []*Node{}}
	assert.NotNil(t, n.Children, "constructors must allow non-nil empty Children so JSON serialises [] not null")
}
