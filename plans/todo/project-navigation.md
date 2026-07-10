# Frontend project navigation

## Problem

`projects` and `files` mirror Figma API boundaries but do not support common frontend starting point: pasted design/file URL from Jira. Developer cannot easily answer “which project contains this file?” or navigate to sibling files and team context.

## Solution

Provide URL-first navigation from current file to available project/team context, while keeping direct team/project ID commands for explicit exploration.

## How

- Verify which parent/team/project metadata Figma API exposes for file endpoints and document limitations clearly.
- Add discoverable composition path such as `meta --context` or dedicated context command rather than hidden API calls.
- Return file metadata, containing project/team when available, and next commands for projects/files.
- Preserve direct `projects <team>` and `files <project>` workflows.
- Cache or avoid repeated file fetches within composed command.
- Add tests for file URL, unavailable parent metadata, permission errors, and direct IDs.

## Success criteria

- A Jira-provided Figma URL is sufficient starting point for discovery.
- Missing parent metadata is explicit, not guessed.
- Existing direct-ID exploration remains available.
