package architecture

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestParseImportsHandlesAliasesAndGroupedImports writes a temporary
// source file with an aliased import inside a grouped block, then
// confirms parseImports returns the resolved import path (not the alias).
func TestParseImportsHandlesAliasesAndGroupedImports(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	src := `package sample

import (
	"fmt"
	cobracli "github.com/spf13/cobra"

	_ "net/http"
)

var _ = fmt.Sprint
var _ = cobracli.Command{}
`
	require.NoError(t, os.WriteFile(path, []byte(src), 0o644))

	imports, err := parseImports(path)
	require.NoError(t, err)
	assert.Contains(t, imports, "github.com/spf13/cobra", "aliased import resolves to real path")
	assert.Contains(t, imports, "net/http", "blank/grouped import resolves to real path")
	assert.Contains(t, imports, "fmt")
}
