package assets

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultExportOutputPath(t *testing.T) {
	got := DefaultExportOutputPath("file123", "1:2", "png")
	assert.Equal(t, "file123_1-2.png", got)
}
