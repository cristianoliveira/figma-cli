package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// NewFlagUsageError turns Cobra flag parse failures into compact usage errors.
// Unknown flags include at most one nearby correction instead of dumping every
// flag, which keeps agent recovery to one short round trip.
func NewFlagUsageError(command *cobra.Command, err error) error {
	if err == nil {
		return nil
	}

	message := err.Error()
	input := ""
	if unknown := unknownFlagName(message); unknown != "" {
		input = "--" + unknown
		if suggestion := nearestFlag(command, unknown); suggestion != "" {
			message += fmt.Sprintf("\n\nDid you mean `--%s`?", suggestion)
		}
	}
	recovery := fmt.Sprintf("Run `%s --help` for valid flags.", command.CommandPath())
	return NewUsageErrorWithDetails(fmt.Errorf("%s", message), input, recovery)
}

func unknownFlagName(message string) string {
	const prefix = "unknown flag: --"
	if !strings.HasPrefix(message, prefix) {
		return ""
	}
	fields := strings.Fields(strings.TrimPrefix(message, prefix))
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func nearestFlag(command *cobra.Command, unknown string) string {
	candidates := make([]string, 0)
	seen := map[string]struct{}{}
	collect := func(flag *pflag.Flag) {
		if _, exists := seen[flag.Name]; exists {
			return
		}
		seen[flag.Name] = struct{}{}
		candidates = append(candidates, flag.Name)
	}
	command.LocalFlags().VisitAll(collect)
	command.InheritedFlags().VisitAll(collect)
	sort.Strings(candidates)

	best := ""
	bestDistance := 3
	for _, candidate := range candidates {
		distance := editDistance(unknown, candidate)
		if distance < bestDistance {
			best = candidate
			bestDistance = distance
		}
	}
	return best
}

func editDistance(left, right string) int {
	leftRunes := []rune(left)
	rightRunes := []rune(right)
	previous := make([]int, len(rightRunes)+1)
	for index := range previous {
		previous[index] = index
	}
	for leftIndex, leftRune := range leftRunes {
		current := make([]int, len(rightRunes)+1)
		current[0] = leftIndex + 1
		for rightIndex, rightRune := range rightRunes {
			cost := 0
			if leftRune != rightRune {
				cost = 1
			}
			current[rightIndex+1] = min(
				current[rightIndex]+1,
				previous[rightIndex+1]+1,
				previous[rightIndex]+cost,
			)
		}
		previous = current
	}
	return previous[len(rightRunes)]
}
