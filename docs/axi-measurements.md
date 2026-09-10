# AXI measurements

Repeat from repository root:

```bash
scripts/measure-axi.sh
```

The harness builds the executable, captures stdout and stderr separately, and
reports exit status, command count, and expected recovery round trips. It performs
no network calls. The discovery row includes the resolved executable path, so its
byte count varies with checkout path length.

Measured 2026-04-13:

| scenario | stdout bytes | stderr bytes | exit | commands | recovery round trips |
|---|---:|---:|---:|---:|---:|
| figma-discovery | 224 | 0 | 0 | 1 | 0 |
| figma-invalid-toon | 157 | 0 | 2 | 1 | 1 |
| figma-invalid-json | 194 | 0 | 2 | 1 | 1 |

## Decisions supported by evidence

- Keep structured usage recovery: invalid command takes one follow-up rather
  than a help dump in initial response.
- No new `--fields` or detail command is justified by these measurements. Existing
  bounds and conditional hints remove flood/recovery problems without another API.
