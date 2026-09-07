# Pixel Perfect implementation evals

Evaluate whether the skill helps an agent deliver a working component from a
PNG—not whether it can recite CLI flags. No Figma token or live upload service
is needed. Cases and expectations are in `evals.json` (skill-creator format).

## Cases and success criteria

1. **Reusable upload panel:** build real UI from the PNG; implement accessible
   local cancel/retry/collapse actions; demonstrate measured visual improvement.
2. **App integration:** add the same component to the existing workspace;
   preserve task/attachment behavior and prove both isolated and integrated use.
3. **Larger capture, tight budget:** keep 900 × 700 screenshots; compare the
   component using explicit `--actual-crop 32,24,436,406`; stop after at most two
   refinement iterations and report remaining differences honestly.

Each case gets the same source archive, reference PNG, and fixed data. Do not
pre-implement the target component in the fixture. See
[fixture instructions](../fixtures/README.md) for dependencies and repacking.

## Offline validation

From the repository root:

```sh
python3 -B -m unittest discover -s skills/pixel-perfect/evals -p 'test_*.py' -v
```

These checks validate packaging, fixture integrity, schema essentials, and crop
geometry. They are **not** model quality scores. Verify an extracted app with
`npm ci --ignore-scripts`, `npm run test:coverage`, and `npm run build` before
spending model calls.

## Run a comparison after scope approval

Agree on model/tier, cases, repetitions, concurrency, timeouts, and iteration
limits first. Suggested smoke: **case 1, one candidate/baseline pair, sequential,
900 seconds per run, five refinement iterations**. This is two model executions;
grading calls, if used, are additional. A smoke run does not establish reliability.
Then consider all three cases with three repetitions per side (18 executions).

Use the installed `skill-creator` runner rather than adding another eval engine.
Snapshot the complete original skill outside the candidate before edits. For a
legacy snapshot containing installed dependencies, prepare a separate baseline
execution copy containing its unchanged `SKILL.md` and required references/scripts
only; supply the exact same fixtures through `--file` to both configurations.
Record that preparation in eval metadata. Never send `node_modules` or prior
session transcripts as skill content.

Example setup (absolute paths; choose a fresh iteration directory):

```sh
export SKILL_ROOT="$PWD/skills/pixel-perfect"
export CREATOR="/absolute/path/to/skill-creator"
export BASELINE_SKILL="/absolute/path/to/pre-change-execution-copy"
export CASE_ROOT="/absolute/path/to/pixel-perfect-workspace/iteration-1/eval-1-upload-panel"
python3 - <<'PY'
import json, os
from pathlib import Path
root = Path(os.environ['CASE_ROOT'])
root.mkdir(parents=True, exist_ok=False)
case = json.loads((Path(os.environ['SKILL_ROOT']) / 'evals/evals.json').read_text())['evals'][0]
(root / 'prompt.txt').write_text(case['prompt'])
(root / 'eval_metadata.json').write_text(json.dumps({
    'eval_id': case['id'], 'eval_name': 'upload-panel',
    'prompt': case['prompt'], 'assertions': case['expectations'],
    'baseline_skill': os.environ['BASELINE_SKILL'],
}, indent=2) + '\n')
PY

cd "$CREATOR"
python3 -m scripts.run_case \
  --prompt-file "$CASE_ROOT/prompt.txt" \
  --skill-path "$SKILL_ROOT" \
  --file "$SKILL_ROOT/fixtures/todoapp.tar.gz" \
  --file "$SKILL_ROOT/fixtures/upload-modal-multifiles.png" \
  --file "$SKILL_ROOT/fixtures/upload-state.json" \
  --run-dir "$CASE_ROOT/with_skill/run-1" --tier high --timeout 900

python3 -m scripts.run_case \
  --prompt-file "$CASE_ROOT/prompt.txt" \
  --skill-path "$BASELINE_SKILL" \
  --file "$SKILL_ROOT/fixtures/todoapp.tar.gz" \
  --file "$SKILL_ROOT/fixtures/upload-modal-multifiles.png" \
  --file "$SKILL_ROOT/fixtures/upload-state.json" \
  --run-dir "$CASE_ROOT/without_skill/run-1" --tier high --timeout 900
```

`--tier high` is illustrative; use the agreed, configured tier identically on
both sides. The `without_skill` directory holds the old-skill baseline here,
not a no-skill run. The runner copies inputs under `inputs/` and asks for
outputs under `outputs/`. It isolates context, **not the host filesystem or
network**. Use disposable workspaces, loopback servers, and no production data.
Stop browsers/dev servers after each run; never reuse a run directory.

## Grade evidence, not assertions in the final message

Follow skill-creator's `agents/grader.md`. Save `grading.json` beside each run's
`timing.json` with expectation `text`, `passed`, and specific `evidence`, plus
summary counts and pass rate. Review these gates:

- Run delivered source/build/tests; inspect controls in a browser. Test-file
  presence alone does not prove behavior. Check the original tests were retained.
- Check screenshots against the actual reference and source. Require a real DOM
  implementation, not the PNG in an `<img>`, canvas, or traced substitute.
- Re-run recorded comparisons from retained images with the same CLI version and
  options. Compare emitted JSON to reported values. A mask filename is not proof.
- Inspect transcript/capture commands for baseline timing, fixed viewport/device
  scale, stable state, unchanged thresholds/crops, and bounded iteration count.
- Assess visual resemblance and remaining differences by human review. Metric
  improvement alone can reward a poor first render; it is not an absolute quality
  gate. Do not invent an uncalibrated visual pass threshold after seeing results.
- Count runtime/provider/tool setup failures as infrastructure errors, not
  successful negative cases. Absence of metrics is not a verified match.

Use independent grading where practical; disclose inline-grading limitations.
Keep expectations fixed for both sides. If an assertion is flawed, revise it
explicitly and regrade both. Report pass rates **and** time/tokens/variation.

From skill-creator, aggregate and generate its existing review viewer:

```sh
python3 -m scripts.aggregate_benchmark "$(dirname "$CASE_ROOT")" --skill-name pixel-perfect
python3 eval-viewer/generate_review.py "$(dirname "$CASE_ROOT")" \
  --skill-name pixel-perfect \
  --benchmark "$(dirname "$CASE_ROOT")/benchmark.json" \
  --static "$(dirname "$CASE_ROOT")/review.html"
```

Keep screenshots, agent outputs, reports, and benchmark workspaces outside the
skill and untracked. Current scope is content execution; these prompts do not
measure whether an agent chooses the skill from the full installed catalog.
