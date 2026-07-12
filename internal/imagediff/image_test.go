package imagediff

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeNRGBANormalizesDecodedPixelsAndBounds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "offset.png")
	input := image.NewNRGBA(image.Rect(3, 4, 5, 5))
	input.SetNRGBA(3, 4, color.NRGBA{R: 10, G: 20, B: 30, A: 40})
	writeTestPNG(t, path, input)

	decoded, err := decodeNRGBA(path)

	require.NoError(t, err)
	assert.Equal(t, image.Rect(0, 0, 2, 1), decoded.Bounds())
	assert.Equal(t, color.NRGBA{R: 10, G: 20, B: 30, A: 40}, decoded.NRGBAAt(0, 0))
}

func TestCompareImagesWritesMaskAndMeasuresChangedArea(t *testing.T) {
	dir := t.TempDir()
	reference := filepath.Join(dir, "reference.png")
	actual := filepath.Join(dir, "actual.png")
	mask := filepath.Join(dir, "diff.png")
	writeTestPNG(t, reference, image.NewRGBA(image.Rect(0, 0, 3, 2)))
	changed := image.NewRGBA(image.Rect(0, 0, 3, 2))
	changed.Set(1, 0, color.RGBA{R: 255, A: 255})
	changed.Set(2, 1, color.RGBA{B: 255, A: 255})
	writeTestPNG(t, actual, changed)

	result, err := CompareImages(reference, actual, mask, 0)

	require.NoError(t, err)
	assert.Equal(t, 3, result.Width)
	assert.Equal(t, 2, result.Height)
	assert.Equal(t, 2, result.ChangedPixels)
	assert.InDelta(t, 1.0/3.0, result.ChangedRatio, 0.0001)
	assert.Greater(t, result.PerceptualRMSE, 0.0)
	assert.Equal(t, 2, result.PerceptualChangedPixels)
	assert.InDelta(t, 1.0/3.0, result.PerceptualChangedRatio, 0.0001)
	assert.Equal(t, DefaultPerceptualThreshold, result.PerceptualThreshold)
	assert.Equal(t, EvidenceBreakdown{RawAndPerceptualPixels: 2}, result.Evidence)
	assert.Equal(t, &Bounds{X: 1, Y: 0, Width: 2, Height: 2}, result.Bounds)
	assert.Equal(t, []int{0, 1}, result.ChangedRows)
	require.Len(t, result.Regions, 2)
	assert.Equal(t, Region{Bounds: Bounds{X: 1, Y: 0, Width: 1, Height: 1}, ChangedPixels: 1}, result.Regions[0])
	assert.Equal(t, Region{Bounds: Bounds{X: 2, Y: 1, Width: 1, Height: 1}, ChangedPixels: 1}, result.Regions[1])
	_, err = os.Stat(mask)
	require.NoError(t, err)
}

func TestCompareImagesWithThresholdsConfiguresPerceptualEvidence(t *testing.T) {
	dir := t.TempDir()
	reference := filepath.Join(dir, "reference.png")
	actual := filepath.Join(dir, "actual.png")
	writeTestPNG(t, reference, image.NewRGBA(image.Rect(0, 0, 1, 1)))
	changed := image.NewRGBA(image.Rect(0, 0, 1, 1))
	changed.Set(0, 0, color.White)
	writeTestPNG(t, actual, changed)

	result, err := CompareImagesWithThresholds(reference, actual, filepath.Join(dir, "mask.png"), 0, 1.1, nil, nil)

	require.NoError(t, err)
	assert.Equal(t, 1.1, result.PerceptualThreshold)
	assert.Zero(t, result.PerceptualChangedPixels)
	assert.Zero(t, result.PerceptualChangedRatio)
	assert.Equal(t, 1, result.ChangedPixels)
	assert.Equal(t, EvidenceBreakdown{RawOnlyPixels: 1}, result.Evidence)
}

func TestCompareImagesInRegionMeasuresOnlySelectedArea(t *testing.T) {
	dir := t.TempDir()
	reference := filepath.Join(dir, "reference.png")
	actual := filepath.Join(dir, "actual.png")
	writeTestPNG(t, reference, image.NewRGBA(image.Rect(0, 0, 4, 2)))
	changed := image.NewRGBA(image.Rect(0, 0, 4, 2))
	changed.Set(0, 0, color.White)
	changed.Set(3, 1, color.White)
	writeTestPNG(t, actual, changed)

	result, err := CompareImagesInRegion(reference, actual, filepath.Join(dir, "mask.png"), 0, &Bounds{X: 2, Y: 0, Width: 2, Height: 2})

	require.NoError(t, err)
	assert.Equal(t, 2, result.Width)
	assert.Equal(t, 2, result.Height)
	assert.Equal(t, 1, result.ChangedPixels)
	assert.Equal(t, &Bounds{X: 2, Y: 0, Width: 2, Height: 2}, result.ComparedRegion)
	assert.Equal(t, &Bounds{X: 3, Y: 1, Width: 1, Height: 1}, result.Bounds)
	assert.Equal(t, []int{1}, result.ChangedRows)
	require.Len(t, result.Regions, 1)
	assert.Equal(t, Bounds{X: 3, Y: 1, Width: 1, Height: 1}, result.Regions[0].Bounds)
}

