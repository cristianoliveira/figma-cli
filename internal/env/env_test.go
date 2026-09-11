package env

import (
	"errors"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/operr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFigmaTokenMissingClassifiesAuthentication(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "")

	_, err := GetFigmaToken()

	require.Error(t, err)
	var classified *operr.ClassifiedError
	require.ErrorAs(t, err, &classified)
	assert.Equal(t, operr.CategoryAuthentication, classified.Category)
	assert.Equal(t, "Figma authentication is not configured.", classified.Message)
	assert.NotContains(t, classified.Error(), "environment variable not set")

	// Cause retention.
	var tokenErr *ErrTokenNotSet
	require.ErrorAs(t, err, &tokenErr)
	assert.True(t, errors.As(err, &tokenErr))
}

func TestGetFigmaTokenPresent(t *testing.T) {
	t.Setenv("FIGMA_ACCESS_TOKEN", "secret-token")

	token, err := GetFigmaToken()

	require.NoError(t, err)
	assert.Equal(t, "secret-token", token)
}
