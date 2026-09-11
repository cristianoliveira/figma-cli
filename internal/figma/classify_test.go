package figma

import (
	"errors"
	"net/http"
	"testing"

	"github.com/cristianoliveira/figma-cli/internal/operr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyTransportErrorStatusCategories(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		category operr.Category
	}{
		{name: "unauthorized", status: http.StatusUnauthorized, category: operr.CategoryAuthentication},
		{name: "forbidden", status: http.StatusForbidden, category: operr.CategoryAuthorization},
		{name: "rate limited", status: http.StatusTooManyRequests, category: operr.CategoryRateLimit},
		{name: "not found", status: http.StatusNotFound, category: operr.CategoryDependencyUnavailable},
		{name: "server error", status: http.StatusInternalServerError, category: operr.CategoryDependencyUnavailable},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := &ResponseError{StatusCode: test.status}
			got := classifyTransportError(original)

			var classified *operr.ClassifiedError
			require.ErrorAs(t, got, &classified)
			assert.Equal(t, test.category, classified.Category)

			// Cause retention: errors.As still reaches the concrete type.
			var response *ResponseError
			require.ErrorAs(t, got, &response)
			assert.Equal(t, test.status, response.StatusCode)
		})
	}
}

func TestClassifyTransportErrorNetworkFailure(t *testing.T) {
	got := classifyTransportError(errors.New("dial tcp: lookup api.figma.com: no such host"))

	var classified *operr.ClassifiedError
	require.ErrorAs(t, got, &classified)
	assert.Equal(t, operr.CategoryDependencyUnavailable, classified.Category)
	assert.NotContains(t, classified.Error(), "api.figma.com", "raw host must not leak")
}

func TestClassifyTransportErrorRetainsCauseForWrapped(t *testing.T) {
	original := &ResponseError{StatusCode: http.StatusForbidden}
	// Simulate a caller wrapping the classified error.
	wrapped := errors.Join(errors.New("fetching"), classifyTransportError(original))

	var classified *operr.ClassifiedError
	require.ErrorAs(t, wrapped, &classified)
	assert.Equal(t, operr.CategoryAuthorization, classified.Category)

	var response *ResponseError
	require.ErrorAs(t, wrapped, &response)
	assert.Equal(t, http.StatusForbidden, response.StatusCode)
}
