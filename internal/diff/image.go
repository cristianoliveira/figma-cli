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
	Bounds        Bounds `json:"bounds"`
	ChangedPixels int    `json:"changedPixels"`
}

type ImageComparison struct {
	Width         int      `json:"width"`
	Height        int      `json:"height"`
	ChangedPixels int      `json:"changedPixels"`
	ChangedRatio  float64  `json:"changedRatio"`
	RMSE          float64  `json:"rmse"`
	Bounds        *Bounds  `json:"bounds,omitempty"`
	Regions       []Region `json:"regions,omitempty"`
	Mask          string   `json:"mask"`
}

func CompareImages(referencePath, actualPath, maskPath string, threshold uint8) (ImageComparison, error) {
	return CompareImagesInRegion(referencePath, actualPath, maskPath, threshold, nil)
}

func CompareImagesInRegion(referencePath, actualPath, maskPath string, threshold uint8, region *Bounds) (ImageComparison, error) {
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
	changed, minX, minY, maxX, maxY := 0, width, height, -1, -1
	var squaredError float64
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r := color.NRGBAModel.Convert(reference.At(reference.Bounds().Min.X+area.X+x, reference.Bounds().Min.Y+area.Y+y)).(color.NRGBA)
			a := color.NRGBAModel.Convert(actual.At(actual.Bounds().Min.X+area.X+x, actual.Bounds().Min.Y+area.Y+y)).(color.NRGBA)
			delta := [4]uint8{absDiff(r.R, a.R), absDiff(r.G, a.G), absDiff(r.B, a.B), absDiff(r.A, a.A)}
			maxDelta := max(delta[0], delta[1], delta[2], delta[3])
			for _, value := range delta {
				squaredError += float64(value) * float64(value)
			}
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
	result := ImageComparison{Width: width, Height: height, ChangedPixels: changed, ChangedRatio: float64(changed) / float64(width*height), RMSE: math.Sqrt(squaredError/float64(width*height*4)) / 255, Mask: maskPath}
	if changed > 0 {
		result.Bounds = &Bounds{X: minX, Y: minY, Width: maxX - minX + 1, Height: maxY - minY + 1}
		result.Regions = findRegions(changedPixels, width, height)
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
	if len(regions) > 20 {
		return regions[:20]
	}
	return regions
}

func absDiff(a, b uint8) uint8 {
	if a > b {
		return a - b
	}
	return b - a
}
