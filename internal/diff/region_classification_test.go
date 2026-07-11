package diff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassifyImageRegionRecognizesSolidFillMismatch(t *testing.T) {
	metrics := RegionMetrics{
		ChangedPixels: 100,
		ChangedRatio:  1,
		RMSE:          0.12,
		EdgeRMSE:      0.01,
		DominantColorPairs: []ColorPair{{
			Reference: "#0667C8",
			Actual:    "#1676D2",
			Pixels:    95,
		}},
	}

	assert.Equal(t, "solid-fill", ClassifyImageRegion(metrics))
}

func TestClassifyImageRegionRecognizesGeometryMismatch(t *testing.T) {
	metrics := RegionMetrics{ChangedPixels: 20, ChangedRatio: 0.2, RMSE: 0.1, EdgeRMSE: 0.09}

	assert.Equal(t, "geometry", ClassifyImageRegion(metrics))
}
