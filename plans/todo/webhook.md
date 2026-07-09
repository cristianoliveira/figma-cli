# webhook (CI/automation bridge)

## Problem

Design files change. Code needs to react. Today, the trigger mechanism is human:

"Hey @dev, I updated the Figma file" — Slack message, seen hours later.

This means:
- **CI pipelines don't start** when the design changes — they start when someone remembers
- **Screenshot tests don't auto-run** against the latest design
- **Storybook builds don't regenerate** when tokens/components update
- **No audit trail** of "what triggered this build?"

Figma webhooks exist but require an HTTP server to receive them. No one sets them up
because "setting up a webhook receiver" is infrastructure work, not design work.

## Success criteria

- Create/list/delete Figma webhooks from the CLI (no server needed for CRUD)
- A companion `figma webhook listen` command that runs a lightweight HTTP server
  to receive events locally (for dev/debug)
- A companion `figma webhook forward` to relay events to a CI system (GitHub Actions,
  CircleCI, etc.)
- Event types: file_update, file_version_update, file_comment, library_publish,
  file_delete, dev_mode_status_update

## API mapping

| Action | API endpoint |
|---|---|
| Create webhook | `POST /v2/webhooks` |
| List webhooks | `GET /v2/webhooks` (by team or by endpoint) |
| Get webhook | `GET /v2/webhooks/{webhook_id}` |
| Update webhook | `PUT /v2/webhooks/{webhook_id}` |
| Delete webhook | `DELETE /v2/webhooks/{webhook_id}` |
| View webhook requests | `GET /v2/webhooks/{webhook_id}/requests` |
| Event types | `FILE_UPDATE`, `FILE_VERSION_UPDATE`, `FILE_COMMENT`, `LIBRARY_PUBLISH`, `FILE_DELETE` |

## CLI shape

```
figma webhook create <file-url> [flags]
figma webhook list [--file <file-url>]
figma webhook delete <webhook-id>
figma webhook test <webhook-id>
figma webhook listen [flags]           # Local server for debugging
figma webhook forward [flags]          # Forward to CI
```

### Create
```
figma webhook create <file-url> \
  --name "design-update" \
  --event file_update \
  --endpoint "https://ci.example.com/figma-hook" \
  --secret "$WEBHOOK_SECRET"
```

### Listen (local debug)
```
figma webhook listen --port 3000
# → Listening on :3000
# → Waiting for Figma webhook events...
# ← [file_update] Drive (Cells) updated by Olga Skoczylas
```

## Output example

```json
{
  "webhooks": [
    {
      "id": "wh_abc123",
      "name": "design-update",
      "endpoint": "https://ci.example.com/figma-hook",
      "event_type": "FILE_UPDATE",
      "file_key": "grnVU2vAihHXwYgHryu2xE",
      "status": "ACTIVE",
      "created_at": "2026-07-07T10:00:00Z"
    }
  ]
}
```

## CI integration patterns

### GitHub Actions
```yaml
# .github/workflows/figma-update.yml
on:
  repository_dispatch:
    types: [figma-update]
jobs:
  update:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: figma tokens --format css --team wire --output src/styles/tokens.css
      - run: git diff --exit-code || (echo "Tokens updated" && exit 1)
```

### General pattern
```
Figma file update
  → Figma webhook fires
  → CI endpoint receives it
  → CI runs: figma tokens / figma icons / figma components --diff
  → Artifacts (tokens, icons, diff report) are generated
  → PR is created if changes detected
```

## Edge cases

- Webhook requires a Team or Enterprise plan → error message should be clear
- Duplicate webhook names → allow, warn
- Endpoint unreachable → Figma retries, CLI shows last delivery status
- Secret rotation → `--secret` flag overwrites existing
- Local testing without a public URL → `listen` mode + ngrok suggestion
