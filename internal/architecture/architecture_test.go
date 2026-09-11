package architecture

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestDependencyDirectionIsClean scans the real module and asserts the
// documented rules hold. It is the integration positive case; the pure
// negative cases live in checker_test.go.
func TestDependencyDirectionIsClean(t *testing.T) {
	root, err := ModuleRoot()
	require.NoError(t, err)

	files, err := ScanModule(root)
	require.NoError(t, err)
	require.NotEmpty(t, files, "module scan produced no files")

	offenders := Check(files)
	if len(offenders) == 0 {
		return
	}

	var b strings.Builder
	for _, o := range offenders {
		b.WriteString("\n  ")
		b.WriteString(o.Rule)
		b.WriteString(": ")
		b.WriteString(o.File)
		b.WriteString(" imports ")
		b.WriteString(o.Import)
	}
	t.Fatalf("dependency-direction violations:%s", b.String())
}
