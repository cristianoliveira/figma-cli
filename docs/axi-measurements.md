# AXI measurements

Repeat from repository root:

```bash
scripts/measure-axi.sh
```

The harness builds both executables, uses committed PNG fixtures, captures stdout
and stderr separately, and reports exit status, command count, and expected
recovery round trips. It performs no network calls. The discovery row includes
resolved executable path, so its byte count varies with checkout path length.

Measured 2026-04-13:

| scenario | stdout bytes | stderr bytes | exit | commands | recovery round trips |
|---|---:|---:|---:|---:|---:|
| figma-discovery | 224 | 0 | 0 | 1 | 0 |
| figma-invalid-toon | 157 | 0 | 2 | 1 | 1 |
| figma-invalid-json | 194 | 0 | 2 | 1 | 1 |
| pixel-identical-toon | 374 | 0 | 0 | 1 | 0 |
| pixel-identical-json | 484 | 0 | 0 | 1 | 0 |
| pixel-failed-gate-toon | 1,455 | 78 | 1 | 1 | 0 |
| pixel-failed-gate-json | 2,028 | 78 | 1 | 1 | 0 |
| pixel-probe-truncated-csv | 905 | 0 | 0 | 1 | 0 |
| pixel-probe-truncated-json | 11,349 | 0 | 0 | 1 | 0 |
| pixel-scan-csv | 133 | 0 | 0 | 1 | 0 |
| pixel-scan-json | 628 | 0 | 0 | 1 | 0 |

## Decisions supported by evidence

- Keep TOON as structured default for comparison results and errors in these
  measured shapes; JSON remains explicit interoperability output. This is not a
  fixed savings claim—TOON can be larger for other shapes.
- Keep probe and scan CSV defaults. Probe JSON carries nested RGBA and delta
  detail and is much larger for same bounded selection.
- Keep failed-gate result plus stderr diagnosis: agent receives evidence in one
  command and needs no recovery query.
- Keep structured usage recovery: invalid command takes one follow-up rather
  than a help dump in initial response.
- No new `--fields` or detail command is justified by these measurements. Existing
  bounds and conditional hints remove flood/recovery problems without another API.
