package diff

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sort"
)

type Bounds struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Region struct {
	Bounds        Bounds  `json:"bounds"`
	ChangedPixels int     `json:"changedPixels"`
	ChangedRatio  float64 `json:"changedRatio"`
	RMSE          float64 `json:"rmse"`
	EdgeRMSE      float64 `json:"edgeRmse"`
}

type ImageComparison struct {
	Width           int              `json:"width"`
	Height          int              `json:"height"`
	ChangedPixels   int              `json:"changedPixels"`
	ComparedPixels  int              `json:"comparedPixels"`
	ChangedRatio    float64          `json:"changedRatio"`
	RMSE            float64          `json:"rmse"`
	RGBRMSE         float64          `json:"rgbRmse"`
	LuminanceRMSE   float64          `json:"luminanceRmse"`
	AlphaRMSE       float64          `json:"alphaRmse"`
	EdgeRMSE        float64          `json:"edgeRmse"`
	ComparedRegion  *Bounds          `json:"comparedRegion,omitempty"`
	Bounds          *Bounds          `json:"bounds,omitempty"`
	Regions         []Region         `json:"regions,omitempty"`
	Mask            string           `json:"mask"`
	Overlay         string           `json:"overlay,omitempty"`
	SuggestedOffset *SuggestedOffset `json:"suggestedOffset,omitempty"`
}

func CompareImages(referencePath, actualPath, maskPath string, threshold uint8) (ImageComparison, error) {
	return CompareImagesInRegion(referencePath, actualPath, maskPath, threshold, nil)
}

func CompareImagesInRegion(referencePath, actualPath, maskPath string, threshold uint8, region *Bounds) (ImageComparison, error) {
	return CompareImagesWithIgnoredRegions(referencePath, actualPath, maskPath, threshold, region, nil)
}

func CompareImagesWithIgnoredRegions(referencePath, actualPath, maskPath string, threshold uint8, region *Bounds, ignored []Bounds) (ImageComparison, error) {
	reference, err := decodePNG(referencePath)
	if err != nil {
		return ImageComparison{}, fmt.Errorf("decode reference: %w", err)
	}
	actual, err := decodePNG(actualPath)
	if err != nil {
		return ImageComparison{}, fmt.Errorf("decode actual: %w", err)
	}
	if reference.Bounds().Dx() != actual.Bounds().Dx() || reference.Bounds().Dy() != actual.Bounds().Dy() {
		return ImageComparison{}, fmt.Errorf("image dimensions differ: reference is %dx%d, actual is %dx%d", reference.Bounds().Dx(), reference.Bounds().Dy(), actual.Bounds().Dx(), actual.Bounds().Dy())
	}

	imageWidth, imageHeight := reference.Bounds().Dx(), reference.Bounds().Dy()
	area := Bounds{Width: imageWidth, Height: imageHeight}
	if region != nil {
		area = *region
		if area.X < 0 || area.Y < 0 || area.Width <= 0 || area.Height <= 0 || area.X+area.Width > imageWidth || area.Y+area.Height > imageHeight {
			return ImageComparison{}, fmt.Errorf("region %d,%d,%d,%d is outside image bounds %dx%d", area.X, area.Y, area.Width, area.Height, imageWidth, imageHeight)
		}
	}
	width, height := area.Width, area.Height
	mask := image.NewNRGBA(image.Rect(0, 0, width, height))
	changedPixels := make([]bool, width*height)
	changed, compared, minX, minY, maxX, maxY := 0, 0, width, height, -1, -1
	var rgbSquaredError, luminanceSquaredError, alphaSquaredError float64
	hasTransparency := false
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			absoluteX, absoluteY := area.X+x, area.Y+y
			if pointIgnored(absoluteX, absoluteY, ignored) {
				continue
			}
			compared++
			r := color.NRGBAModel.Convert(reference.At(reference.Bounds().Min.X+area.X+x, reference.Bounds().Min.Y+area.Y+y)).(color.NRGBA)
			a := color.NRGBAModel.Convert(actual.At(actual.Bounds().Min.X+area.X+x, actual.Bounds().Min.Y+area.Y+y)).(color.NRGBA)
			delta := [4]uint8{absDiff(r.R, a.R), absDiff(r.G, a.G), absDiff(r.B, a.B), absDiff(r.A, a.A)}
			maxDelta := max(delta[0], delta[1], delta[2], delta[3])
			for _, value := range delta[:3] {
				rgbSquaredError += float64(value) * float64(value)
			}
			luminanceDelta := luminance(r) - luminance(a)
			luminanceSquaredError += luminanceDelta * luminanceDelta
			alphaSquaredError += float64(delta[3]) * float64(delta[3])
			hasTransparency = hasTransparency || r.A != 255 || a.A != 255
			if maxDelta <= threshold {
				continue
			}
			changed++
			changedPixels[y*width+x] = true
			minX, minY, maxX, maxY = min(minX, x), min(minY, y), max(maxX, x), max(maxY, y)
			mask.SetNRGBA(x, y, color.NRGBA{R: 255, A: maxDelta})
		}
	}
	if err := encodePNG(maskPath, mask); err != nil {
		return ImageComparison{}, fmt.Errorf("write mask: %w", err)
	}
	channelCount := 3
	squaredError := rgbSquaredError
	if hasTransparency {
		channelCount = 4
		squaredError += alphaSquaredError
	}
	result := ImageComparison{Width: width, Height: height, ChangedPixels: changed, ComparedPixels: compared, Mask: maskPath}
	if compared > 0 {
		result.ChangedRatio = float64(changed) / float64(compared)
		result.RMSE = math.Sqrt(squaredError/float64(compared*channelCount)) / 255
		result.RGBRMSE = math.Sqrt(rgbSquaredError/float64(compared*3)) / 255
		result.LuminanceRMSE = math.Sqrt(luminanceSquaredError/float64(compared)) / 255
		result.AlphaRMSE = math.Sqrt(alphaSquaredError/float64(compared)) / 255
		result.EdgeRMSE = imageEdgeRMSE(reference, actual, area, ignored)
	}
	if region != nil {
		comparedRegion := area
		result.ComparedRegion = &comparedRegion
	}
	if changed > 0 {
		result.Bounds = &Bounds{X: area.X + minX, Y: area.Y + minY, Width: maxX - minX + 1, Height: maxY - minY + 1}
		result.Regions = findRegions(changedPixels, width, height)
		for index := range result.Regions {
			result.Regions[index].Bounds.X += area.X
			result.Regions[index].Bounds.Y += area.Y
		}
	}
	return result, nil
}

