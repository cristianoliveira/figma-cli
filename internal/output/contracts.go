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
	Scope   Scope `json:"scope"`
	Results []T   `json:"results"`
}

// NewQuery creates a query contract with non-null collections.
func NewQuery[T any](scope Scope, results []T) Query[T] {
	if scope.NodeIDs == nil {
		scope.NodeIDs = []string{}
	}
	if results == nil {
		results = []T{}
	}
	return Query[T]{Scope: scope, Results: results}
}
