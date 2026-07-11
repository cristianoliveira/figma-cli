package imagediff

import (
	"fmt"
	"image"
	"image/color"
)

// WriteImageOverlay writes a transparent directional difference image.
// Red pixels are stronger in the reference; green pixels are stronger in the actual image.
func WriteImageOverlay(referencePath, actualPath, outputPath string, region *Bounds, ignored []Bounds) error {
	reference, err := decodePNG(referencePath)
	if err != nil {
		return fmt.Errorf("decode reference: %w", err)
	}
	actual, err := decodePNG(actualPath)
	if err != nil {
		return fmt.Errorf("decode actual: %w", err)
	}
	if reference.Bounds().Dx() != actual.Bounds().Dx() || reference.Bounds().Dy() != actual.Bounds().Dy() {
		return fmt.Errorf("image dimensions differ: reference is %dx%d, actual is %dx%d", reference.Bounds().Dx(), reference.Bounds().Dy(), actual.Bounds().Dx(), actual.Bounds().Dy())
	}
	area := Bounds{Width: reference.Bounds().Dx(), Height: reference.Bounds().Dy()}
	if region != nil {
		area = *region
		if area.X < 0 || area.Y < 0 || area.Width <= 0 || area.Height <= 0 || area.X+area.Width > reference.Bounds().Dx() || area.Y+area.Height > reference.Bounds().Dy() {
			return fmt.Errorf("region %d,%d,%d,%d is outside image bounds %dx%d", area.X, area.Y, area.Width, area.Height, reference.Bounds().Dx(), reference.Bounds().Dy())
		}
	}
	overlay := image.NewNRGBA(image.Rect(0, 0, area.Width, area.Height))
	for y := 0; y < area.Height; y++ {
		for x := 0; x < area.Width; x++ {
			if pointIgnored(area.X+x, area.Y+y, ignored) {
				continue
			}
			referencePixel := color.NRGBAModel.Convert(reference.At(reference.Bounds().Min.X+area.X+x, reference.Bounds().Min.Y+area.Y+y)).(color.NRGBA)
			actualPixel := color.NRGBAModel.Convert(actual.At(actual.Bounds().Min.X+area.X+x, actual.Bounds().Min.Y+area.Y+y)).(color.NRGBA)
			referenceStrength := directionalDifference(referencePixel, actualPixel)
			actualStrength := directionalDifference(actualPixel, referencePixel)
			overlay.SetNRGBA(x, y, color.NRGBA{R: referenceStrength, G: actualStrength, A: max(referenceStrength, actualStrength)})
		}
	}
	if err := encodePNG(outputPath, overlay); err != nil {
		return fmt.Errorf("write overlay: %w", err)
	}
	return nil
}

func directionalDifference(first, second color.NRGBA) uint8 {
	return max(positiveDifference(first.R, second.R), positiveDifference(first.G, second.G), positiveDifference(first.B, second.B), positiveDifference(first.A, second.A))
}

func positiveDifference(first, second uint8) uint8 {
	if first <= second {
		return 0
	}
	return first - second
}
