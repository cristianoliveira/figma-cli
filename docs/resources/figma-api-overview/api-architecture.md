---
title: "API Architecture"
tags: ["figma-api", "resources", "api-docs", "architecture"]
---
# API Architecture

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

