package cmd

import (
	"fmt"

	"github.com/cristianoliveira/figma-cli/internal/cli"
	"github.com/cristianoliveira/figma-cli/internal/output"
	"github.com/spf13/cobra"
)

const defaultResultLimit = 100

type resultLimit struct {
	maximum int
	full    bool
}

// limit returns the maximum number of results to emit. A negative value
// signals "no limit" (full output), which downstream consumers treat as
// an unbounded request.
func (r resultLimit) limit() int {
	if r.full {
		return -1
	}
	return r.maximum
}

func addResultLimitFlags(command *cobra.Command) {
	command.Flags().Int("limit", defaultResultLimit, "maximum results to emit")
	command.Flags().Bool("full", false, "emit every result")
}

func readResultLimit(command *cobra.Command) (resultLimit, error) {
	maximum, _ := command.Flags().GetInt("limit")
	full, _ := command.Flags().GetBool("full")
	if maximum < 1 {
		return resultLimit{}, cli.NewUsageError(fmt.Errorf("--limit must be greater than zero"))
	}
	if full && command.Flags().Changed("limit") {
		return resultLimit{}, cli.NewUsageError(fmt.Errorf("--full cannot be combined with --limit"))
	}
	return resultLimit{maximum: maximum, full: full}, nil
}

func limitResults[T any](options resultLimit, results []T) ([]T, int) {
	total := len(results)
	if options.full || total <= options.maximum {
		return results, total
	}
	return results[:options.maximum], total
}

func newLimitedQuery[T any](command *cobra.Command, scope output.Scope, query map[string]any, total int, results []T) output.Query[T] {
	contract := output.NewLimitedQuery(scope, query, total, results)
	if contract.Truncated {
		contract.Hint = cli.FullHint(command, command.Flags().Args())
	}
	return contract
}
