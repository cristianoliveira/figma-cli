package output

// Scope identifies the Figma file and nodes used to produce query results.
type Scope struct {
	FileKey string   `json:"fileKey"`
	NodeIDs []string `json:"nodeIds"`
}

// Detail is the stable envelope for commands returning one result.
type Detail[T any] struct {
	Scope  Scope `json:"scope"`
	Result T     `json:"result"`
}

// Query is the stable envelope for commands returning a result collection.
type Query[T any] struct {
	Scope     Scope          `json:"scope"`
	Query     map[string]any `json:"query,omitempty"`
	Total     int            `json:"total"`
	Truncated bool           `json:"truncated,omitempty"`
	Results   []T            `json:"results"`
}

// NewQuery creates a query contract with non-null collections and an exact count.
func NewQuery[T any](scope Scope, results []T) Query[T] {
	return NewFilteredQuery(scope, nil, results)
}

// NewFilteredQuery preserves filters so empty results remain self-describing.
func NewFilteredQuery[T any](scope Scope, query map[string]any, results []T) Query[T] {
	if scope.NodeIDs == nil {
		scope.NodeIDs = []string{}
	}
	if results == nil {
		results = []T{}
	}
	return Query[T]{Scope: scope, Query: query, Total: len(results), Results: results}
}

// NewLimitedQuery reports the pre-limit total and whether results were truncated.
func NewLimitedQuery[T any](scope Scope, query map[string]any, total int, results []T) Query[T] {
	contract := NewFilteredQuery(scope, query, results)
	if total < len(contract.Results) {
		total = len(contract.Results)
	}
	contract.Total = total
	contract.Truncated = len(contract.Results) < total
	return contract
}
