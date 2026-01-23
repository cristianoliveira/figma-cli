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
go build -o figma-cli ./cmd/figma-cli
```

### Install Globally

```bash
go install github.com/cristianoliveira/figma-cli/cmd/figma-cli@latest
```

### Using Nix (Development)

If you have Nix installed, you can enter a development shell with all dependencies:

```bash
nix develop
```

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
figma-cli [command] [options] <figma-url>
```

### Examples

```bash
# Parse a Figma URL and extract file metadata
figma-cli parse "https://www.figma.com/design/grnVU2vAihHXwYgHryu2xE/-Cells--Drive?node-id=2270-190221"

# Fetch node hierarchy for a specific node
figma-cli nodes --hierarchy "https://www.figma.com/file/abc123/My-Design"

# Extract all text layers from a design
figma-cli text "https://www.figma.com/design/xyz456/Another-Design"

# Export assets from a frame
figma-cli export --format png --scale 2 "https://www.figma.com/design/abc123/My-Design?node-id=123-456"
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
├── cmd/figma-cli/          # Main CLI entry point
├── internal/               # Private application code
│   ├── api/               # Figma API client
│   ├── parser/            # URL and node parsing
│   └── export/            # Asset export logic
├── pkg/                   # Public reusable packages
│   ├── figma/             # Figma data structures
│   └── cli/               # CLI utilities
├── docs/                  # Documentation
├── .beads/                # Issue tracking (bd)
├── go.mod                 # Go module definition
└── README.md              # This file
```

## Development

### Issue Tracking

This project uses [bd (beads)](https://github.com/beads) for issue tracking. To get started:

```bash
bd onboard
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --status in_progress  # Claim work
bd close <id>         # Complete work
```

### Multi-Agent Workflow

Refer to [AGENTS.md](AGENTS.md) for guidelines on ordering agents and completing work sessions.

### Design Documentation

See [doc-cli-design.md](doc-cli-design.md) for detailed CLI design and use cases.

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