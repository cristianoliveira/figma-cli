---
id: TASK-0012
title: Publish release artifacts
status: doing
depends_on: []
priority: high
tags: []
---

# Publish release artifacts

## Problem
The repository has no GitHub Actions workflow, so tagged releases do not produce downloadable, checksummed Figma CLI binaries.

## Context
Create one GitHub Actions release workflow. A pushed semantic-version tag triggers a reproducible Go build matrix for Linux, macOS, and Windows on amd64 and arm64. Publish archives and a SHA-256 checksum manifest to the matching GitHub Release. Keep pull-request validation separate from publishing and use least-privilege workflow permissions.

## Acceptance criteria
- [x] A tag matching `v*` triggers the workflow; pull requests cannot publish a release.
- [x] The workflow builds `figma` for linux/darwin/windows and amd64/arm64 using the repository Go module.
- [x] Each artifact has an unambiguous OS/architecture filename; Windows artifacts are executable archives.
- [x] A deterministic SHA-256 checksum manifest covers every published archive.
- [x] The workflow creates or updates the GitHub Release for the tag and uploads only its generated artifacts and checksum manifest.
- [x] Workflow permissions are minimal and release publishing is explicit.
- [x] A local reproducible artifact-build command or documented verification procedure exists.
- [x] The workflow YAML is validated and its shell steps fail safely.

## Notes
Do not attach untracked local artifacts or change CLI runtime behavior. Keep the release workflow focused on tagged publication.
