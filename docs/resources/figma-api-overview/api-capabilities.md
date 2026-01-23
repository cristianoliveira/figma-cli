# API Capabilities

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

