package extract

import (
	"testing"

	documentmodel "github.com/cristianoliveira/figma-cli/internal/document"
	"github.com/stretchr/testify/assert"
)

func TestFindLayoutVariantsByNamePreservesRequestedOrder(t *testing.T) {
	document := map[string]any{"id": "section", "children": []any{
		map[string]any{"id": "desktop", "name": "Desktop", "type": "FRAME"},
		map[string]any{"id": "mobile", "name": "Mobile", "type": "FRAME"},
	}}

	variants, err := FindLayoutVariantsByName(document, []string{"Mobile", "Desktop"})

	assert.NoError(t, err)
	assert.Equal(t, []string{"mobile", "desktop"}, []string{
		documentmodel.StringValue(variants[0].(map[string]any)["id"]), documentmodel.StringValue(variants[1].(map[string]any)["id"]),
	})
}

func TestFindLayoutVariantsByNameRejectsMissingAndAmbiguousNames(t *testing.T) {
	document := map[string]any{"children": []any{
		map[string]any{"id": "one", "name": "Mobile"},
		map[string]any{"id": "two", "name": "mobile"},
	}}

	_, ambiguousErr := FindLayoutVariantsByName(document, []string{"Mobile"})
	_, missingErr := FindLayoutVariantsByName(document, []string{"Desktop"})

	assert.EqualError(t, ambiguousErr, `layout variant name "Mobile" matched 2 nodes; use explicit --id`)
	assert.EqualError(t, missingErr, `layout variant name "Desktop" was not found`)
}
