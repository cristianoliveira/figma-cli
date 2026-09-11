package figma

import (
	"errors"
	"net/http"

	"github.com/cristianoliveira/figma-cli/internal/operr"
)

// classifyTransportError translates a concrete Figma transport failure
// into a neutral operational error, retaining the original cause for
// errors.Is / errors.As. Response bodies and tokens are never included
// in the safe user-facing text.
func classifyTransportError(err error) error {
	var response *ResponseError
	if errors.As(err, &response) {
		switch response.StatusCode {
		case http.StatusUnauthorized:
			return operr.New(operr.CategoryAuthentication,
				"Figma rejected authentication or access.",
				"Check FIGMA_ACCESS_TOKEN and file permissions, then retry.",
				err)
		case http.StatusForbidden:
			return operr.New(operr.CategoryAuthorization,
				"Figma rejected authentication or access.",
				"Check FIGMA_ACCESS_TOKEN and file permissions, then retry.",
				err)
		case http.StatusTooManyRequests:
			return operr.New(operr.CategoryRateLimit,
				"Figma rate limit reached.",
				"Wait before retrying the Figma request.",
				err)
		default:
			return operr.New(operr.CategoryDependencyUnavailable,
				"Figma request failed.",
				"Check Figma availability and retry.",
				err)
		}
	}
	// Network timeouts, DNS failures, and malformed responses are
	// provider-unavailability failures.
	return operr.New(operr.CategoryDependencyUnavailable,
		"Figma request failed.",
		"Check Figma availability and retry.",
		err)
}
