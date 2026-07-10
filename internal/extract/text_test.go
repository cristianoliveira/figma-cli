package extract

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTextEqual(t *testing.T) {
	a := []TextNode{{ID: "1:1", Text: "hi"}, {ID: "1:2", Text: "yo"}}

	assert.True(t, TextEqual(a, []TextNode{{ID: "1:1", Text: "hi"}, {ID: "1:2", Text: "yo"}}))
	assert.False(t, TextEqual(a, []TextNode{{ID: "1:1", Text: "hi"}, {ID: "1:2", Text: "no"}}), "changed text")
	assert.False(t, TextEqual(a, []TextNode{{ID: "1:1", Text: "hi"}}), "removed node")
	assert.False(t, TextEqual(a, []TextNode{{ID: "1:1", Text: "hi"}, {ID: "1:2", Text: "yo"}, {ID: "1:3", Text: "x"}}), "added node")
}
