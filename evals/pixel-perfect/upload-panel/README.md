# Upload-panel calibration controls

Keep the accepted answer and evaluator checks here, **outside `skills/`**. The
skill-creator runner copies the entire skill into the implementing agent's run;
putting the answer there would leak the solution. These files are grader inputs,
not task inputs. Never pass this folder through `--skill-path` or `--file`.

Context isolation is not a filesystem sandbox. For a true holdout, run the
implementing agent in an environment that cannot access this directory. Merely
keeping it outside the copied skill does not prevent host filesystem access.

## Accepted reference

`accepted/` contains the source-only app archive, reference PNG, accepted browser
screenshot, original comparison metrics/mask/overlay, capture settings, and
checksummed provenance. Cristian accepted this result with **“Looks good”**.
The original app tests and six component tests remain inside the source archive.
There are no installed dependencies, transcripts, credentials, or browser profiles.

This is a known acceptable implementation, **not the only correct source code**.
Do not replace the blank starter app in `skills/pixel-perfect/fixtures/` with it.
Treat this snapshot as immutable; record a separately named example if accepting
a new version. The original metric artifact paths are historical; `mask.png` and
`overlay.png` are the corresponding durable files here.

## Four one-defect controls

`variants.json` describes exact source replacements. Each variant starts from
the accepted archive, never from another mutated app. Full copies are generated
only in disposable workspaces, avoiding duplicate archives in Git.

| Control | Deliberate defect | Content check | Retry check | Initial image differs from accepted |
| --- | --- | --- | --- | --- |
| Accepted | None | Pass | Pass | No |
| Row spacing | Rows 40px instead of 46px | Pass | Pass | Yes |
| Missing row | Omit completed video item | Fail | Pass | Yes |
| Status colors | Completed text purple instead of green | Pass | Pass | Yes |
| Broken Retry | Retry leaves failed item unchanged | Pass | Fail | **No** |

The checker does not receive the variant's name or expected outcome. It checks
all seven visible names/states, exercises Retry, and captures real DOM. Only the
runner compares observations to the declared expected outcome. A negative
control being detected is a **successful evaluator check**, not a good component.

## Implementation alternatives

`alternatives.json` adds two candidates without changing the accepted snapshot:

- **Grid rows:** use CSS Grid instead of Flexbox. In the measured capture, it is
  pixel-identical at threshold 0 despite different source code.
- **Retry icon:** use a thinner 1.5px stroke. It changes 113 pixels at threshold 8
  while keeping the control name, hit area, content, and Retry behavior.

Both pass all 22 app tests, coverage gates, build, and the browser content/Retry
checks. These are small edits to one implementation, not independently built apps.
The **thinner Retry icon is accepted**: Cristian answered “Yes” to “Would you
accept this thinner icon too?”. Exact feedback, reviewed screenshot, and source
hashes are preserved in `reviews/retry-icon.json` and `reviews/retry-icon.png`.
Grid remains pending separate human review. Passing technical checks alone does
not change `review_status` to accepted.

Run them alongside the original controls:

```sh
python3 -B evals/pixel-perfect/upload-panel/verify.py \
  "$PWD/.tmp/pixel-perfect-controls/alternatives-NEW" --include-alternatives
```

Unlike deliberate defects, alternatives do not prescribe `visual_difference`
in their expectations. The runner still records that measurement, but does not
reject a candidate just because some pixels differ. Content and Retry must pass.
Read the generated `retry-icon/accepted-delta-report.html` to review the icon
against the original accepted render. The recorded approval covers only the
thinner Retry icon, not all icon changes or a universal pixel tolerance.

`alternative-observations.json` stores the measured results and screenshot/checker
hashes. The older `observations.json` is retained as a historical measurement of
the original controls; its checker hashes describe the earlier code, not the
current runner. The observations also retain their original pre-review pending
statuses; `alternatives.json` and linked review records show current decisions.
Neither observation file is an automatically calibrated tolerance.

## Coordinator-owned A/B verification

`ab_verify.py` grades one agent run without trusting its screenshots, claims, or
metrics. It requires the app directly at `outputs/app`, installs locked
dependencies, runs tests and a production build, and uses the fixed Chromium
profile in `accepted/capture.json`. `ab_browser_check.js` then takes raw
transparent `initial.png` and `final.png` screenshots from fresh seven-item
source state, tests Cancel/Cancel all/Retry separately, and checks keyboard
collapse/expand. A threshold-zero comparison must reproduce the two raw source
captures before reference metrics are recorded.

Run it after an agent run, from a new coordinator workspace:

```sh
python3 -B evals/pixel-perfect/upload-panel/ab_verify.py \
  /absolute/run-directory /absolute/new-evidence-directory --port 5193
```

The coordinator, not the candidate, produces `reference-metrics.json`, mask,
overlay, report, screenshots, source hashes, and `result.json`. Metrics remain
review evidence, not an automated visual-quality threshold.

## Mid-tier bundle pilot