func TestMeasureImageRegionReportsLocalMetrics(t *testing.T) {
	dir := t.TempDir()
	reference := filepath.Join(dir, "reference.png")
	actual := filepath.Join(dir, "actual.png")
	referenceImage := image.NewRGBA(image.Rect(0, 0, 4, 2))
	changed := image.NewRGBA(image.Rect(0, 0, 4, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 4; x++ {
			referenceImage.Set(x, y, color.Black)
			changed.Set(x, y, color.Black)
		}
	}
	writeTestPNG(t, reference, referenceImage)
	changed.Set(2, 0, color.White)
	writeTestPNG(t, actual, changed)

	metrics, err := MeasureImageRegion(reference, actual, Bounds{X: 2, Y: 0, Width: 2, Height: 2}, 0, nil)

	require.NoError(t, err)
	assert.Equal(t, 1, metrics.ChangedPixels)
	assert.InDelta(t, 0.25, metrics.ChangedRatio, 0.000001)
	assert.Greater(t, metrics.RMSE, 0.0)
	assert.Greater(t, metrics.EdgeRMSE, 0.0)
	assert.Greater(t, metrics.PerceptualRMSE, 0.0)
	assert.Equal(t, 1, metrics.PerceptualChangedPixels)
	assert.InDelta(t, 0.25, metrics.PerceptualChangedRatio, 0.000001)
}

func TestMeasureImageRegionsMatchesIndividualMeasurements(t *testing.T) {
	dir := t.TempDir()
	reference := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	actual := image.NewNRGBA(image.Rect(0, 0, 4, 1))
	actual.SetNRGBA(0, 0, color.NRGBA{R: 255, A: 255})
	actual.SetNRGBA(3, 0, color.NRGBA{B: 255, A: 255})
	referencePath, actualPath := filepath.Join(dir, "reference.png"), filepath.Join(dir, "actual.png")
	writeTestPNG(t, referencePath, reference)
	writeTestPNG(t, actualPath, actual)
	bounds := []Bounds{{Width: 2, Height: 1}, {X: 2, Width: 2, Height: 1}}

	metrics, err := MeasureImageRegionsWithThresholds(referencePath, actualPath, bounds, 0, DefaultPerceptualThreshold, nil)

	require.NoError(t, err)
	require.Len(t, metrics, 2)
	for index, region := range bounds {
		expected, measureErr := MeasureImageRegionWithThresholds(referencePath, actualPath, region, 0, DefaultPerceptualThreshold, nil)
		require.NoError(t, measureErr)
		assert.Equal(t, expected, metrics[index])
	}
}

func TestMeasureImageRegionReportsDominantColorPairs(t *testing.T) {
	dir := t.TempDir()
	referenceImage := image.NewRGBA(image.Rect(0, 0, 3, 1))
	actualImage := image.NewRGBA(image.Rect(0, 0, 3, 1))
	for x := 0; x < 3; x++ {
		referenceImage.Set(x, 0, color.RGBA{R: 0x06, G: 0x67, B: 0xC8, A: 255})
		actualImage.Set(x, 0, color.RGBA{R: 0x16, G: 0x76, B: 0xD2, A: 255})
	}
	reference, actual := filepath.Join(dir, "reference.png"), filepath.Join(dir, "actual.png")
	writeTestPNG(t, reference, referenceImage)
	writeTestPNG(t, actual, actualImage)

	metrics, err := MeasureImageRegion(reference, actual, Bounds{Width: 3, Height: 1}, 0, nil)

	require.NoError(t, err)
	require.Len(t, metrics.DominantColorPairs, 1)
	assert.Equal(t, ColorPair{Reference: "#0667C8", Actual: "#1676D2", Pixels: 3}, metrics.DominantColorPairs[0])
}

func TestIgnoredPixelMapClipsRegionsAndCombinesOverlaps(t *testing.T) {
	ignored := newIgnoredPixelMap(4, 3, []Bounds{
		{X: -1, Y: 1, Width: 3, Height: 3},
		{X: 1, Y: 1, Width: 3, Height: 1},
	})

	assert.False(t, ignored.Contains(0, 0))
	assert.True(t, ignored.Contains(0, 1))
	assert.True(t, ignored.Contains(3, 1))
	assert.True(t, ignored.Contains(1, 2))
	assert.False(t, ignored.Contains(2, 2))
	assert.False(t, ignored.Contains(-1, 1))
	assert.False(t, ignored.Contains(4, 1))
}

func TestCompareImagesIgnoresSelectedRegions(t *testing.T) {
	dir := t.TempDir()
	reference := filepath.Join(dir, "reference.png")
	actual := filepath.Join(dir, "actual.png")
	writeTestPNG(t, reference, image.NewRGBA(image.Rect(0, 0, 3, 1)))
	changed := image.NewRGBA(image.Rect(0, 0, 3, 1))
	changed.Set(0, 0, color.White)
	changed.Set(2, 0, color.White)
	writeTestPNG(t, actual, changed)

	result, err := CompareImagesWithIgnoredRegions(reference, actual, filepath.Join(dir, "mask.png"), 0, nil, []Bounds{{X: 0, Y: 0, Width: 1, Height: 1}})

	require.NoError(t, err)
	assert.Equal(t, 1, result.ChangedPixels)
	assert.Equal(t, 2, result.ComparedPixels)
	assert.InDelta(t, 0.5, result.ChangedRatio, 0.000001)
	assert.Equal(t, &Bounds{X: 2, Y: 0, Width: 1, Height: 1}, result.Bounds)
}

func TestCompareImagesUsesRGBRMSEForOpaqueRegions(t *testing.T) {
	dir := t.TempDir()
	reference := image.NewRGBA(image.Rect(0, 0, 2, 2))
	actual := image.NewRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			reference.Set(x, y, color.Black)
			actual.Set(x, y, color.Black)
		}
	}
	actual.Set(0, 0, color.White)
	referencePath, actualPath := filepath.Join(dir, "reference.png"), filepath.Join(dir, "actual.png")
	writeTestPNG(t, referencePath, reference)
	writeTestPNG(t, actualPath, actual)

	result, err := CompareImages(referencePath, actualPath, filepath.Join(dir, "mask.png"), 0)

	require.NoError(t, err)
	assert.InDelta(t, 0.5, result.RMSE, 0.000001)
	assert.InDelta(t, 0.5, result.RGBRMSE, 0.000001)
	assert.InDelta(t, 0.5, result.LuminanceRMSE, 0.000001)
	assert.Zero(t, result.AlphaRMSE)
	assert.Greater(t, result.EdgeRMSE, 0.0)
}

