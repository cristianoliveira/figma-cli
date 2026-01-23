---
title: "Authorization Scopes"
tags: ["figma-api", "resources", "api-docs", "authorization"]
---
# Authorization Scopes

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

