package imagediff

import (
	"image/color"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPerceptualColorDistanceUsesHumanColorSpace(t *testing.T) {
	assert.Zero(t, perceptualColorDistance(color.NRGBA{R: 10, G: 20, B: 30, A: 255}, color.NRGBA{R: 10, G: 20, B: 30, A: 255}))
	assert.InDelta(t, 1, perceptualColorDistance(color.NRGBA{A: 255}, color.NRGBA{R: 255, G: 255, B: 255, A: 255}), 0.000001)
	assert.Zero(t, perceptualColorDistance(color.NRGBA{R: 255}, color.NRGBA{B: 255}))
}