func TestWriteImageOverlayShowsReferenceInRedAndActualInGreen(t *testing.T) {
	dir := t.TempDir()
	reference := image.NewRGBA(image.Rect(0, 0, 2, 1))
	actual := image.NewRGBA(image.Rect(0, 0, 2, 1))
	reference.Set(0, 0, color.White)
	actual.Set(1, 0, color.White)
	referencePath, actualPath := filepath.Join(dir, "reference.png"), filepath.Join(dir, "actual.png")
	writeTestPNG(t, referencePath, reference)
	writeTestPNG(t, actualPath, actual)
	overlayPath := filepath.Join(dir, "overlay.png")

	err := WriteImageOverlay(referencePath, actualPath, overlayPath, nil, nil)

	require.NoError(t, err)
	overlay, err := decodePNG(overlayPath)
	require.NoError(t, err)
	assert.Equal(t, color.NRGBA{R: 255, A: 255}, color.NRGBAModel.Convert(overlay.At(0, 0)))
	assert.Equal(t, color.NRGBA{G: 255, A: 255}, color.NRGBAModel.Convert(overlay.At(1, 0)))
}

func TestCompareImagesRejectsDifferentDimensions(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.png")
	b := filepath.Join(dir, "b.png")
	writeTestPNG(t, a, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	writeTestPNG(t, b, image.NewRGBA(image.Rect(0, 0, 3, 2)))

	_, err := CompareImages(a, b, filepath.Join(dir, "diff.png"), 0)

	assert.EqualError(t, err, "image dimensions differ: reference is 2x2, actual is 3x2")
}

func BenchmarkCompareIdenticalImages(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "identical.png")
	img := image.NewNRGBA(image.Rect(0, 0, 1024, 768))
	for index := range img.Pix {
		img.Pix[index] = byte(index)
	}
	file, err := os.Create(path)
	require.NoError(b, err)
	require.NoError(b, png.Encode(file, img))
	require.NoError(b, file.Close())

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_, err := CompareImages(path, path, "", 0)
		require.NoError(b, err)
	}
}

func writeTestPNG(t *testing.T, path string, img image.Image) {
	t.Helper()
	file, err := os.Create(path)
	require.NoError(t, err)
	require.NoError(t, png.Encode(file, img))
	require.NoError(t, file.Close())
}
