# Figma CLI Design Document

## Overview

This document outlines the design and command structure for `figma-cli`, a command-line interface tool for exploring and interacting with Figma designs. The CLI is designed to be LLM-friendly, enabling AI agents to easily fetch, analyze, and understand Figma design data.

## Design Goals

1. **LLM-Friendly**: Commands should be easily parseable by AI agents with JSON output
2. **URL-Based**: Accept Figma URLs directly without manual ID extraction
3. **Context-Aware**: Always provide node hierarchy and relationships
4. **Efficient**: Minimize API calls through batch operations and caching
5. **Developer-Centric**: Support common workflows like design inspection, code generation, and asset export

## Driving Use Cases

### Use Case 1: Design Review - Text Changes
**User Request:**
> "Astrid noticed one small thing - title of Global files list should have been updated to 'Files' https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221&t=FJHwt3uytkvpEqRd-11"

**Agent Needs:**
- Extract file key and node ID from URL
- Fetch current text content
- Understand context (what is this node, where is it located)
- Verify what needs to change

**Required Commands:**
- URL parsing and node retrieval
- Text extraction
- Context/hierarchy inspection

### Use Case 2: Design Requirements - Complex Component Analysis
**User Request:**
> "When user starts a new conversation (Group and Channel) with Cells on, display system messages to user with information regarding:
> 1. Lack of self-deleting messages
> 2. Cells being on
> 3. Update encryption message for Wire Cells enabled conversation only - green banner (no link updates at this stage).
> Design: https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Cells---Files?node-id=7450-15847&t=7K5F93ii9XbzeCbh-4"

**Agent Needs:**
- Access specific design node and understand its structure
- Read multiple text components within a container
- Extract design specifications (colors, banners, styling)
- Find related system messages across the file
- Identify component hierarchy and variants
- Check existing implementations vs. requirements

**Required Commands:**
- Node retrieval with children
- Text content extraction (recursive)
- Design spec extraction (colors, typography)
- Search functionality to find related content
- Component inspection

## Command Taxonomy

### 1. Core Navigation & Retrieval
Commands for fetching and accessing Figma data

### 2. Structure & Context
Commands for understanding node relationships and hierarchy

### 3. Design Specs & Analysis
Commands for extracting design specifications and visual data

### 4. Search & Discovery
Commands for finding content across files

### 5. Comparison & Diff
Commands for comparing designs and tracking changes

### 6. Bulk Operations
Commands for working with multiple nodes efficiently

### 7. Utility & Configuration
Commands for CLI setup and helper functions

---

## Command Specifications

### 1. Core Navigation & Retrieval

#### `figma get <url>`
Parse a Figma URL and fetch the specified node(s).

```bash
figma get "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221"
```

**Options:**
- `--depth <number>` - How many levels of children to include (default: 1, max: 10)
- `--output <format>` - Output format: `json` (default), `yaml`, `pretty`
- `--include <fields>` - Comma-separated fields to include: `styles`, `components`, `pluginData`

**Output (JSON):**
```json
{
  "id": "2270-190221",
  "name": "Global files list",
  "type": "TEXT",
  "content": "Global files list",
  "visible": true,
  "parent": {
    "id": "2270-190000",
    "name": "Header Container",
    "type": "FRAME"
  },
  "styles": {
    "fontSize": 16,
    "fontWeight": "600",
    "color": "#1A1A1A"
  }
}
```

#### `figma get <file_key> --node-id <id>`
Get a specific node by file key and node ID.

```bash
figma get grnVU2vAihHXwYgHryu2xE --node-id 2270-190221 --depth 3
```

#### `figma get <file_key> --ids <id1,id2,id3>`
Batch fetch multiple nodes in a single API call.

```bash
figma get grnVU2vAihHXwYgHryu2xE --ids 2270-190221,2270-190222,2270-190223
```

#### `figma text <url>`
Extract text content from a node.

```bash
figma text "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221"
```

**Output:**
```
Global files list
```

**Options:**
- `--recursive` - Extract text from all child nodes
- `--format <format>` - Output format: `plain` (default), `json`, `csv`

