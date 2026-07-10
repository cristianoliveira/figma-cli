package figma

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateExportFormatAcceptsAPIFormats(t *testing.T) {
	for _, format := range []string{"png", "jpg", "svg", "pdf"} {
		t.Run(format, func(t *testing.T) {
			require.NoError(t, ValidateExportFormat(format))
		})
	}
}

func TestValidateExportFormatRejectsUnsupportedFormat(t *testing.T) {
	err := ValidateExportFormat("gif")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "png, jpg, svg, or pdf")
}
