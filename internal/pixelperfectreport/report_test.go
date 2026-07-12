package pixelperfectreport

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/imagediff"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderEscapesUserProvidedPathsAndEmbedsImages(t *testing.T) {
	dir := t.TempDir()
	reference := filepath.Join(dir, `reference-<script>.png`)
	actual := filepath.Join(dir, "actual.png")
	mask := filepath.Join(dir, "mask.png")
	writePNG(t, reference)
	writePNG(t, actual)
	writePNG(t, mask)

	html, err := Render(Input{
		ReferencePath: reference,
		ActualPath:    actual,
		MaskPath:      mask,
		Result:        imagediff.ImageComparison{ChangedPixels: 1, RMSE: 0.5},
	})

	require.NoError(t, err)
	content := string(html)
	assert.Contains(t, content, "Pixel Perfect Report")
	assert.Contains(t, content, "data:image/png;base64,")
	assert.Contains(t, content, "reference-&lt;script&gt;.png")
	assert.NotContains(t, content, `reference-<script>.png`)
}

func writePNG(t *testing.T, path string) {
	t.Helper()
	file, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, png.Encode(file, image.NewRGBA(image.Rect(0, 0, 1, 1))))
	require.NoError(t, file.Close())
}
