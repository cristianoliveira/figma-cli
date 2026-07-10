# Figma CLI

A command-line interface tool for exploring and interacting with Figma designs, designed to be LLM-friendly for AI agents.

[![Go Version](https://img.shields.io/badge/go-1.25.5-blue)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

## Overview

`figma-cli` enables developers and AI agents to fetch, analyze, and understand Figma design data directly from the command line. It parses Figma URLs, retrieves design metadata, and provides structured JSON output for easy integration with automation workflows.

The tool is built with a focus on being **LLM-friendly**, **URL-based**, and **context-aware**, making it ideal for AI-assisted design review, code generation, and asset export workflows.

## Features

- **URL Parsing**: Accept Figma URLs directly—no manual ID extraction required
- **Structured JSON Output**: All commands produce JSON for easy parsing by AI agents
- **Node Hierarchy**: Always provides full context of node relationships and parent/child structure
- **Batch Operations**: Minimize API calls through efficient batch fetching and caching
- **Design Inspection**: Explore layers, components, styles, and text content
- **Asset Export**: Export images, SVGs, and other design assets
- **Plugin Simulation**: Simulate Figma plugin behavior for local testing

## Installation

### Prerequisites

- [Go 1.25.5+](https://golang.org/dl/)
- [Figma Personal Access Token](https://www.figma.com/developers/api#access-tokens)

### Build from Source

```bash
git clone https://github.com/cristianoliveira/figma-cli.git
cd figma-cli
go build -o figma ./cmd/figma
```

### Install Globally

```bash
go install github.com/cristianoliveira/figma-cli/cmd/figma@latest
```

### Using Nix (Development)

If you have Nix and direnv installed, enable the reproducible development shell once:

```bash
direnv allow
```

Otherwise enter it manually:

```bash
nix develop
```

The shell provides the project Go toolchain, `golangci-lint`, and `goimports`.

## Configuration

Set your Figma access token as an environment variable:

```bash
export FIGMA_ACCESS_TOKEN="your-personal-access-token"
```

Or create a `.env` file in the project root:

```bash
FIGMA_ACCESS_TOKEN=your-personal-access-token
```

## Usage

### Basic Command Structure

```bash
figma [command] [options] <figma-url>
```

### Examples

```bash
# Fetch file metadata
figma meta "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221"

# List components within a specific node tree
figma components --id 123:456 "https://www.figma.com/file/abc123/My-Design"

# Extract ordered text layers from a design
figma texts "https://www.figma.com/design/xyz456/Another-Design?node-id=123-456"

# Export one frame as PNG
figma export --format png "https://www.figma.com/design/abc123/My-Design?node-id=123-456"

# Download image, instance, and vector assets from a frame
figma assets --output ./assets "https://www.figma.com/design/abc123/My-Design?node-id=123-456"
figma assets --json --output ./assets "https://www.figma.com/design/abc123/My-Design?node-id=123-456"

# Generate design tokens (CSS variables, Tailwind theme, or JSON)
figma tokens --format css "https://www.figma.com/design/abc123/My-Design"
figma tokens --format tailwind --output tailwind.tokens.js "https://www.figma.com/design/abc123/My-Design"
figma tokens --source styles --prefix fig- "abc123"

# Generate CSS (layout + fills + type) for a frame — scriptable Dev Mode
echo '/* customization-page.css */' > styles.css
figma css "https://www.figma.com/design/abc123/My-Design?node-id=42:1" >> styles.css
```

### JSON Output

All commands produce JSON output for easy parsing:

```json
{
  "file": {
    "key": "grnVU2vAihHXwYgHryu2xE",
    "name": "-Cells--Drive",
    "lastModified": "2025-01-23T08:33:00Z"
  },
  "node": {
    "id": "2270:190221",
    "name": "Global files list",
    "type": "FRAME",
    "children": [...]
  }
}
```

## Project Structure

```
figma-cli/
├── cmd/                    # Cobra command implementations (root, texts, export, ...)
│   └── figma/              # Main CLI entry point (main.go)
├── internal/               # Private application code
│   ├── assets/             # Asset discovery, export, and downloads
│   ├── cli/                # CLI runtime wiring
│   ├── comments/           # Comment retrieval and node scoping
│   ├── diff/               # Pure design-diff use cases
│   ├── env/                # Environment configuration (token loading)
│   ├── extract/            # Pure document-tree transformations
│   └── figma/              # Figma API boundary (client, document, URLs, export)
│       └── api/            # Generated Figma REST API types (from openapi/, DO NOT EDIT)
├── openapi/                # OpenAPI spec and oapi-codegen config (source of truth for api.gen.go)
├── scripts/                # Codegen helpers (regenerate API types)
├── testdata/               # Test fixtures
├── plans/                  # Work planning notes
├── research/               # Research notes
├── docs/                   # Documentation
├── flake.nix               # Nix flake (dev shell)
├── go.mod                  # Go module definition
└── README.md               # This file
```

## Development

Run quality checks from the flake shell:

```bash
golangci-lint run ./...
go test ./...
go vet ./...
```

### Multi-Agent Workflow

Refer to [AGENTS.md](AGENTS.md) for guidelines on ordering agents and completing work sessions.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure your code follows Go conventions and includes appropriate tests.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Figma API](https://www.figma.com/developers/api) for the underlying design platform
- The Go community for excellent CLI tooling libraries
- AI agents for helping build this tool
