package testhelpers

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// LoadFixture loads a JSON fixture from the testdata directory.
func LoadFixture(t *testing.T, name string) []byte {
	t.Helper()
	path := filepath.Join("testdata", "fixtures", name)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "failed to load fixture %s", name)
	return data
}

// LoadFixtureJSON loads a JSON fixture and unmarshals it into v.
func LoadFixtureJSON(t *testing.T, name string, v interface{}) {
	t.Helper()
	data := LoadFixture(t, name)
	err := json.Unmarshal(data, v)
	require.NoError(t, err, "failed to unmarshal fixture %s", name)
}

// AssertGolden compares the actual output with a golden file.
// If the update flag is set (via -update), it updates the golden file.
func AssertGolden(t *testing.T, goldenFile string, actual []byte) {
	t.Helper()
	goldenPath := filepath.Join("testdata", "golden", goldenFile)

	if *update {
		require.NoError(t, os.MkdirAll(filepath.Dir(goldenPath), 0755))
		require.NoError(t, os.WriteFile(goldenPath, actual, 0644))
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if os.IsNotExist(err) {
		t.Errorf("golden file %s does not exist, run with -update to create", goldenFile)
		return
	}
	require.NoError(t, err)
	assert.Equal(t, string(expected), string(actual))
}

// update is a flag that can be set via test flag -update.
var update = flag.Bool("update", false, "update golden files")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}
