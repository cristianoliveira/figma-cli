# Resources

## Figma

Figma is a collaborative web application for interface design, with additional offline features enabled by desktop applications for macOS and Windows.

### Figma API Overview

The Figma API is a comprehensive REST API that provides programmatic access to Figma's design platform. It follows modern API design patterns with clear resource separation, versioning, and authentication mechanisms. The API is officially maintained with a complete OpenAPI 3.1.0 specification (v0.36.0) that enables code generation and type-safe client libraries.

#### API Architecture

- **Base URLs**:
  - Standard API: `https://api.figma.com`
  - Government: `https://api.figma-gov.com`

- **Versioning**: Path-based versioning (`/v1/` for most endpoints, `/v2/` for webhooks)

- **Request/Response Format**:
  - Content-Type: `application/json` for all API payloads
  - Authentication headers:
    - Personal Access Token: `X-Figma-Token: <token>`
    - OAuth 2.0: `Authorization: Bearer <token>`

- **OpenAPI Specification**:
  - Official spec at [github.com/figma/rest-api-spec](https://github.com/figma/rest-api-spec)
  - Version 0.36.0 (beta status due to API complexity)
  - 39 endpoints across 14 functional categories
  - 359 distinct schemas
  - TypeScript types available via `@figma/rest-api-spec` npm package

#### Authentication Methods

1. **Personal Access Tokens**:
   - Generated from Account settings → Security → Personal access tokens
   - Use case: Scripts and personal automation
   - Rate limits tracked per-user, per-plan basis

2. **OAuth 2.0** (Recommended for Production):
   - Authorization code flow with PKCE (S256 method)
   - Tokens expire after 90 days (refresh tokens available)
   - Required for Activity Logs API and Discovery API
   - Rate limits tracked per-user, per-plan, per-app basis

#### Authorization Scopes

Granular scopes control API access (20+ defined scopes):

| Scope | Purpose |
|-------|---------|
| `file_content:read` | Read file contents, nodes, editor type |
| `file_comments:read` | Read comments in files |
| `file_comments:write` | Post/delete comments and reactions |
| `file_variables:read` | Read variables (Enterprise only) |
| `file_variables:write` | Write variables (Enterprise only) |
| `file_dev_resources:read` | Read Dev Mode resources |
| `file_dev_resources:write` | Create/manage Dev Mode resources |
| `file_metadata:read` | Read file metadata |
| `projects:read` | List projects and files |
| `team_library_content:read` | Read published components/styles |
| `webhooks:read` | List webhooks |
| `webhooks:write` | Create and manage webhooks |
| `org:activity_log_read` | Read organization activity logs (Enterprise admin) |

#### Rate Limiting

Rate limits are multi-dimensional based on:
1. User seat type (View/Collab vs Dev/Full)
2. API endpoint tier (Tier 1, 2, or 3)
3. Resource plan (Starter, Professional, Organization, Enterprise)

| API Tier | Seat Type | Starter | Professional | Organization | Enterprise |
|----------|-----------|---------|--------------|--------------|------------|
| Tier 1   | View/Collab | Up to 6/month | Up to 6/month | Up to 6/month | Up to 6/month |
|          | Dev/Full  | 10/min  | 15/min       | 20/min       | 20/min     |
| Tier 2   | View/Collab | Up to 5/min | Up to 5/min | Up to 5/min | Up to 5/min |
|          | Dev/Full  | 25/min  | 50/min       | 100/min      | 100/min    |
| Tier 3   | View/Collab | Up to 10/min | Up to 10/min | Up to 10/min | Up to 10/min |
|          | Dev/Full  | 50/min  | 100/min      | 150/min      | 150/min    |

**Key Points**:
- Uses leaky bucket algorithm
- 429 errors include `Retry-After` header
- View/Collab seats have significantly lower limits
- Dev/Full seats have higher limits (up to 150/min for Tier 3 Enterprise)

#### API Capabilities

The Figma API provides 14 major capability categories with 39 distinct endpoints:

**1. Files (6 endpoints)** - Primary access point for design file data:
- `GET /v1/files/{file_key}` - Get complete file JSON with document structure
- `GET /v1/files/{file_key}/nodes` - Get JSON for specific nodes by ID
- `GET /v1/images/{file_key}` - Render images (PNG, JPG, SVG, PDF)
- `GET /v1/files/{file_key}/images` - Get image fills
- `GET /v1/files/{file_key}/meta` - Get file metadata
- `GET /v1/files/{file_key}/versions` - Get version history

**2. Comments & Reactions (6 endpoints)** - Collaboration features:
- `GET /v1/files/{file_key}/comments` - List comments
- `POST /v1/files/{file_key}/comments` - Add comment
- `DELETE /v1/files/{file_key}/comments/{comment_id}` - Delete comment
- `POST /v1/files/{file_key}/comments/{comment_id}/reactions` - Add reaction
- `GET /v1/files/{file_key}/comments/{comment_id}/reactions` - List reactions
- `DELETE /v1/files/{file_key}/comments/{comment_id}/reactions/{emoji}` - Delete reaction

**3. Projects & Teams (2 endpoints)**:
- `GET /v1/teams/{team_id}/projects` - List projects in a team
- `GET /v1/projects/{project_id}/files` - List files in a project

**4. Users (1 endpoint)**:
- `GET /v1/me` - Get current authenticated user

**5. Components & Component Sets (6 endpoints)** - Design system access:
- `GET /v1/teams/{team_id}/components` - List published components
- `GET /v1/components/{key}` - Get specific component
- `GET /v1/teams/{team_id}/component_sets` - List published component sets
- `GET /v1/file_components` - Get components in a specific file
- `GET /v1/component_sets/{key}` - Get specific component set

**6. Styles (4 endpoints)**:
- `GET /v1/teams/{team_id}/styles` - List published styles
- `GET /v1/styles/{key}` - Get specific style
- `GET /v1/file_styles` - Get styles in a specific file

**7. Webhooks (4 endpoints, v2)** - Real-time event notifications:
- `GET /v2/webhooks` - List webhooks
- `POST /v2/webhooks` - Create webhook
- `DELETE /v2/webhooks/{webhook_id}` - Delete webhook
- `PATCH /v2/webhooks/{webhook_id}` - Update webhook

**8. Variables (3 endpoints, Enterprise)** - Design token management:
- `GET /v1/files/{file_key}/variables/local` - Get local variables
- `GET /v1/teams/{team_id}/variables/published` - Get published variables
- `POST /v1/files/{file_key}/variables` - Bulk create/update/delete variables

**9. Dev Resources (3 endpoints)** - Dev Mode resource management:
- `GET /v1/files/{file_key}/dev_resources` - Get dev resources
- `POST /v1/dev_resources` - Bulk create dev resources
- `PATCH /v1/dev_resources` - Bulk update dev resources

**10. Activity Logs (1 endpoint, Enterprise admin)**:
- `GET /v1/activity_logs` - Get organization activity logs

**11. Library Analytics (6 endpoints)** - Design system usage analytics:
- `GET /v1/analytics/libraries/{file_key}/component/usage` - Component usage
- `GET /v1/analytics/libraries/{file_key}/style/usage` - Style usage
- `GET /v1/analytics/libraries/{file_key}/variable/usage` - Variable usage
- `GET /v1/analytics/libraries/{file_key}/actions` - Library actions

#### Capabilities Summary

**Strong Read Capabilities:**
- Complete file document structure with node hierarchies
- Design system access (components, styles, variables)
- Image export (PNG, JPG, SVG, PDF)
- Version history and metadata
- Analytics and usage data

**Limited Write Capabilities:**
- Comments and reactions (collaboration)
- Variables (Enterprise only, bulk operations)
- Dev resources (bulk operations)
- Webhooks management

**Not Available:**
- No direct design creation/modification (cannot create layers, frames, etc.)
- No file upload capabilities
- No plugin data manipulation beyond `pluginData` parameter

#### Data Models

**Node Types:**
- `DOCUMENT` - Root node containing pages
- `CANVAS` (Page) - Top-level page in a file
- `FRAME` - Container with layout
- `GROUP` - Grouped elements
- `SECTION` - Organizational grouping
- `VECTOR` - Vector paths
- `BOOLEAN_OPERATION` - Union, subtract, intersect, exclude
- `STAR`, `LINE`, `ELLIPSE`, `REGULAR_POLYGON`, `RECTANGLE` - Primitive shapes
- `TEXT` - Text elements
- `COMPONENT` - Reusable design element
- `COMPONENT_SET` - Collection of variants
- `INSTANCE` - Instance of a component with overrides

**Design Properties:**
- **Paint**: Fill definitions (solid, gradient, image, video)
- **Layout Constraints**: Positioning behavior within frames
- **Transform**: 2D transformation matrices

**Component System:**
- **Component**: Reusable design element published to team libraries
- **Component Set**: Collection of component variants
- **Instance**: Instance of a component with overrides

**Variable System (Enterprise):**
- **Variable**: Design variable with values per mode
- **Variable Collection**: Group of variables with shared modes
- **Variable Alias**: Reference to a variable value

#### Best Practices

**Authentication:**
1. Use OAuth for production apps - provides better security and per-app rate limiting
2. Use personal tokens for CLI/simple scripts - suitable for personal automation
3. Implement proper token refresh - OAuth access tokens expire after 90 days

**Error Handling & Rate Limiting:**
1. Implement retry logic with exponential backoff - especially for 429 errors
2. Monitor rate limit headers - use `Retry-After`, `X-Figma-Rate-Limit-Type`
3. Handle all HTTP status codes - 400, 403, 404, 429, 500
4. Provide helpful error messages - guide users when they hit limits

**Performance:**
1. Cache results - reduce API calls by caching file data locally
2. Use specific node queries - `GET /v1/files/{file_key}/nodes` for partial data
3. Batch requests when possible - combine multiple operations (e.g., variables)
4. Limit response size - use `depth` parameter to control nested nodes

**Security:**
1. Use granular scopes - avoid deprecated `files:read` scope
2. Store tokens securely - never commit tokens to version control
3. Validate input - sanitize all user-provided parameters
4. Use HTTPS - always use HTTPS for API requests

#### Use Cases

The Figma API is particularly suited for:

1. **Design-to-Code Pipelines** - Extract design tokens, generate component code, export assets
2. **Design System Tooling** - Document design systems, validate consistency, sync tokens
3. **Design Review Automation** - Automated comment bots, compliance analysis
4. **Analytics Dashboards** - Track library usage, monitor adoption, analyze activity
5. **CI/CD Integration** - Trigger builds on publishes, validate changes, sync to version control
6. **Asset Management** - Bulk export images/icons, organize assets, generate catalogs
7. **Dev Mode Integration** - Attach code snippets, link dev resources, streamline handoff

#### Official Documentation

- [Figma REST API Documentation](https://developers.figma.com/docs/rest-api/)
- [Figma REST API Authentication](https://developers.figma.com/docs/rest-api/authentication/)
- [Figma REST API Rate Limits](https://developers.figma.com/docs/rest-api/rate-limits/)
- [Figma REST API Errors](https://developers.figma.com/docs/rest-api/errors/)
- [Figma REST API Scopes](https://developers.figma.com/docs/rest-api/scopes/)
- [Figma API Demo Repository](https://github.com/figma/figma-api-demo)

### Other projects like this

- https://github.com/mttwhlly/get-figma-text
- https://github.com/kataras/figma-extractor

### GitHub Resources & References

Comprehensive exploration of GitHub repositories related to Figma API, CLI tools, and design system integration.

#### SDKs and Client Libraries

**TypeScript/JavaScript (Most Mature Ecosystem)**

1. **[jem-computer/figma-js](https://github.com/jem-computer/figma-js)** (495 ⭐, updated 2026-01-13)
   - Little wrapper (+ types) for the Figma API
   - TypeScript support, personal access token & OAuth authentication, promise-based API
   - Used by other projects like figma-graphql
   - **Recommended for**: Simple wrapper ideal for CLI tools, well-documented with full TypeScript types

2. **[didoo/figma-api](https://github.com/didoo/figma-api)** (257 ⭐, updated 2026-01-21)
   - Figma REST API implementation with TypeScript, Promises & ES6
   - Thin layer on official REST API spec, uses Axios, supports Node.js & browser
   - Version 2.0 aligns with official API, comprehensive coverage including variables, webhooks, analytics
   - **Recommended for**: Direct mapping to Figma API endpoints, comprehensive coverage

3. **[braposo/figma-graphql](https://github.com/braposo/figma-graphql)** (394 ⭐, updated 2026-01-13)
   - The reimagined Figma API (super)powered by GraphQL
   - GraphQL wrapper over Figma API, potentially simpler query interface
   - **Recommended for**: Alternative query pattern if GraphQL preferred over REST

**Python**

4. **[Amatobahn/FigmaPy](https://github.com/Amatobahn/FigmaPy)** (57 ⭐, updated 2025-07-14)
   - An unofficial Python3+ wrapper for Figma API
   - Object-oriented interface, supports file and image operations
   - PyPI package available

**Go**

5. **[torie/figma](https://github.com/torie/figma)** (12 ⭐, updated 2025-08-04)
   - A Golang package for interacting with the Figma APIs
   - Go-native client, covers core API endpoints
   - **Recommended for**: Natural fit for Go CLI tools

6. **[figma/terraform-provider-figma](https://github.com/figma/terraform-provider-figma)** (2 ⭐, updated 2024-10-10)
   - Terraform provider for Figma (official)
   - Infrastructure-as-code approach to Figma resources
   - **Recommended for**: Reference implementation for CLI tool patterns

**Dart**

7. **[arnemolland/figma](https://github.com/arnemolland/figma)** (27 ⭐, updated 2025-11-09)
   - Figma API client written in pure Dart
   - Full API coverage, typed responses, OAuth support, variables support
   - **Recommended for**: Excellent for Dart/Flutter CLI tools

**Rust**

8. **[gridaco/figma-api](https://github.com/gridaco/figma-api)** (2 ⭐, updated 2025-12-28)
   - Figma Rest API rust bindings
   - Rust-native client, recently updated

**Official Figma Resources**

9. **[figma/rest-api-spec](https://github.com/figma/rest-api-spec)** (185 ⭐, updated 2026-01-23)
   - OpenAPI specification and types for the Figma REST API
   - Official TypeScript types package (`@figma/rest-api-spec`), OpenAPI 3.1 spec
   - **Recommended for**: Essential for type-safe implementations, can generate clients

10. **[figma/figma-api-demo](https://github.com/figma/figma-api-demo)** (1,337 ⭐, updated 2026-01-13)
    - Official Figma API demo application
    - Reference implementation, authentication examples
    - **Recommended for**: Good for understanding API patterns and authentication flows

**Utility Libraries**

11. **[figma-tools/figma-transformer](https://github.com/figma-tools/figma-transformer)** (58 ⭐, updated 2026-01-05)
    - A tiny utility library that makes the Figma API more human friendly
    - Transforms API responses with shortcuts, enriches style/component data
    - **Recommended for**: Useful for processing Figma file data, simplifies complex nested structures

**Key Insights for SDK Selection:**
- TypeScript/JavaScript ecosystem is most active with several mature options
- Use `@figma/rest-api-spec` for official types regardless of language
- Few libraries explicitly handle rate limiting automatically
- Consider using `didoo/figma-api` or `jem-computer/figma-js` as foundation

#### CLI Tools and Projects

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

#### Design System & Design-to-Code Tools

**Official Figma Tools**

1. **[figma/code-connect](https://github.com/figma/code-connect)** (1,359 ⭐, updated 2026-01-23)
   - Tool for connecting design system components in code with Figma design systems
   - Generates code snippets for Dev Mode, supports React, React Native, HTML, SwiftUI, Jetpack Compose
   - Maps component properties from code to Figma, enables dynamic code examples
   - **Learning Points**: Official Figma tool showing enterprise design system integration patterns

2. **[tokens-studio/figma-plugin](https://github.com/tokens-studio/figma-plugin)** (1,534 ⭐, updated 2026-01-21)
   - Official Figma plugin for design tokens management
   - Token management, style synchronization, variable support
   - **Learning Points**: Token management patterns and variable handling

**Popular Code Generation Tools**

3. **[bernaferrari/FigmaToCode](https://github.com/bernaferrari/FigmaToCode)** (4,686 ⭐, updated 2026-01-23)
   - Generate responsive pages/apps in HTML, Tailwind, Flutter, SwiftUI
   - Multi-step conversion process, intermediate representation (AltNodes), layout optimization
   - **Learning Points**: Plugin architecture with sophisticated transformation pipeline

4. **[aloisdeniel/figma-to-flutter](https://github.com/aloisdeniel/figma-to-flutter)** (879 ⭐, updated 2026-01-13)
   - Dart code generator that converts Figma components to Flutter widgets
   - Flutter widget generation, platform-specific

**Token & Design System Tools**

5. **[RedMadRobot/figma-export](https://github.com/RedMadRobot/figma-export)** (801 ⭐, updated 2026-01-23)
   - CLI utility to export colors, typography, icons, images to Xcode/Android Studio
   - Dark mode support, SwiftUI/Jetpack Compose generation, template system, CI/CD integration
   - Uses endpoints: `/v1/files/:fileId/styles`, `/v1/files/:fileId/variables/local`, `/v1/files/:fileId/components`, `/v1/files/:fileId/nodes`, `/v1/images/:fileId`
   - **Most Relevant**: Most relevant CLI reference - shows complete Figma API integration with configurable templates

6. **[mikaelvesavuori/figmagic](https://github.com/mikaelvesavuori/figmagic)** (852 ⭐, updated 2026-01-20)
   - Generate design tokens, export graphics, extract React components from Figma
   - 15+ token types, React component generation, graphics export, GitHub Action
   - Design tokens as first-class concept, structured Figma document requirements
   - **Recommended for**: Extensive configuration system and token-focused approach

**Common API Endpoints Used:**
- `GET /v1/files/:key` - Full document tree
- `GET /v1/files/:key/styles` - Color/text styles
- `GET /v1/files/:key/variables/local` - Design tokens (variables)
- `GET /v1/files/:key/components` - Component metadata
- `GET /v1/files/:key/nodes?ids=...` - Specific nodes
- `GET /v1/images/:key?ids=...&format=...` - Image export

**Architecture Approaches Observed:**
- **Intermediate Representation**: Convert Figma nodes to custom JSON structures before code generation
- **Template Systems**: Use Stencil, Handlebars, or custom templating for code generation
- **Configuration First**: Extensive YAML/JSON config files for customization
- **Token Processing**: Support for 15+ token types with unit conversion and platform-specific formatting

**Recommended CLI Structure for figma-cli:**
- `figma-cli tokens` - design token extraction
- `figma-cli components` - component/code generation
- `figma-cli assets` - image/icon export
- `figma-cli sync` - continuous synchronization

**Key Implementation Insights:**
- Implement robust rate limiting and retry logic
- Use secure token storage (system keychain where available)
- Support both interactive and scriptable modes
- Provide clear configuration options
- Implement caching to respect API limits
- Support webhook integration for real-time updates
