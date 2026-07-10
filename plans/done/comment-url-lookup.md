# Deterministic comment URL lookup

## Problem

A copied Figma comment URL contains both node scope and comment identity:

```text
https://www.figma.com/design/QAhpkgySSOJ6gwJUTB0glb?node-id=4707-15501&m=dev#1838610593
```

`figma comments` currently parses the file and node but ignores the fragment. It therefore returns every comment in the scoped subtree instead of comment `1838610593`. When developer checks a nearby child node, the result can be empty even though relevant unresolved comment is attached to an ancestor frame. Comment objects also lack direct URLs back to Figma.

## Solution

Make copied Figma comment URLs deterministic entry points:

1. Parse numeric URL fragment as comment ID and return that exact comment.
2. Add `--unresolved-only` filtering.
3. Add explicit `--include-ancestors` scope expansion.
4. Include canonical Figma URL in every comment output.
5. Report matched scope/ancestor rather than silently broadening results.

Exact fragment lookup should take precedence over node filtering because comment ID is the strongest identifier. If ID is absent from API response, return a clear not-found error.

## How

- Extend parsed Figma input with optional `CommentID` extracted from numeric URL fragment; reject or ignore non-comment fragments deliberately and test both cases.
- Move comment mapping/filtering from Cobra closure into pure extractor/service helpers.
- Filter by exact comment ID before applying broader node scope; define whether replies are returned separately or only exact match.
- Add `--unresolved-only` using `ResolvedAt == nil`.
- Add `--include-ancestors`. Fetch full file document, locate selected node, derive ancestor ID path, then include comments anchored to those IDs. Do not guess ancestry from node ID syntax.
- Add canonical URL builder using file ID, comment node ID, and comment ID. Preserve stable query format; URL need not preserve irrelevant original query parameters.
- Return scope metadata when ancestor search finds comment, for example `matchedNodeId` and `scope: "ancestor"`.
- Add tests for exact hash match, unresolved filtering, child with ancestor comment, missing comment ID, reply comments, and canonical URL generation.

## Ideal output

```json
{
  "id": "1838610593",
  "node_id": "4707:15501",
  "resolved": false,
  "user": "Astrid Pahl",
  "message": "@Cristian Oliveira https://support.wire.com/hc/en-us/articles/37518608388125-Prevent-adminless-groups",
  "url": "https://www.figma.com/design/QAhpkgySSOJ6gwJUTB0glb?node-id=4707-15501#1838610593"
}
```

## Success criteria

- Pasting a Figma comment URL returns one deterministic comment.
- `--unresolved-only` excludes resolved comments.
- Ancestor lookup is explicit and reports where match was found.
- Every comment can be opened directly in Figma.
- Happy and unhappy paths require no live Figma token in tests.
