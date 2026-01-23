# Design System & Design-to-Code Tools

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
