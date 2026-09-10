/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package main

import "github.com/cristianoliveira/figma-cli/cmd"

// main is the production composition root. It wires real collaborators into
// the Deps struct, builds the complete command tree via the root factory,
// and hands control to cmd.ExecuteRoot for process-level error rendering
// and exit codes.
//
// Keeping construction here means no package-level command instances or
// init() registrations exist anywhere else in cmd/. Tests can call
// newRootCommand with fakes to exercise isolated roots without state
// leakage.
func main() {
	root := cmd.NewRootCommand(cmd.NewProductionDeps())
	cmd.ExecuteRoot(root)
}
