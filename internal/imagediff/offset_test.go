package imagediff

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuggestImageOffsetReportsTranslationWithoutApplyingIt(t *testing.T) {
	dir := t.TempDir()
	reference := image.NewRGBA(image.Rect(0, 0, 5, 3))
	actual := image.NewRGBA(image.Rect(0, 0, 5, 3))
	reference.Set(1, 1, color.White)
	actual.Set(2, 1, color.White)
	referencePath, actualPath := filepath.Join(dir, "reference.png"), filepath.Join(dir, "actual.png")
	writeTestPNG(t, referencePath, reference)
	writeTestPNG(t, actualPath, actual)

	offset, err := SuggestImageOffset(referencePath, actualPath, 2, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, -1, offset.X)
	assert.Equal(t, 0, offset.Y)
	assert.Zero(t, offset.RMSE)
}
