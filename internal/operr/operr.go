// Package operr defines a small, data-focused, transport-neutral
// operational-error contract. Adapters translate concrete infrastructure
// failures (Figma statuses, HTTP timeouts, missing tokens, filesystem
// errors) into a ClassifiedError at their boundary; the CLI renderer
// consumes only this contract.
//
// The package is deliberately provider-agnostic: it contains data and
// wrapping/classification behaviour only, with no Cobra, HTTP,
// environment, filesystem, generated-API, or output dependencies.
package operr

// Category is a stable, deterministic operational-error classification.
type Category string

// Stable categories. The CLI renderer maps these to the process error
// envelope; keep values stable for scripts and agents.
const (
	CategoryAuthentication        Category = "authentication"
	CategoryAuthorization         Category = "authorization"
	CategoryRateLimit             Category = "rate_limit"
	CategoryDependencyUnavailable Category = "dependency_unavailable"
	CategoryInvalidInput          Category = "invalid_input"
	CategoryArtifactAccess        Category = "artifact_access"
	CategoryOperational           Category = "operational"
)

// ClassifiedError is a neutral, data-focused operational error. Message
// and Recovery are safe, user-facing strings (never tokens, bodies,
// headers, or private paths). Err retains the underlying cause so
// callers can still use errors.Is / errors.As.
type ClassifiedError struct {
	Category Category
	Message  string
	Recovery string
	Err      error
}

func (e *ClassifiedError) Error() string { return e.Message }
func (e *ClassifiedError) Unwrap() error { return e.Err }

// New builds a ClassifiedError with the given category and safe
// user-facing text, wrapping the underlying cause.
func New(category Category, message, recovery string, cause error) *ClassifiedError {
	return &ClassifiedError{Category: category, Message: message, Recovery: recovery, Err: cause}
}
