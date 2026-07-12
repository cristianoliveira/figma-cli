# AGENTS.md — qmd (Librarian CLI) Brief

## Overview

This project uses `qmd` for indexing and searching documentation and agent reports.
Agents should use `qmd` to discover relevant documentation, previous agent work, and research findings.

## Available Collections

- **figma-cli** - Project documentation (28 files)
- **research** - Research documents, agent reports, temporary files (indexes `.tmp/reports/`, `.tmp/researches/`, and `research/`)

## Core Commands

### Keyword search (BM25)
```bash
qmd search "query" -c figma-cli
```

### Semantic search
```bash
qmd vsearch "query" -c figma-cli
```

### Hybrid search (keyword + semantic + LLM reranking)
```bash
qmd query "query" -c figma-cli
```

### List documents in collection
```bash
qmd list -c figma-cli
```

### Get a document
```bash
qmd get <file>[:line] -l N
```

### List collections
```bash
qmd collection list
```

## Recommended Flow

1. Search (`qmd search` or `qmd query`) for relevant documentation
2. Select the most relevant result
3. Read the document using `qmd get` or open the file
4. Use `qmd list` to browse collection contents
5. Stop once enough context is found

## Rules

- Use `qmd` to search for documentation instead of scanning filesystem directly
- Write agent reports to `.tmp/reports/` (indexed in `research` collection)
- Write research notes to `.tmp/researches/` for future reference

## Summary

Use `qmd` to discover documentation, read focused content, and find previous agent work.
Search → Read → Stop Early