**Recursive Output (JSON):**
```json
{
  "id": "7450-15847",
  "text": "System Messages Container",
  "children": [
    {
      "id": "7450-15848",
      "text": "Messages are not self-deleting"
    },
    {
      "id": "7450-15849",
      "text": "Wire Cells enabled"
    }
  ]
}
```

---

### 2. Structure & Context

#### `figma context <url>`
Show the node's hierarchy and relationships.

```bash
figma context "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221"
```

**Output:**
```
Root (DOCUMENT)
└── Page: Cells - Drive
    └── FRAME: Header Container (2270-190000)
        └── TEXT: Global files list (2270-190221) ← CURRENT
```

**Options:**
- `--ancestors <number>` - How many parent levels to show (default: 5)
- `--siblings` - Include sibling nodes at same level
- `--children <number>` - How many child levels to show (default: 2)

#### `figma tree <file_key> --node-id <id> --depth <number>`
Show the subtree starting from a node.

```bash
figma tree grnVU2vAihHXwYgHryu2xE --node-id 7450-15847 --depth 3
```

#### `figma component <url>`
Get component definition if node is an instance, or component details if it's a component.

```bash
figma component "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221"
```

**Output (if instance):**
```json
{
  "instanceId": "2270-190221",
  "componentId": "comp-123",
  "componentName": "Button / Primary",
  "overrides": {
    "text": "Click me",
    "color": "#4CAF50"
  }
}
```

#### `figma instances <file_key> <component_key>`
Find all instances of a component.

```bash
figma instances grnVU2vAihHXwYgHryu2xE comp-123
```

#### `figma styles <url>`
Get styles applied to a node.

```bash
figma styles "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221"
```

**Output:**
```json
{
  "textStyle": {
    "id": "style-456",
    "name": "Heading / 16px / SemiBold",
    "fontSize": 16,
    "fontWeight": 600,
    "lineHeight": 24,
    "letterSpacing": 0
  },
  "fills": [
    {
      "type": "SOLID",
      "color": "#1A1A1A"
    }
  ]
}
```

#### `figma props <url>`
Get all properties of a node.

```bash
figma props "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221"
```

---

### 3. Design Specs & Analysis

#### `figma spec <url>`
Extract comprehensive design specifications.

```bash
figma spec "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/Cells---Files?node-id=7450-15847"
```

**Output:**
```json
{
  "node": {
    "id": "7450-15847",
    "name": "System Messages Container",
    "type": "FRAME",
    "size": {
      "width": 375,
      "height": 200
    }
  },
  "colors": {
    "background": "#FFFFFF",
    "text": "#666666",
    "accent": "#4CAF50",
    "border": "#E0E0E0"
  },
  "typography": {
    "heading": {
      "fontFamily": "Inter",
      "fontSize": 18,
      "fontWeight": 600
    },
    "body": {
      "fontFamily": "Inter",
      "fontSize": 14,
      "fontWeight": 400
    }
  },
  "spacing": {
    "padding": {
      "top": 16,
      "right": 16,
      "bottom": 16,
      "left": 16
    },
    "gap": 12
  },
  "borderRadius": 8,
  "shadows": [
    {
      "x": 0,
      "y": 2,
      "blur": 4,
      "color": "rgba(0,0,0,0.1)"
    }
  ]
}
```

#### `figma colors <file_key> --node-id <id>`
Extract color palette from a node.

```bash
figma colors grnVU2vAihHXwYgHryu2xE --node-id 7450-15847
```

**Output:**
```json
{
  "colors": [
    {
      "name": "Background",
      "hex": "#FFFFFF",
      "rgba": "rgba(255,255,255,1)"
    },
    {
      "name": "Primary Green",
      "hex": "#4CAF50",
      "rgba": "rgba(76,175,80,1)"
    }
  ],
  "palette": {
    "primary": "#4CAF50",
    "secondary": "#666666",
    "background": "#FFFFFF"
  }
}
```

