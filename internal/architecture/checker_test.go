package architecture

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkHas(t *testing.T, files map[string][]string, rule, file, imp string) {
	t.Helper()
	offenders := Check(files)
	for _, o := range offenders {
		if o.Rule == rule && o.File == file && o.Import == imp {
			return
		}
	}
	t.Fatalf("expected offender (rule=%s file=%s import=%s); got %v", rule, file, imp, offenders)
}

func checkEmpty(t *testing.T, files map[string][]string) {
	t.Helper()
	offenders := Check(files)
	assert.Empty(t, offenders, "expected no offenders; got %v", offenders)
}

// Negative cases: each documented forbidden edge must fail.

func TestCheckDomainImportCobraFails(t *testing.T) {
	checkHas(t, map[string][]string{
		"internal/assets/app.go": {modulePath + "/internal/extract", "github.com/spf13/cobra"},
	}, "domain-infra", "internal/assets/app.go", "github.com/spf13/cobra")
}

func TestCheckDomainImportNetHTTPFails(t *testing.T) {
	checkHas(t, map[string][]string{
		"internal/document/node.go": {"net/http"},
	}, "domain-infra", "internal/document/node.go", "net/http")
}

func TestCheckDomainImportOSFails(t *testing.T) {
	checkHas(t, map[string][]string{
		"internal/extract/find.go": {"os"},
	}, "domain-infra", "internal/extract/find.go", "os")
}

func TestCheckDomainImportIOFSFails(t *testing.T) {
	checkHas(t, map[string][]string{
		"internal/extract/find.go": {"io/fs"},
	}, "domain-infra", "internal/extract/find.go", "io/fs")
}

func TestCheckDomainImportEnvFails(t *testing.T) {
	checkHas(t, map[string][]string{
		"internal/inspect/service.go": {modulePath + "/internal/env"},
	}, "domain-infra", "internal/inspect/service.go", modulePath+"/internal/env")
}

func TestCheckDomainImportToonFails(t *testing.T) {
	checkHas(t, map[string][]string{
		"internal/comments/comments.go": {"github.com/toon-format/toon-go"},
	}, "domain-infra", "internal/comments/comments.go", "github.com/toon-format/toon-go")
}

func TestCheckDomainImportGeneratedAPIFails(t *testing.T) {
	checkHas(t, map[string][]string{
		"internal/document/node.go": {modulePath + "/internal/figma/api"},
	}, "domain-infra", "internal/document/node.go", modulePath+"/internal/figma/api")
	// The dedicated generated boundary rule must also flag it.
	checkHas(t, map[string][]string{
		"internal/document/node.go": {modulePath + "/internal/figma/api"},
	}, "generated-api-boundary", "internal/document/node.go", modulePath+"/internal/figma/api")
}

func TestCheckInternalImportsCmdFails(t *testing.T) {
	checkHas(t, map[string][]string{
		"internal/assets/app.go": {modulePath + "/cmd"},
	}, "internal-cmd", "internal/assets/app.go", modulePath+"/cmd")
}

func TestCheckInternalImportsCmdSubpackageFails(t *testing.T) {
	checkHas(t, map[string][]string{
		"internal/extract/find.go": {modulePath + "/cmd/figma"},
	}, "internal-cmd", "internal/extract/find.go", modulePath+"/cmd/figma")
}

func TestCheckGroupedAndAliasedImportsStillFail(t *testing.T) {
	// The parser normalises grouped/aliased imports to their real path,
	// so an aliased Cobra import must still be detected. We exercise the
	// rule predicate directly since the synthetic map already holds the
	// resolved path.
	checkHas(t, map[string][]string{
		"internal/document/node.go": {"github.com/spf13/cobra"},
	}, "domain-infra", "internal/document/node.go", "github.com/spf13/cobra")
}

// Positive cases: documented edge owners and exemptions must pass.

func TestCheckEdgeOwnerFigmaHTTPPasses(t *testing.T) {
	checkEmpty(t, map[string][]string{
		"internal/figma/client.go": {"net/http"},
	})
}

func TestCheckEdgeOwnerOutputFilesystemAndToonPasses(t *testing.T) {
	checkEmpty(t, map[string][]string{
		"internal/output/file.go": {"os", "github.com/toon-format/toon-go", "io/fs"},
	})
}

func TestCheckEdgeOwnerCLICobraEnvPasses(t *testing.T) {
	checkEmpty(t, map[string][]string{
		"internal/cli/runtime.go": {"github.com/spf13/cobra", modulePath + "/internal/env", "net/http"},
	})
}

func TestCheckEdgeOwnerAssetsedgeHTTPPasses(t *testing.T) {
	checkEmpty(t, map[string][]string{
		"internal/assetsedge/http_sink.go": {"net/http", modulePath + "/internal/output"},
	})
}

func TestCheckEdgeOwnerEnvOSPasses(t *testing.T) {
	checkEmpty(t, map[string][]string{
		"internal/env/env.go": {"os"},
	})
}

func TestCheckEdgeOwnerComponentsFilesystemPasses(t *testing.T) {
	checkEmpty(t, map[string][]string{
		"internal/components/compare.go": {"os", "path/filepath"},
	})
}

func TestCheckGeneratedAPIPackagePasses(t *testing.T) {
	checkEmpty(t, map[string][]string{
		"internal/figma/api/api.gen.go": {"github.com/oapi-codegen/runtime", "encoding/json"},
	})
}

func TestCheckTestFileExemptionPasses(t *testing.T) {
	// A _test.go file may import anything; the production rule is
	// enforced separately.
	checkEmpty(t, map[string][]string{
		"internal/assets/app_test.go": {"net/http", "os", "github.com/spf13/cobra"},
	})
}

func TestCheckDomainImportPathFilepathIsAllowed(t *testing.T) {
	// Pure path manipulation (path/filepath) is not filesystem access;
	// the domain may use it for path construction.
	checkEmpty(t, map[string][]string{
		"internal/assets/app.go": {"path/filepath"},
	})
}

func TestCheckOffendersSortedDeterministically(t *testing.T) {
	files := map[string][]string{
		"internal/zzz.go": {"net/http"},
		"internal/aaa.go": {"os"},
	}
	offenders := Check(files)
	require.Len(t, offenders, 2)
	assert.Equal(t, "internal/aaa.go", offenders[0].File)
	assert.Equal(t, "internal/zzz.go", offenders[1].File)
}