`mid-bundle-pilot.json` records one GPT-5.6 Luna/high pair: skill + CLI versus
neither, with the same neutral prompt and a 900-second cap. **No visual winner
was established.** Baseline completed but postprocessed its screenshots and used
a five-item final scene. Assisted timed out after browser setup detours and
measurement work; its unfinished test file fails the production type check.
Both snapshots retain the original tests and pass 18 tests plus coverage gates.

Before another pair, pin the tested browser and explicitly require raw captures
of the same seven-item scene. The record retains protocol, usage, artifact hashes,
and local evidence paths. It is a pilot failure record, not a skill benchmark win.

## Revised mid-tier bundle result

`mid-bundle-v2.json` records one completed GPT-5.6 Luna/high pair. Both
conditions passed independent build, tests, fixed seven-row content, Cancel /
Cancel all / Retry, keyboard collapse/expand, and raw source-capture
reproduction. Skill + CLI reduced coordinator-measured changed ratio from
23.2747% to 17.8888% and perceptual RMSE from 0.14251 to 0.13117, but used
2.23× wall time and 2.69× reported total tokens.

The completed blind review selected **Condition B** as closer; unblinded, B is
**neither skill nor CLI**. Exact feedback, “This looks closer”, is preserved in
`mid-bundle-v2-feedback-blind.json`. It does not accept either render. This
human result conflicts with full-frame metrics, so this pair establishes **no
visual-quality win** for skill + CLI. Keep human review; diagnose regional and
alpha/shadow metric weighting before another model pair.

## Verify offline integrity

From repository root (Python 3.10+):

```sh
python3 -B -m unittest discover -s evals/pixel-perfect/upload-panel -p 'test_*.py' -v
```

Tests cover provenance, archive paths/links/size limits, exact single-file
mutations, missing/ambiguous replacements, starter separation, and runner policy.
Runner-policy tests mock browser/process execution and do not prove UI behavior.

## Run real controls

Requires Node.js 22.12+, npm 11, installed `pixel-perfect`, and `playwright-cli`
with Chromium. npm registry access is needed for the first locked install;
there are **no model calls, Figma requests, or actual uploads**.

```sh
python3 -B evals/pixel-perfect/upload-panel/verify.py \
  "$PWD/.tmp/pixel-perfect-controls/run-NEW" --port 5191
```

Use a new workspace each time. The runner:

1. Verifies accepted checksums and safely materializes five source trees (seven
   with `--include-alternatives`).
2. Installs locked dependencies once with lifecycle scripts disabled; variants
   share that install through local symlinks. Only trusted accepted code runs.
3. Runs accepted app tests/coverage and, when selected, both alternatives' tests
   and coverage. Builds **every** control so syntax or build failures cannot
   masquerade as detected defects.
4. Uses fresh, short-named browser sessions and loopback-only servers. Captures
   436 × 406 at DPR 1, waits for fonts, and disables animations/carets.
5. Requires the current accepted render to reproduce the saved screenshot at
   threshold 0 before interpreting controls. Browser/OS/system-font drift blocks
   calibration; it is not a failed component or permission to update the golden.
6. Applies the same content/Retry checks to all controls. Compares each image
   both with the design PNG and accepted render at threshold 8, without resizing,
   cropping, or ignored regions. Saves screenshots, overlays, reports, and JSON.
7. Closes only its own browser/server and saves `results.json`.

Exit codes: **0** all controls behaved as declared; **1** at least one defect
escaped its expected checks; **2** infrastructure error. Inspect logs, not just
an exit code. Browser crashes and malformed results never count as passing
negative controls. Failed workspace artifacts remain available for diagnosis.

To inspect one source variant without running a browser:

```sh
python3 -B evals/pixel-perfect/upload-panel/controls.py \
  broken-retry "$PWD/.tmp/broken-retry-app"
```

The original component regression test also catches broken Retry. After installing
dependencies, run this inside the accepted or broken app directory:

```sh
npm test -- src/UploadStatus.test.tsx -t 'Cancel all preserves'
```

Accepted passes. Broken Retry produces a failed assertion (not a setup error).
The browser runner already checks the same transition independently.

## What this establishes—and does not

`observations.json` records one real run, not expected numbers invented by tests.
The accepted render reproduced exactly; all four intentional defects were
caught. Broken Retry has **zero visual difference**, proving screenshots cannot
replace behavior checks. Wrong status color changes the whole-image ratio only
slightly, despite being a deliberate color defect.

`visual_difference` means **different from the accepted control**, not
**unacceptable to a human**. Exact reproduction is an environment guard for this
mutation experiment, not a demand that future implementations use identical code
or pixels. These checks establish sensitivity to four known defects; they do not
establish specificity against other acceptable implementations.

Do not turn 13.36% (the accepted design difference) into a universal cutoff. Next:
review the two alternatives, then gather more acceptable implementations and
small defects before proposing and validating region-specific gates. Human review remains the visual acceptance gate
until those limits are calibrated. Behavior/content checks remain independent.