#### `figma typography <file_key> --node-id <id>`
Extract typography styles.

```bash
figma typography grnVU2vAihHXwYgHryu2xE --node-id 7450-15847
```

**Output:**
```json
{
  "typography": [
    {
      "name": "Heading",
      "fontFamily": "Inter",
      "fontSize": 18,
      "fontWeight": 600,
      "lineHeight": 24
    },
    {
      "name": "Body",
      "fontFamily": "Inter",
      "fontSize": 14,
      "fontWeight": 400,
      "lineHeight": 20
    }
  ]
}
```

#### `figma export <url> --format <format>`
Export a node as SVG, PNG, or PDF.

```bash
figma export "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221" --format svg --output ./exports
```

**Options:**
- `--format <format>` - Export format: `svg`, `png`, `jpg`, `pdf` (default: png)
- `--scale <number>` - Export scale: 1, 2, 3 (default: 1)
- `--output <path>` - Output directory (default: ./figma-exports)

#### `figma screenshot <url>`
Get a visual preview of a node (returns image URL).

```bash
figma screenshot "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221"
```

---

### 4. Search & Discovery

#### `figma search <file_key> <query>`
Search for content by text matching.

```bash
figma search grnVU2vAihHXwYgHryu2xE "Global files list"
figma search grnVU2vAihHXwYgHryu2xE "self-deleting messages"
```

**Output:**
```json
{
  "results": [
    {
      "id": "2270-190221",
      "name": "Global files list",
      "type": "TEXT",
      "content": "Global files list",
      "url": "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221"
    }
  ]
}
```

**Options:**
- `--type <type>` - Filter by node type: `TEXT`, `FRAME`, `COMPONENT`, etc.
- `--case-sensitive` - Match case exactly
- `--exact` - Exact match only (no partial matches)

#### `figma search <file_key> --type <type>`
Find all nodes of a specific type.

```bash
figma search grnVU2vAihHXwYgHryu2xE --type TEXT
figma search grnVU2vAihHXwYgHryu2xE --type COMPONENT
```

#### `figma search <file_key> --name <pattern>`
Find nodes by name pattern.

```bash
figma search grnVU2vAihHXwYgHryu2xE --name "*message*"
figma search grnVU2vAihHXwYgHryu2xE --name "*banner*"
```

---

### 5. Comparison & Diff

#### `figma diff <url1> <url2>`
Compare two designs (can be from same or different files).

```bash
figma diff "https://www.figma.com/design/.../node-id=2270-190221" \
          "https://www.figma.com/design/.../node-id=2270-190222"
```

**Output:**
```json
{
  "changes": [
    {
      "type": "TEXT_CONTENT",
      "old": "Global files list",
      "new": "Files",
      "nodeId": "2270-190221"
    },
    {
      "type": "COLOR",
      "property": "fill",
      "old": "#4CAF50",
      "new": "#2196F3",
      "nodeId": "2270-190222"
    }
  ],
  "summary": {
    "textChanges": 1,
    "colorChanges": 1,
    "totalChanges": 2
  }
}
```

#### `figma diff <file_key> <id1> <id2>`
Compare two nodes within the same file.

```bash
figma diff grnVU2vAihHXwYgHryu2xE 2270-190221 2270-190222
```

#### `figma similar <url>`
Find similar nodes in the file based on properties.

```bash
figma similar "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221"
```

---

### 6. Bulk Operations

#### `figma list <file_key> --type <type>`
List all nodes of a specific type.

```bash
figma list grnVU2vAihHXwYgHryu2xE --type FRAME
figma list grnVU2vAihHXwYgHryu2xE --type TEXT --name "*message*"
```

**Output:**
```json
{
  "nodes": [
    {
      "id": "2270-190000",
      "name": "Header Container",
      "type": "FRAME"
    },
    {
      "id": "2270-200000",
      "name": "Content Container",
      "type": "FRAME"
    }
  ],
  "count": 2
}
```

#### `figma overview <file_key>`
Get a summary of the file structure.

```bash
figma overview grnVU2vAihHXwYgHryu2xE
```

