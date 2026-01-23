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
