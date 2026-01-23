---
title: "Using Cobra CLI for Command Output"
tags: ["cobra", "cli", "go", "output", "best-practices"]
---
# Using Cobra CLI for Command Output

**Problem**: Agents are using pure `fmt.Print*` statements for CLI output, which is incorrect for Cobra-based applications.

**Solution**: Use Cobra's built-in output methods that respect the command's configured output writers and enable proper testing and output redirection.

## Cobra Output Methods
Cobra provides six output methods on the `Command` struct:

- `Print`, `Println`, `Printf` – for standard output
- `PrintErr`, `PrintErrln`, `PrintErrf` – for error output

## Correct Usage Example
**Instead of** `fmt.Println("Starting process...")` **use** `cmd.Println("Starting process...")`

**Correct pattern:**
```go
func Run(cmd *Command, args []string) {
    cmd.Println("Starting process...")
    cmd.Printf("Processing %d items\n", count)
    if err != nil {
        cmd.PrintErrf("Error: %v\n", err)
    }
}
```

## Best Practices
1. **Never use `fmt.Print*`** in command `Run`/`RunE` functions
2. **Use `cmd.Print*` for normal output** (information, results, progress)
3. **Use `cmd.PrintErr*` for errors, warnings, and diagnostic messages**
4. **Be consistent** – Use the same pattern across all commands

## Error Handling Pattern
When returning errors from `RunE`, use Cobra's error output for user-facing messages:
```go
RunE: func(cmd *Command, args []string) error {
    if err := validate(args); err != nil {
        cmd.PrintErrln("Validation failed:", err)
        return err
    }
    cmd.Println("Operation successful")
    return nil
}
```

## Notes
- The project uses Go's `flag` package but includes `cobra-cli` in dev environment
- When adding new commands, prefer Cobra patterns
- Follow existing patterns in codebase where Cobra is gradually adopted