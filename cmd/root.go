package cmd

// root.go hosts shared helpers for command construction. The previous
// package-level rootCmd was removed as part of TASK-0002: command
// composition now happens exclusively in cmd/figma/main.go via
// newRootCommand(deps Deps). Tests build isolated roots with the same
// factory; production entry funnels through cmd.ExecuteRoot.
//
// No production init() runs in this package; root_factory.go is the only
// site that knows the shape of the tree.
