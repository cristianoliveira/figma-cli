package diff

import (
	"fmt"
	"image/color"
	"math"
)

type RegionMetrics struct {
	ChangedPixels int     `json:"changedPixels"`
	ChangedRatio  float64 `json:"changedRatio"`
	RMSE          float64 `json:"rmse"`
	EdgeRMSE      float64 `json:"edgeRmse"`
}

func MeasureImageRegion(referencePath, actualPath string, bounds Bounds, threshold uint8, ignored []Bounds) (RegionMetrics, error) {
	reference, err := decodePNG(referencePath)
	if err != nil {
		return RegionMetrics{}, fmt.Errorf("decode reference: %w", err)
	}
	actual, err := decodePNG(actualPath)
	if err != nil {
		return RegionMetrics{}, fmt.Errorf("decode actual: %w", err)
	}
	if reference.Bounds().Dx() != actual.Bounds().Dx() || reference.Bounds().Dy() != actual.Bounds().Dy() {
		return RegionMetrics{}, fmt.Errorf("image dimensions differ: reference is %dx%d, actual is %dx%d", reference.Bounds().Dx(), reference.Bounds().Dy(), actual.Bounds().Dx(), actual.Bounds().Dy())
	}
	if bounds.X < 0 || bounds.Y < 0 || bounds.Width <= 0 || bounds.Height <= 0 || bounds.X+bounds.Width > reference.Bounds().Dx() || bounds.Y+bounds.Height > reference.Bounds().Dy() {
		return RegionMetrics{}, fmt.Errorf("region %d,%d,%d,%d is outside image bounds %dx%d", bounds.X, bounds.Y, bounds.Width, bounds.Height, reference.Bounds().Dx(), reference.Bounds().Dy())
	}
	changed, compared, channels := 0, 0, 3
	var rgbError, alphaError float64
	transparent := false
	for y := bounds.Y; y < bounds.Y+bounds.Height; y++ {
		for x := bounds.X; x < bounds.X+bounds.Width; x++ {
			if pointIgnored(x, y, ignored) {
				continue
			}
			compared++
			r := color.NRGBAModel.Convert(reference.At(x, y)).(color.NRGBA)
			a := color.NRGBAModel.Convert(actual.At(x, y)).(color.NRGBA)
			deltas := []uint8{absDiff(r.R, a.R), absDiff(r.G, a.G), absDiff(r.B, a.B), absDiff(r.A, a.A)}
			if max(deltas[0], deltas[1], deltas[2], deltas[3]) > threshold {
				changed++
			}
			for _, d := range deltas[:3] {
				rgbError += float64(d) * float64(d)
			}
			alphaError += float64(deltas[3]) * float64(deltas[3])
			transparent = transparent || r.A != 255 || a.A != 255
		}
	}
	if compared == 0 {
		return RegionMetrics{}, nil
	}
	total := rgbError
	if transparent {
		channels = 4
		total += alphaError
	}
	return RegionMetrics{ChangedPixels: changed, ChangedRatio: float64(changed) / float64(compared), RMSE: math.Sqrt(total/float64(compared*channels)) / 255, EdgeRMSE: imageEdgeRMSE(reference, actual, bounds, ignored)}, nil
}