**Output:**
```json
{
  "file": {
    "key": "grnVU2vAihHXwYgHryu2xE",
    "name": "Cells - Drive",
    "lastModified": "2026-01-23T10:30:00Z"
  },
  "summary": {
    "pages": 3,
    "frames": 45,
    "components": 12,
    "textNodes": 89,
    "totalNodes": 156
  },
  "pages": [
    {
      "id": "page-1",
      "name": "Cells - Drive",
      "frameCount": 23
    },
    {
      "id": "page-2",
      "name": "Cells - Files",
      "frameCount": 22
    }
  ]
}
```

#### `figma tree <file_key>`
Get the complete node tree for a file.

```bash
figma tree grnVU2vAihHXwYgHryu2xE --max-depth 5
```

---

### 7. Utility & Configuration

#### `figma parse <url>`
Parse a Figma URL and extract file key and node ID.

```bash
figma parse "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221&t=FJHwt3uytkvpEqRd-11"
```

**Output:**
```json
{
  "fileKey": "grnVU2vAihHXwYgHryu2xE",
  "nodeId": "2270-190221",
  "fileName": "Cells - Drive",
  "version": null,
  "timestamp": "FJHwt3uytkvpEqRd-11"
}
```

#### `figma config`
Manage CLI configuration.

```bash
# Set authentication token
figma config --token <personal-access-token>

# Set output format preference
figma config --output json

# Set default export directory
figma config --export-dir ./exports

# Show current configuration
figma config --show
```

**Configuration file (`~/.figma/config.json`):**
```json
{
  "token": "figd_xxxxxx",
  "defaultOutputFormat": "json",
  "exportDirectory": "./figma-exports",
  "maxDepth": 3
}
```

#### `figma batch <file_path>`
Execute multiple commands from a file.

```bash
figma batch commands.txt
```

**commands.txt:**
```
get "https://www.figma.com/design/.../node-id=2270-190221"
text "https://www.figma.com/design/.../node-id=7450-15847" --recursive
search grnVU2vAihHXwYgHryu2xE "self-deleting"
```

---

## MVP Implementation Phases

### Phase 1 - Core Navigation (Essential)
**Goal:** Enable basic node retrieval and text extraction

Commands:
- ✅ `figma get <url>` - Parse URL and fetch node
- ✅ `figma text <url>` - Extract text content
- ✅ `figma context <url>` - Show hierarchy
- ✅ `figma parse <url>` - URL parsing utility
- ✅ `figma config` - Configuration management

**Estimated Effort:** 2-3 weeks

### Phase 2 - Inspection & Discovery (Very Useful)
**Goal:** Enable searching and detailed inspection

Commands:
- ✅ `figma search <file_key> <query>` - Find content
- ✅ `figma search <file_key> --type <type>` - Find by type
- ✅ `figma search <file_key> --name <pattern>` - Find by name
- ✅ `figma spec <url>` - Get design specs
- ✅ `figma list <file_key> --type <type>` - List nodes
- ✅ `figma styles <url>` - Get applied styles
- ✅ `figma props <url>` - Get node properties

**Estimated Effort:** 2-3 weeks

### Phase 3 - Advanced Features (Nice to Have)
**Goal:** Enable export, comparison, and analysis

Commands:
- ✅ `figma export <url>` - Export visual assets
- ✅ `figma colors <file_key>` - Extract color palette
- ✅ `figma typography <file_key>` - Extract typography
- ✅ `figma diff <url1> <url2>` - Compare designs
- ✅ `figma similar <url>` - Find similar nodes
- ✅ `figma component <url>` - Component inspection
- ✅ `figma instances <file_key>` - Find component instances

**Estimated Effort:** 3-4 weeks

---

## LLM-Friendly Design Principles

### 1. URL-Based Input
Agents can simply paste Figma URLs without manual ID extraction:
```
figma get "https://www.figma.com/design/.../node-id=2270-190221"
```

### 2. JSON Output
All commands support JSON output for easy parsing:
```bash
figma get <url> --output json
```