func decodePNG(path string) (image.Image, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	img, decodeErr := png.Decode(file)
	closeErr := file.Close()
	if decodeErr != nil {
		return nil, decodeErr
	}
	return img, closeErr
}

func encodePNG(path string, img image.Image) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		return err
	}
	return file.Close()
}

func imageEdgeRMSE(reference, actual image.Image, area Bounds, ignored []Bounds) float64 {
	var squaredError float64
	samples := 0
	for y := area.Y; y < area.Y+area.Height; y++ {
		for x := area.X; x < area.X+area.Width; x++ {
			if pointIgnored(x, y, ignored) {
				continue
			}
			referencePixel := color.NRGBAModel.Convert(reference.At(reference.Bounds().Min.X+x, reference.Bounds().Min.Y+y)).(color.NRGBA)
			actualPixel := color.NRGBAModel.Convert(actual.At(actual.Bounds().Min.X+x, actual.Bounds().Min.Y+y)).(color.NRGBA)
			for _, previous := range [][2]int{{x - 1, y}, {x, y - 1}} {
				if previous[0] < area.X || previous[1] < area.Y || pointIgnored(previous[0], previous[1], ignored) {
					continue
				}
				referencePrevious := color.NRGBAModel.Convert(reference.At(reference.Bounds().Min.X+previous[0], reference.Bounds().Min.Y+previous[1])).(color.NRGBA)
				actualPrevious := color.NRGBAModel.Convert(actual.At(actual.Bounds().Min.X+previous[0], actual.Bounds().Min.Y+previous[1])).(color.NRGBA)
				delta := (visibleLuminance(referencePixel) - visibleLuminance(referencePrevious)) - (visibleLuminance(actualPixel) - visibleLuminance(actualPrevious))
				squaredError += delta * delta
				samples++
			}
		}
	}
	if samples == 0 {
		return 0
	}
	return math.Sqrt(squaredError/float64(samples)) / 255
}

func visibleLuminance(pixel color.NRGBA) float64 {
	return luminance(pixel) * float64(pixel.A) / 255
}

func luminance(pixel color.NRGBA) float64 {
	return 0.2126*float64(pixel.R) + 0.7152*float64(pixel.G) + 0.0722*float64(pixel.B)
}

func pointIgnored(x, y int, regions []Bounds) bool {
	for _, region := range regions {
		if x >= region.X && x < region.X+region.Width && y >= region.Y && y < region.Y+region.Height {
			return true
		}
	}
	return false
}

func findRegions(changed []bool, width, height int) []Region {
	visited := make([]bool, len(changed))
	regions := make([]Region, 0)
	for start := range changed {
		if !changed[start] || visited[start] {
			continue
		}
		queue := []int{start}
		visited[start] = true
		minX, maxX, minY, maxY, count := width, -1, height, -1, 0
		for len(queue) > 0 {
			index := queue[0]
			queue = queue[1:]
			x, y := index%width, index/width
			minX, maxX, minY, maxY, count = min(minX, x), max(maxX, x), min(minY, y), max(maxY, y), count+1
			for _, point := range [][2]int{{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1}} {
				nx, ny := point[0], point[1]
				if nx < 0 || nx >= width || ny < 0 || ny >= height {
					continue
				}
				next := ny*width + nx
				if changed[next] && !visited[next] {
					visited[next] = true
					queue = append(queue, next)
				}
			}
		}
		regions = append(regions, Region{Bounds: Bounds{X: minX, Y: minY, Width: maxX - minX + 1, Height: maxY - minY + 1}, ChangedPixels: count})
	}
	sort.SliceStable(regions, func(i, j int) bool { return regions[i].ChangedPixels > regions[j].ChangedPixels })
	return regions
}

func absDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}
