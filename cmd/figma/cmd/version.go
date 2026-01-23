package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Show the version number of the Figma CLI.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Printf("figma version %s\n", version)
		return nil
	},
}
