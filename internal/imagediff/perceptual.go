package imagediff

import (
	"image/color"
	"math"
)

const DefaultPerceptualThreshold = 0.1

type oklabColor struct {
	lightness float64
	a         float64
	b         float64
}

// perceptualColorDistance returns normalized OKLab HyAB distance after alpha
// compositing onto white. Black-to-white has distance 1.
func perceptualColorDistance(first, second color.NRGBA) float64 {
	firstLab := rgbaToOKLab(first)
	secondLab := rgbaToOKLab(second)
	lightness := math.Abs(firstLab.lightness - secondLab.lightness)
	chroma := math.Hypot(firstLab.a-secondLab.a, firstLab.b-secondLab.b)
	return lightness + chroma
}

func rgbaToOKLab(value color.NRGBA) oklabColor {
	alpha := float64(value.A) / 255
	red := srgbToLinear((float64(value.R)/255)*alpha + 1 - alpha)
	green := srgbToLinear((float64(value.G)/255)*alpha + 1 - alpha)
	blue := srgbToLinear((float64(value.B)/255)*alpha + 1 - alpha)

	long := math.Cbrt(0.4122214708*red + 0.5363325363*green + 0.0514459929*blue)
	medium := math.Cbrt(0.2119034982*red + 0.6806995451*green + 0.1073969566*blue)
	short := math.Cbrt(0.0883024619*red + 0.2817188376*green + 0.6299787005*blue)

	return oklabColor{
		lightness: 0.2104542553*long + 0.793617785*medium - 0.0040720468*short,
		a:         1.9779984951*long - 2.428592205*medium + 0.4505937099*short,
		b:         0.0259040371*long + 0.7827717662*medium - 0.808675766*short,
	}
}

func srgbToLinear(value float64) float64 {
	if value <= 0.04045 {
		return value / 12.92
	}
	return math.Pow((value+0.055)/1.055, 2.4)
}
