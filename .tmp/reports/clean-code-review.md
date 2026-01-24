# Clean Code Review Report
**Date**: Sat Jan 24 2026
**Reviewer**: Janitor Agent
**Focus**: Uncle Bob's Clean Code Principles + Test Quality

## Executive Summary
Overall Clean Code Score: 6/10
- Test Quality Score: 7/10
- Architecture Score: 5/10
- Code Quality Score: 6/10

## Key Findings
1. **Critical**: 25+ API methods unimplemented (P0) - violates "Complete Implementation" principle
2. **Major**: Client interface has 38 methods - violates Interface Segregation Principle (ISP)
3. **Major**: Error handling gaps - ignoring flag retrieval errors
4. **Moderate**: Over-mocking in unit tests (testing the mock itself)
5. **Moderate**: TODO comments in production code (should be issues)

---

## Detailed Findings

### Test Quality Issues

#### Over-Mocking in Client Tests
- **File**: internal/api/client_test.go
- **Line**: 24-29
- **Principle**: Test behavior, not implementation details
- **Problem**: Test is verifying that the mock returns what was set up, not testing real client behavior. This is testing the mock framework, not the system under test.
- **Example**: 
  \`\`\`go
  s.mockClient.On("GetFile", ctx, "test", mock.Anything).Return(expectedFile, nil)
  file, err := s.mockClient.GetFile(ctx, "test")
  assert.NoError(s.T(), err)
  assert.Equal(s.T(), expectedFile, file)
  s.mockClient.AssertExpectations(s.T())
  \`\`\`
- **Solution**: Remove this test or replace with integration test using real HTTP client with test server. Mock should only be used for testing components that depend on the Client interface.
- **Priority**: P2

#### Example Tests Using Mock (Acceptable)
- **File**: internal/api/example_test.go
- **Line**: 45-75
- **Principle**: Tests as documentation
- **Note**: These are example tests demonstrating mock usage, not production tests. They can remain as examples.

#### Logging Client Tests (Appropriate Mocking)
- **File**: internal/api/client_logging_test.go
- **Line**: 22-56
- **Principle**: Use mocks for external dependencies
- **Note**: These tests appropriately mock the underlying client to test the logging wrapper behavior.

---

### Architecture & Clean Code Violations

#### Interface Segregation Principle (ISP) Violation
- **File**: internal/api/client.go
- **Line**: 8-107
- **Principle**: ISP - Clients shouldn't depend on interfaces they don't use
- **Problem**: The Client interface has 38 methods, forcing all clients to implement all methods. This creates interface bloat and makes testing harder.
- **Example**: 
  \`\`\`go
  type Client interface {
      GetFile(ctx context.Context, fileKey string, opts ...GetFileOption) (*File, error)
      // ... 37 more methods
  }
  \`\`\`
- **Solution**: Split into smaller, focused interfaces (e.g., FileClient, CommentClient, StyleClient, ComponentClient, WebhookClient). Use interface composition where needed.
- **Priority**: P1

#### Complete Implementation Violation (Unimplemented Methods)
- **File**: internal/api/transport/http_client.go
- **Line**: 220, 226, 232, 238, 244, 250, 256, 343, 349, 355, 361, 367, 373, 379, 385, 391, 397, 432, 438, 444, 450, 456, 462, 468, 474, 480, 486
- **Principle**: Complete Implementation - Never commit "not implemented" stubs
- **Problem**: 25+ API methods return "not implemented" errors, making the CLI unusable for those features.
- **Example**: 
  \`\`\`go
  func (c *HTTPClient) GetImage(ctx context.Context, fileKey string, nodeIDs []string, options *api.ImageOptions) (map[string]string, error) {
      // TODO: implement image endpoint
      return nil, fmt.Errorf("not implemented")
  }
  \`\`\`
- **Solution**: Implement missing methods or remove them from the interface. If feature is not ready, create issues and implement incrementally.
- **Priority**: P0

#### Error Handling Gaps
- **File**: cmd/figma/cmd/context.go, get.go, versions.go
- **Line**: 198, 95, 130
- **Principle**: Always check errors
- **Problem**: Flag retrieval errors are ignored using blank identifier. If flag parsing fails, the error is silently discarded.
- **Example**: 
  \`\`\`go
  outputFormat, _ = cmd.Flags().GetString("output-format")
  \`\`\`
- **Solution**: Handle errors properly: `outputFormat, err := cmd.Flags().GetString(...); if err != nil { return err }`
- **Priority**: P1

#### TODO Comments in Production Code
- **File**: Multiple files
- **Line**: internal/api/transport/http_client.go (many), internal/auth/manager.go:80, internal/config/save.go:51, etc.
- **Principle**: No TODO comments in production code
- **Problem**: 36 TODO comments scattered across production code, indicating incomplete work.
- **Example**: 
  \`\`\`go
  // TODO: implement OAuth token refresh using Figma API
  \`\`\`
- **Solution**: Convert each TODO to a beads issue and remove the comment. Implement immediately if <2 hours effort.
- **Priority**: P2

#### Code Duplication (Generated Code)
- **File**: internal/api/retry_client_gen.go, logging_client_gen.go
- **Line**: Generated files
- **Principle**: DRY (Don't Repeat Yourself)
- **Note**: These files are generated from templates. While there is duplication, it's acceptable as generated code. However, consider if generation could be replaced with composition or reflection to reduce boilerplate.
- **Solution**: Evaluate if code generation is necessary vs using decorator pattern with reflection.
- **Priority**: P3

---

## Recommendations
1. **P0**: Implement all missing API methods in HTTP client. Start with core methods (GetImage, GetComments, etc.) based on user needs.
2. **P1**: Split Client interface into smaller interfaces grouped by domain (File, Comment, Style, Component, Webhook, Variable, DevResource).
3. **P1**: Fix error handling for flag retrieval across all command files.
4. **P2**: Remove over-mocking test (client_test.go) or convert to integration test.
5. **P2**: Convert all TODO comments to beads issues and remove from code.
6. **P3**: Evaluate code generation strategy - consider using runtime decoration instead of compile-time generation.

## Best Practices Found
- Good use of dependency injection (Transport, RequestBuilder)
- Clean architecture layers (config, logging, API client)
- Well-structured tests for core functionality (text search, node traversal)
- Appropriate use of mocks for testing logging decorator
- Consistent naming conventions and formatting
- Comprehensive configuration management with validation
- Good separation of concerns in command structure

