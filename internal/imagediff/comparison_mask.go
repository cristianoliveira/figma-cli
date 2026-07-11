package imagediff

import (
	"fmt"
	"image/color"
)

// IgnoredRegionsFromMask converts excluded mask pixels into horizontal runs.
// Visible non-black pixels are compared; black or transparent pixels are ignored.
func IgnoredRegionsFromMask(maskPath, referencePath string) ([]Bounds, error) {
	mask, err := decodePNG(maskPath)
	if err != nil {
		return nil, fmt.Errorf("decode comparison mask: %w", err)
	}
	reference, err := decodePNG(referencePath)
	if err != nil {
		return nil, fmt.Errorf("decode reference: %w", err)
	}
	if mask.Bounds().Dx() != reference.Bounds().Dx() || mask.Bounds().Dy() != reference.Bounds().Dy() {
		return nil, fmt.Errorf("comparison mask dimensions differ: mask is %dx%d, reference is %dx%d", mask.Bounds().Dx(), mask.Bounds().Dy(), reference.Bounds().Dx(), reference.Bounds().Dy())
	}
	regions := make([]Bounds, 0)
	for y := 0; y < mask.Bounds().Dy(); y++ {
		runStart := -1
		for x := 0; x <= mask.Bounds().Dx(); x++ {
			excluded := x < mask.Bounds().Dx() && maskPixelExcluded(color.NRGBAModel.Convert(mask.At(mask.Bounds().Min.X+x, mask.Bounds().Min.Y+y)).(color.NRGBA))
			if excluded && runStart < 0 {
				runStart = x
			}
			if !excluded && runStart >= 0 {
				regions = append(regions, Bounds{X: runStart, Y: y, Width: x - runStart, Height: 1})
				runStart = -1
			}
		}
	}
	return regions, nil
}

func maskPixelExcluded(pixel color.NRGBA) bool {
	return pixel.A == 0 || pixel.R == 0 && pixel.G == 0 && pixel.B == 0
}