### 3. Recursive Options
Get all child data in a single call:
```bash
figma text <url> --recursive
figma get <url> --depth 3
```

### 4. Comprehensive Filtering
Filter results by type, name, or content:
```bash
figma search <file_key> --type TEXT --name "*message*"
```

### 5. Context Awareness
Always provide parent/child/sibling information:
```bash
figma context <url> --siblings --children 2
```

### 6. Human-Readable + Machine-Readable
Support both formats for different use cases:
```bash
figma get <url> --output json   # For LLMs/scripts
figma get <url> --output pretty # For humans
```

### 7. Batch Operations
Minimize API calls with batch fetching:
```bash
figma get <file_key> --ids id1,id2,id3
figma batch commands.txt
```

---

## Example Workflows

### Workflow 1: Design Review - Text Change
```bash
# 1. Parse URL and extract node info
figma parse "https://www.figma.com/design/.../node-id=2270-190221"

# 2. Get current text content
figma text "https://www.figma.com/design/.../node-id=2270-190221"

# 3. Get full node context
figma context "https://www.figma.com/design/.../node-id=2270-190221" --siblings

# 4. Get design specs if needed
figma spec "https://www.figma.com/design/.../node-id=2270-190221"
```

### Workflow 2: Complex Component Analysis
```bash
# 1. Get design node with children
figma get "https://www.figma.com/design/.../node-id=7450-15847" --depth 3

# 2. Extract all text recursively
figma text "https://www.figma.com/design/.../node-id=7450-15847" --recursive

# 3. Get design specifications
figma spec "https://www.figma.com/design/.../node-id=7450-15847"

# 4. Search for related system messages
figma search grnVU2vAihHXwYgHryu2xE "self-deleting"

# 5. Find all message banners
figma list grnVU2vAihHXwYgHryu2xE --type FRAME --name "*message*"

# 6. Extract color palette
figma colors grnVU2vAihHXwYgHryu2xE --node-id 7450-15847
```

### Workflow 3: Design System Audit
```bash
# 1. Get file overview
figma overview grnVU2vAihHXwYgHryu2xE

# 2. List all components
figma list grnVU2vAihHXwYgHryu2xE --type COMPONENT

# 3. Extract color palette from design system frame
figma colors grnVU2vAihHXwYgHryu2xE --node-id design-system-frame-id

# 4. Export components as images
figma export "https://www.figma.com/design/.../node-id=comp-id" --format svg
```

---

## Technical Considerations

### Rate Limiting
- Implement exponential backoff for 429 errors
- Respect Figma's rate limit tiers (10-150 requests/minute)
- Use caching to minimize redundant API calls

### Authentication
- Support personal access tokens (primary)
- OAuth 2.0 for production use (future)
- Secure token storage (system keychain where available)

### Performance
- Batch API calls when possible
- Parallel requests for independent operations
- Cache frequently accessed data

### Error Handling
- Clear error messages with actionable guidance
- Retry logic for transient failures
- Graceful degradation for rate limit scenarios

---

## Future Enhancements

### Potential Future Commands
- `figma tokens <url>` - Extract design tokens/variables
- `figma generate <url> --template <template>` - Generate code from designs
- `figma sync <file_key>` - Sync design changes to codebase
- `figma watch <file_key>` - Watch for file changes via webhooks
- `figma validate <file_key> --rules <rules>` - Validate design against rules
- `figma annotate <url> --message <message>` - Add comments to Figma
- `figma version <file_key>` - Version history and comparison

### Integration Possibilities
- GitHub Actions for CI/CD
- VS Code extension for inline Figma inspection
- IDE plugins for real-time design updates
- Jira/Notion integrations for design tracking

---

## References

- [Figma REST API Documentation](https://developers.figma.com/docs/rest-api/)
- [Figma REST API OpenAPI Spec](https://github.com/figma/rest-api-spec)
- [GitHub Resources](./resources.md) - SDKs, CLI tools, and design system projects
- [Figma API Overview](./figma-api-overview.md) - Comprehensive API reference
