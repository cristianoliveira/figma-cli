package cli

import (
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FullHint reconstructs current command scope and replaces local limiting with
// --full. Only parsed flags are used, so generated command is deterministic.
func FullHint(command *cobra.Command, args []string) string {
	parts := []string{command.CommandPath()}
	for _, arg := range args {
		parts = append(parts, shellQuote(arg))
	}
	command.Flags().Visit(func(flag *pflag.Flag) {
		if flag.Name == "full" || flag.Name == "limit" {
			return
		}
		if flag.Value.Type() == "bool" {
			parts = append(parts, "--"+flag.Name)
			return
		}
		if slice, ok := flag.Value.(pflag.SliceValue); ok {
			for _, value := range slice.GetSlice() {
				parts = append(parts, "--"+flag.Name, shellQuote(value))
			}
			return
		}
		parts = append(parts, "--"+flag.Name, shellQuote(flag.Value.String()))
	})
	parts = append(parts, "--full")
	return strings.Join(parts, " ")
}

func shellQuote(value string) string {
	if value != "" && strings.IndexFunc(value, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !strings.ContainsRune("_@%+=:,./-", r)
	}) == -1 {
		return value
	}
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}
