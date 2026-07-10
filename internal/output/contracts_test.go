package output

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryContractKeepsScopeAndNamedResults(t *testing.T) {
	contract := Query[map[string]string]{
		Scope:   Scope{FileKey: "abc", NodeIDs: []string{"1:2"}},
		Results: []map[string]string{{"id": "1:3"}},
	}

	encoded, err := json.Marshal(contract)
	require.NoError(t, err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":["1:2"]},"results":[{"id":"1:3"}]}`, string(encoded))
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

func TestQueryContractUsesEmptyArrayInsteadOfNull(t *testing.T) {
	contract := NewQuery(Scope{FileKey: "abc"}, []string(nil))

	encoded, err := json.Marshal(contract)
	require.NoError(t, err)
	assert.JSONEq(t, `{"scope":{"fileKey":"abc","nodeIds":[]},"results":[]}`, string(encoded))
}
