---
title: "Authentication Methods"
tags: ["figma-api", "resources", "api-docs", "authentication"]
---
# Authentication Methods

1. **Personal Access Tokens**:
   - Generated from Account settings → Security → Personal access tokens
   - Use case: Scripts and personal automation
   - Rate limits tracked per-user, per-plan basis

2. **OAuth 2.0** (Recommended for Production):
   - Authorization code flow with PKCE (S256 method)
   - Tokens expire after 90 days (refresh tokens available)
   - Required for Activity Logs API and Discovery API
   - Rate limits tracked per-user, per-plan, per-app basis

