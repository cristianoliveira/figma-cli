package imagediff

import (
	"fmt"
	"image/color"
	"math"
	"sort"
)

type SuggestedOffset struct {
	X    int     `json:"x"`
	Y    int     `json:"y"`
	RMSE float64 `json:"rmse"`
}

func SuggestImageOffset(referencePath, actualPath string, radius int, region *Bounds, ignored []Bounds) (SuggestedOffset, error) {
	reference, err := decodePNG(referencePath)
	if err != nil {
		return SuggestedOffset{}, fmt.Errorf("decode reference: %w", err)
	}
	actual, err := decodePNG(actualPath)
	if err != nil {
		return SuggestedOffset{}, fmt.Errorf("decode actual: %w", err)
	}
	if reference.Bounds().Dx() != actual.Bounds().Dx() || reference.Bounds().Dy() != actual.Bounds().Dy() {
		return SuggestedOffset{}, fmt.Errorf("image dimensions differ: reference is %dx%d, actual is %dx%d", reference.Bounds().Dx(), reference.Bounds().Dy(), actual.Bounds().Dx(), actual.Bounds().Dy())
	}
	area := Bounds{Width: reference.Bounds().Dx(), Height: reference.Bounds().Dy()}
	if region != nil {
		area = *region
	}
	candidates := make([]SuggestedOffset, 0, (radius*2+1)*(radius*2+1))
	for y := -radius; y <= radius; y++ {
		for x := -radius; x <= radius; x++ {
			candidates = append(candidates, SuggestedOffset{X: x, Y: y})
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		ai, aj := absInt(candidates[i].X)+absInt(candidates[i].Y), absInt(candidates[j].X)+absInt(candidates[j].Y)
		if ai != aj {
			return ai < aj
		}
		if candidates[i].Y != candidates[j].Y {
			return candidates[i].Y < candidates[j].Y
		}
		return candidates[i].X < candidates[j].X
	})
	ignoredPixels := newIgnoredPixelMap(reference.Bounds().Dx(), reference.Bounds().Dy(), ignored)
	best := SuggestedOffset{RMSE: math.Inf(1)}
	for _, candidate := range candidates {
		candidate.RMSE = offsetRMSE(reference, actual, area, ignoredPixels, candidate.X, candidate.Y)
		if candidate.RMSE < best.RMSE {
			best = candidate
		}
	}
	return best, nil
}

func offsetRMSE(reference, actual interface{ At(int, int) color.Color }, area Bounds, ignored ignoredPixelMap, offsetX, offsetY int) float64 {
	var rgbError, alphaError float64
	compared := 0
	transparent := false
	for y := area.Y; y < area.Y+area.Height; y++ {
		for x := area.X; x < area.X+area.Width; x++ {
			actualX, actualY := x-offsetX, y-offsetY
			if actualX < area.X || actualX >= area.X+area.Width || actualY < area.Y || actualY >= area.Y+area.Height || ignored.Contains(x, y) {
				continue
			}
			r := color.NRGBAModel.Convert(reference.At(x, y)).(color.NRGBA)
			a := color.NRGBAModel.Convert(actual.At(actualX, actualY)).(color.NRGBA)
			for _, d := range []uint8{absDiff(r.R, a.R), absDiff(r.G, a.G), absDiff(r.B, a.B)} {
				rgbError += float64(d) * float64(d)
			}
			d := absDiff(r.A, a.A)
			alphaError += float64(d) * float64(d)
			transparent = transparent || r.A != 255 || a.A != 255
			compared++
		}
	}
	if compared == 0 {
		return math.Inf(1)
	}
	channels := 3
	total := rgbError
	if transparent {
		channels = 4
		total += alphaError
	}
	return math.Sqrt(total/float64(compared*channels)) / 255
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
