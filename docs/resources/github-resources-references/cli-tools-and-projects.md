# CLI Tools and Projects

**Design Token Synchronization**

1. **[B3nnyL/figgo](https://github.com/B3nnyL/figgo)** (308 ⭐, updated 2025-11-11)
   - CLI tool to sync Figma design tokens with local codebase
   - Syncs colors, typography, spacing from Figma frames
   - Supports both global and local configuration, output formats: JavaScript, SCSS variables
   - Interactive setup with `--init` flag
   - **Learning Points**: Clean CLI design with config management, token synchronization patterns

2. **[whynotmake-it/figmage](https://github.com/whynotmake-it/figmage)** (70 ⭐, updated 2026-01-21)
   - CLI tool for generating Flutter themes from Figma files
   - Generates Flutter ThemeExtension classes, supports Figma styles and variables with modes
   - Color, typography, number variables support, BuildContext extensions for easy access

**Asset Export Tools**

3. **[alexchantastic/figma-export](https://github.com/alexchantastic/figma-export)** (115 ⭐, updated 2026-01-13)
   - CLI tool to bulk export Figma, FigJam, and Figma Slides files to .fig/.jam/.deck format
   - Bulk export by team, project, or drafts
   - Uses Playwright for automated downloads, parallel downloads support, retry failed downloads
   - **Learning Points**: Hybrid approach using API + browser automation, fault tolerance design

4. **[jacobtyq/export-figma-svg](https://github.com/jacobtyq/export-figma-svg)** (44 ⭐, updated 2025-12-15)
   - Export SVGs from your Figma project via CLI
   - Exports SVGs from Figma components in a frame, rate limiting handling (20 requests per 45 seconds)
   - Filtering of private components
   - **Learning Points**: Simple single-purpose tool, rate limit handling

**Backup Solutions**

5. **[mimshins/figma-backup](https://github.com/mimshins/figma-backup)** (84 ⭐, updated 2026-01-13)
   - Node.js CLI to backup Figma files and store them as local .fig files
   - Interactive and non-interactive modes, uses Puppeteer for downloading .fig files
   - Docker container support, structured backup organization

**Professional CLI Tools**

6. **[alexey1312/ExFig](https://github.com/alexey1312/ExFig)** (3 ⭐, actively maintained)
   - Fast Figma CLI with parallel exports, batch processing & smart caching for iOS, Android & Flutter
   - Parallel exports with smart caching, batch processing for multiple files
   - Support for iOS (SwiftUI/UIKit), Android (Jetpack Compose), Flutter, React/TypeScript
   - Dark mode, high contrast, RTL support, CI/CD ready with GitHub Action
   - **Learning Points**: Professional-grade CLI with extensive platform support, caching strategies

7. **[tonykolomeytsev/figx](https://github.com/tonykolomeytsev/figx)** (31 ⭐, updated 2025-12-10)
   - Pragmatic CLI tool for importing design assets from Figma into codebase
   - Cross-platform (macOS, Windows, Linux), built-in import profiles for Android, Compose, WebP, SVG, PDF
   - Secure token storage using system keychain, resource query and explanation commands

**Simple Utility Tools**

8. **[acpplife/figma-json](https://github.com/acpplife/figma-json)** (2 ⭐, updated 2026-01-14)
   - CLI tool for downloading Figma file data as JSON
   - Secure token management, support for various Figma URL formats
   - Pretty-printed JSON output, node-specific downloads

9. **[schpet/figma-cli](https://github.com/schpet/figma-cli)** (1 ⭐, updated 2025-12-05)
   - Command line tool to copy Figma nodes as images to clipboard (Deno implementation)
   - Copy nodes to clipboard (macOS only), export nodes to files, get direct image URLs

**Other Tools**

10. **[kreako/fig2json](https://github.com/kreako/fig2json)** (11 ⭐, updated 2026-01-22)
    - Convert Figma .fig files to LLM-friendly JSON format
    - Parses local .fig files (no API needed), removes Figma-specific metadata
    - Optimized for AI consumption, raw and transformed JSON outputs

11. **[yuanqing/figma-plugins-stats](https://github.com/yuanqing/figma-plugins-stats)** (66 ⭐, deprecated)
    - CLI to get live and historical stats for Figma plugins
    - Plugin stats with sparklines, historical data back to April 2020
    - Uses internal Figma APIs (not official REST API)

**CLI Design Patterns Observed:**
- Subcommand structure (e.g., `figma node copy`, `figx import`)
- Configuration files (YAML, JSON, .figma)
- Environment variable support for tokens
- Interactive setup wizards
- Rate limiting implementations (10-20 requests/minute)
- Token management with secure storage
- Batch processing for multiple files
- Caching strategies to minimize API calls

