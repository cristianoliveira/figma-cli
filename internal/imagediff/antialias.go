package imagediff

import (
	"image"
	"image/color"
)

// likelyAntialiased follows the neighborhood evidence model used by Pixelmatch
// (ISC, https://github.com/mapbox/pixelmatch):
// a ramp pixel has both darker and brighter neighbors, with one endpoint
// belonging to a stable area in both images.
func likelyAntialiased(reference, actual image.Image, x, y int, area Bounds) bool {
	return antialiasedIn(reference, actual, x, y, area) || antialiasedIn(actual, reference, x, y, area)
}

func antialiasedIn(candidate, other image.Image, x, y int, area Bounds) bool {
	center := color.NRGBAModel.Convert(candidate.At(x, y)).(color.NRGBA)
	equal, darkest, brightest := 0, [2]int{}, [2]int{}
	minDelta, maxDelta := 0.0, 0.0
	for neighborY := max(area.Y, y-1); neighborY <= min(area.Y+area.Height-1, y+1); neighborY++ {
		for neighborX := max(area.X, x-1); neighborX <= min(area.X+area.Width-1, x+1); neighborX++ {
			if neighborX == x && neighborY == y {
				continue
			}
			neighbor := color.NRGBAModel.Convert(candidate.At(neighborX, neighborY)).(color.NRGBA)
			delta := visibleLuminance(neighbor) - visibleLuminance(center)
			if delta == 0 {
				equal++
				if equal > 2 {
					return false
				}
			} else if delta < minDelta {
				minDelta, darkest = delta, [2]int{neighborX, neighborY}
			} else if delta > maxDelta {
				maxDelta, brightest = delta, [2]int{neighborX, neighborY}
			}
		}
	}
	if minDelta == 0 || maxDelta == 0 {
		return false
	}
	return hasStableSiblings(candidate, darkest[0], darkest[1], area) && hasStableSiblings(other, darkest[0], darkest[1], area) ||
		hasStableSiblings(candidate, brightest[0], brightest[1], area) && hasStableSiblings(other, brightest[0], brightest[1], area)
}

func hasStableSiblings(img image.Image, x, y int, area Bounds) bool {
	center := color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA)
	equal := 0
	for neighborY := max(area.Y, y-1); neighborY <= min(area.Y+area.Height-1, y+1); neighborY++ {
		for neighborX := max(area.X, x-1); neighborX <= min(area.X+area.Width-1, x+1); neighborX++ {
			if neighborX == x && neighborY == y {
				continue
			}
			if color.NRGBAModel.Convert(img.At(neighborX, neighborY)).(color.NRGBA) == center {
				equal++
				if equal > 2 {
					return true
				}
			}
		}
	}
	return false
}
