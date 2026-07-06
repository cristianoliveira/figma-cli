# Agent Instructions

## References

 - Local docs in ./docs/**/*.md
 - Use librarian to check `qmd help`

## CLI Design Preferences

- Prefer clean noun commands with options over positional subcommands for filters.
- For extracting text from a named Figma layer, use this shape:
  `figma texts --layer "Rectangle 3 Copy 15" <figma-file-url-or-id>`
- Layer names may not be unique; return all matches with node IDs and extracted text.
- Reuse existing URL parsing and text traversal logic before adding new APIs.
