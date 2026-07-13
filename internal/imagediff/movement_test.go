package imagediff

import (
	"encoding/json"
	"image"
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuggestRegionMovementsReportsLocalTranslation(t *testing.T) {
	reference := image.NewNRGBA(image.Rect(0, 0, 12, 6))
	actual := image.NewNRGBA(image.Rect(0, 0, 12, 6))
	fillRect(reference, image.Rect(1, 1, 4, 4), color.NRGBA{R: 255, A: 255})
	fillRect(actual, image.Rect(3, 1, 6, 4), color.NRGBA{R: 255, A: 255})
	fillRect(reference, image.Rect(8, 1, 10, 3), color.NRGBA{G: 255, A: 255})
	fillRect(actual, image.Rect(8, 1, 10, 3), color.NRGBA{G: 255, A: 255})
	images := &DecodedImages{Reference: reference, Actual: actual}

	movements := images.SuggestRegionMovements([]Bounds{{X: 1, Y: 1, Width: 5, Height: 3}}, 3, nil)

	if assert.Len(t, movements, 1) {
		movement := movements[0]
		assert.Equal(t, 2, movement.DX)
		assert.Zero(t, movement.DY)
		assert.Equal(t, Bounds{X: 0, Y: 0, Width: 9, Height: 6}, movement.Bounds)
		assert.Greater(t, movement.Confidence, 0.1)
	}
}

func TestSuggestRegionMovementsRejectsLocalContentChange(t *testing.T) {
	reference := image.NewNRGBA(image.Rect(0, 0, 8, 5))
	actual := image.NewNRGBA(image.Rect(0, 0, 8, 5))
	fillRect(reference, image.Rect(2, 1, 5, 4), color.NRGBA{R: 255, A: 255})
	fillRect(actual, image.Rect(2, 1, 5, 4), color.NRGBA{B: 255, A: 255})
	images := &DecodedImages{Reference: reference, Actual: actual}

	movements := images.SuggestRegionMovements([]Bounds{{X: 2, Y: 1, Width: 3, Height: 3}}, 2, nil)

	assert.Empty(t, movements)
}

func TestSuggestRegionMovementsCapsTokenHeavyEvidence(t *testing.T) {
	reference := image.NewNRGBA(image.Rect(0, 0, 72, 8))
	actual := image.NewNRGBA(image.Rect(0, 0, 72, 8))
	regions := make([]Bounds, 0, 6)
	for index := range 6 {
		x := 3 + index*11
		fillRect(reference, image.Rect(x, 2, x+3, 5), color.NRGBA{R: 255, A: 255})
		fillRect(actual, image.Rect(x+1, 2, x+4, 5), color.NRGBA{R: 255, A: 255})
		regions = append(regions, Bounds{X: x, Y: 2, Width: 4, Height: 3})
	}
	images := &DecodedImages{Reference: reference, Actual: actual}

	movements := images.SuggestRegionMovements(regions, 2, nil)

	assert.Len(t, movements, maximumReportedRegionMovements)
}

func TestRegionMovementJSONOmitsDerivedEvidence(t *testing.T) {
	encoded, err := json.Marshal(RegionMovement{Bounds: Bounds{X: 1, Y: 2, Width: 3, Height: 4}, DX: 5, DY: -1, Confidence: 0.75})

	require.NoError(t, err)
	assert.JSONEq(t, `{"bounds":{"x":1,"y":2,"width":3,"height":4},"dx":5,"dy":-1,"confidence":0.75}`, string(encoded))
}

func fillRect(target *image.NRGBA, bounds image.Rectangle, value color.NRGBA) {
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			target.SetNRGBA(x, y, value)
		}
	}
}
