# Component implementation fixtures

These are inputs, not a completed solution. The app does not contain the upload
status panel; the evaluated agent must implement it.

| File | Purpose |
| --- | --- |
| `todoapp.tar.gz` | Source-only React/TypeScript/Vite app, including original tests and npm lockfile. |
| `upload-modal-multifiles.png` | Original, unmodified 436 × 406 reference, including shadow padding. |
| `reference.sha256` | Reference checksum; verify before and after an eval. |
| `upload-state.json` | Fixed seven-row demo state: two uploading, one failed, four completed. |

The reference's Shared Drive label is visual content, not permission to build a
remote upload integration. Use explicit local state transitions. Do not advance
progress on timers or depend on the todo app's date-relative seed data for panel
captures. The PNG is the visual source of truth; JSON supplies content and states,
not a hidden pixel-perfect implementation. Exact source font metadata is not
available, so remaining typography differences must be reported, not dismissed.

## Unpack and verify

Use a fresh disposable directory per run. Requires Python 3.10+ for fixture tools,
Node.js 22.12+ and npm 11 for the app; browser capture and `pixel-perfect` are
required for model evals.

```sh
mkdir -p /tmp/pixel-perfect-example
# Run from this fixtures directory; use a new destination for every real eval.
tar -xzf todoapp.tar.gz -C /tmp/pixel-perfect-example
cd /tmp/pixel-perfect-example/todoapp
npm ci --ignore-scripts
npm test
npm run build
npm run dev -- --host 127.0.0.1 --port 5173 --strictPort
```

The archive preserves the original app README and tests. Its normal homepage
uses browser storage and date-relative sample tasks. For captures, create an
isolated preview of the actual new component and supply `upload-state.json`.
Use a fresh browser context per run and a unique session/port if running pairs
concurrently. Do not expose fixture servers beyond loopback.

## Update the archive

Edit an unpacked copy, run its tests/build, then from the repository root:

```sh
python3 skills/pixel-perfect/evals/pack_fixture.py \
  /absolute/unpacked/todoapp skills/pixel-perfect/fixtures/todoapp.tar.gz
python3 -B -m unittest discover -s skills/pixel-perfect/evals -p 'test_*.py' -v
```

The packer uses a source/config/asset allowlist. It keeps `.npmrc`, `.gitignore`,
the lockfile, and tests, but excludes `node_modules`, build/coverage output,
TypeScript caches, browser artifacts, and agent histories. It rejects symlinks
and normalizes archive metadata; identical source produces identical bytes on
the same Python/zlib toolchain. Review the allowlist if adding a new source or
asset type. Do not keep a second expanded app in this skill folder: the eval
runner copies the whole skill into every candidate run.

A source-only archive reduces fixture copies and repository size. It does not
remove the need to install dependencies in disposable runs, nor reduce the
amount of source an agent must read to implement the component.
