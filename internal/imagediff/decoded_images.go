package imagediff

import (
	"fmt"
	"image"
)

// DecodedImages contains invocation-scoped normalized image data.
type DecodedImages struct {
	Reference *image.NRGBA
	Actual    *image.NRGBA
}

func LoadDecodedImages(referencePath, actualPath string) (*DecodedImages, error) {
	type result struct {
		image *image.NRGBA
		err   error
	}
	referenceResult := make(chan result, 1)
	actualResult := make(chan result, 1)
	go func() {
		decoded, err := decodeNRGBA(referencePath)
		referenceResult <- result{image: decoded, err: err}
	}()
	go func() {
		decoded, err := decodeNRGBA(actualPath)
		actualResult <- result{image: decoded, err: err}
	}()
	reference, actual := <-referenceResult, <-actualResult
	if reference.err != nil {
		return nil, fmt.Errorf("decode reference: %w", reference.err)
	}
	if actual.err != nil {
		return nil, fmt.Errorf("decode actual: %w", actual.err)
	}
	if reference.image.Bounds().Size() != actual.image.Bounds().Size() {
		return nil, fmt.Errorf("image dimensions differ: reference is %dx%d, actual is %dx%d", reference.image.Bounds().Dx(), reference.image.Bounds().Dy(), actual.image.Bounds().Dx(), actual.image.Bounds().Dy())
	}
	return &DecodedImages{Reference: reference.image, Actual: actual.image}, nil
}
