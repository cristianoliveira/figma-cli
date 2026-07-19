# Add structured, redacted error output

## Problem

Usage and operational failures currently produce empty stdout plus plain text stderr. Some operational errors expose raw filesystem paths or dependency wording. This is recoverable for humans but not strict AXI.

## Proposed outcome

Create one typed error contract with category, concise message, offending input when safe, exit code, and at most one actionable recovery step. Render it in the selected structured format according to task 01. Reserve stderr for progress and debug diagnostics.

## Pre-analysis

- `cli.UsageError` and `cli.ExitCode` already distinguish usage and operational failures.
- flag correction exists in `internal/cli/flags.go`.
- entrypoints currently prepend `error:` and print to stderr.
- commands return a mix of domain errors, wrapped filesystem errors, HTTP errors, and silent `ExitCodeError` values.
- comparison gate failures intentionally emit result JSON before returning exit 1 and must remain a result, not become a generic error envelope.

## Project approach

1. Add process-level failing tests for unknown command, unknown flag, missing argument, missing token, missing file, invalid image, HTTP authorization, rate limit, provider failure, and silent grep-style exits.
2. Define typed usage and operational errors without parsing arbitrary strings.
3. Translate dependency failures at package boundaries into user-facing domain language.
4. Add one safe recovery action when known; avoid generic help dumps.
5. Redact tokens, headers, provider responses, stack traces, and unnecessary absolute paths.
6. Preserve validation-gate result output plus exit 1.
7. Add explicit debug mode for raw diagnostics if needed; debug output stays on stderr.

## Acceptance criteria

- structured errors are deterministic and valid in every supported structured format.
- usage failures exit 2; operational failures exit 1; success/no-op exits 0.
- unknown flags retain nearest valid suggestion and local help command.
- dependency failures contain domain language and one actionable recovery step.
- secrets and raw dependency output are absent from default output.
- validation gates and intentional silent exits preserve documented semantics.
- tests assert stdout, stderr, and exit status on built binaries.

## Risks

- string-based translation is brittle; use typed causes.
- file paths can be useful evidence; redact only details not required for recovery.
- dual result-plus-error output can confuse consumers unless gate semantics stay distinct.

## Implementation freedom

Choose envelope field names from the contract decision. Do not expose internal Go error type names.
