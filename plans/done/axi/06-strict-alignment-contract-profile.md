# Choose the AXI compatibility profile

## Decision

Adopt the **AXI-first v2 profile**: TOON is the default for every structured stdout result, including errors. JSON is an explicit interoperability format, not the default.

This initiative is a breaking command-contract migration. Existing scripts must opt into JSON; preserving JSON as the default would leave the principal AXI goal incomplete.

## Chosen profile

| Concern | v2 contract |
|---|---|
| Structured success output | Default stdout is TOON. Existing global `--json` provides explicit JSON compatibility. Intentional CSV output remains CSV where tabular output is the command contract. |
| Text and file output | Preserve deterministic text/file output. Do not encode artifacts as TOON. |
| Errors | Usage and operational errors emit a structured, redacted TOON envelope on stdout. Stderr is reserved for progress and diagnostics. |
| Exit codes | `0` success/no-op, `1` operational failure, `2` usage error. |
| Discovery | No-argument views identify the resolved executable path and return compact useful state in the selected structured format. |
| Large output | Potentially large views are bounded by default with exact totals, explicit truncation, and a copyable recovery hint. |

## Migration rules

1. Select and verify a maintained, specification-compatible TOON library before adding rendering. Never hand-roll TOON.
2. Build one shared output boundary: commands pass format-independent domain values to it.
3. Make TOON the default at that boundary. Retain global `--json` as explicit compatibility selection.
4. Keep command-local `--format` controls for intentional artifacts and views such as CSV, CSS, image exports, and recursive inspect text; they do not select the global structured encoding.
5. Preserve CSV for commands that intentionally offer tabular output; it is not converted to TOON.
6. Render structured usage and operational errors through the same boundary before returning the existing exit code. Redact secrets, provider details, and unnecessary absolute paths.
7. Validate flags before loading credentials, reading files, decoding images, or calling providers.
8. Release as v2 with a migration guide containing copyable JSON opt-in commands, stdout/stderr examples, and an output compatibility matrix.

## Compatibility policy

- v2's default TOON contract is intentionally breaking.
- Global `--json` is the JSON compatibility path for structured data and existing text/file envelopes.
- Field removal, rename, type change, semantic reuse, format-default change, or error-channel change requires another explicit major migration.
- Additive fields are compatible only when their absence has documented meaning.
- `--json` remains supported; JSON is no longer the implicit default.

## Consequences for remaining work

- **TOON migration** creates the shared boundary, changes defaults, and proves TOON/JSON semantic round-trips.
- **Structured errors** uses that boundary for deterministic redacted error envelopes on stdout.
- **Bounds and hints** returns exact totals, returned counts, and recovery commands in TOON and JSON.
- **Discovery and measurement** verifies executable identification and compares v2 TOON against explicit JSON for bytes and recovery round trips.

## Acceptance evidence

The decision is complete when later implementation tests enforce:

- structured success and failure output defaults to parseable TOON;
- `--json` produces semantically equivalent JSON domain values;
- CSV commands retain their explicit CSV behavior;
- stdout contains only result/error envelopes and stderr only diagnostics;
- exit statuses remain `0`, `1`, and `2` by contract;
- v2 release notes document the removed `--json` switch and JSON replacement commands.
