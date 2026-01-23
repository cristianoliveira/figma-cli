# SDKs and Client Libraries

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

