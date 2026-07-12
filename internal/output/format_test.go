package output

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFormatAcceptsAllowedFormat(t *testing.T) {
	format, err := ParseFormat("csv", FormatCSV, FormatJSON)

	require.NoError(t, err)
	assert.Equal(t, FormatCSV, format)
}

func TestParseFormatRejectsUnknownFormat(t *testing.T) {
	_, err := ParseFormat("yaml", FormatCSV, FormatJSON)

	assert.EqualError(t, err, `invalid --format "yaml": expected csv or json`)
}
