package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryContractKeepsScopeAndNamedResults(t *testing.T) {
	contract := NewQuery(
		Scope{FileKey: "abc", NodeIDs: []string{"1:2"}},
		[]map[string]string{{"id": "1:3"}},
	)

	encoded, err := json.Marshal(contract)
	require.NoError(t, err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["1:2"]},"total":1,"results":[{"id":"1:3"}]}`, string(encoded))
}

func TestDetailContractKeepsScopeAndResult(t *testing.T) {
	contract := Detail[map[string]string]{
		Scope:  Scope{FileKey: "abc", NodeIDs: []string{"1:2"}},
		Result: map[string]string{"id": "1:2"},
	}

	encoded, err := json.Marshal(contract)
	require.NoError(t, err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["1:2"]},"result":{"id":"1:2"}}`, string(encoded))
}

func TestFilteredQueryContractPreservesEmptyQueryContext(t *testing.T) {
	contract := NewFilteredQuery(Scope{FileKey: "abc"}, map[string]any{"name": "Button", "type": "COMPONENT"}, []string(nil))

	encoded, err := json.Marshal(contract)
	require.NoError(t, err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":[]},"query":{"name":"Button","type":"COMPONENT"},"total":0,"results":[]}`, string(encoded))
}

func TestLimitedQueryContractReportsTotalAndTruncation(t *testing.T) {
	contract := NewLimitedQuery(Scope{FileKey: "abc"}, nil, 3, []string{"first"})

	encoded, err := json.Marshal(contract)
	require.NoError(t, err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":[]},"total":3,"truncated":true,"results":["first"]}`, string(encoded))
}

func TestLimitedQueryContractNeverReportsTotalLowerThanReturnedResults(t *testing.T) {
	contract := NewLimitedQuery(Scope{FileKey: "abc"}, nil, 1, []string{"first", "second"})

	assert.Equal(t, 2, contract.Total)
	assert.False(t, contract.Truncated)
}

func TestQueryContractRequiredCollectionsAreNeverNull(t *testing.T) {
	contract := NewLimitedQuery(Scope{FileKey: "abc"}, map[string]any{"name": "Button"}, 3, []string{"first"})

	encoded, err := json.Marshal(contract)
	require.NoError(t, err)
	assertRequiredArray(t, encoded, "scope", "nodeIds")
	assertRequiredArray(t, encoded, "results")
}

func assertRequiredArray(t *testing.T, encoded []byte, path ...string) {
	t.Helper()

	var document map[string]any
	require.NoError(t, json.Unmarshal(encoded, &document))

	current := any(document)
	for _, key := range path {
		object, ok := current.(map[string]any)
		require.Truef(t, ok, "%s parent must be a JSON object, got %T", key, current)
		current = object[key]
	}

	_, ok := current.([]any)
	assert.Truef(t, ok, "%v must be a JSON array, got %T", path, current)
}

func TestQueryContractUsesEmptyArrayInsteadOfNull(t *testing.T) {
	contract := NewQuery(Scope{FileKey: "abc"}, []string(nil))

	encoded, err := json.Marshal(contract)
	require.NoError(t, err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":[]},"total":0,"results":[]}`, string(encoded))
}
