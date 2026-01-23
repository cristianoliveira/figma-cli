# Best Practices

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

