# Code Optimization and Test Coverage Issues

This document identifies opportunities for optimization and improved test coverage to enhance the maintainability and quality of the Servo codebase.

## Code Deduplication Opportunities

### High Priority

#### **Issue: Duplicated Secret Expansion Logic**
- **Affected Files**: 
  - `clients/vscode/client.go:318-352`
  - `clients/claude_code/client.go:275-309`
  - `clients/cursor/client.go:274-307`
- **Problem**: The `expandSecrets` function is implemented identically across all three client implementations, violating DRY principles
- **Impact**: Makes maintenance difficult - bug fixes or enhancements must be applied in three places, increasing risk of inconsistencies
- **Suggested Fix**: Extract shared secret expansion logic into `internal/client/base.go` as `ExpandSecretsInString(value string, secretsProvider func(string) (string, error)) string`
- **Priority**: High - Critical for maintainability

#### **Issue: Repeated Client Registration Pattern**
- **Affected Files**:
  - `internal/cli/commands/install.go:38-42`
  - `internal/cli/commands/status.go:24-29`
  - `internal/cli/app.go:32-35`
- **Problem**: Client registry setup with identical client registrations is repeated across multiple commands
- **Impact**: New clients must be registered in multiple locations; easy to miss one and create inconsistencies
- **Suggested Fix**: Create `internal/client/registry.go:GetDefaultRegistry()` function that returns a pre-configured registry with all built-in clients
- **Priority**: High - Reduces coupling and improves consistency

#### **Issue: Command Initialization Boilerplate**
- **Affected Files**:
  - `internal/cli/commands/install.go:31-51`
  - `internal/cli/commands/status.go:23-34`
  - `internal/cli/commands/secrets.go:14-23`
- **Problem**: Similar initialization patterns for project managers, session managers, and client registries across commands
- **Impact**: Verbose code that's prone to inconsistencies; makes adding new commands more error-prone
- **Suggested Fix**: Create `internal/cli/commands/base.go` with shared initialization helpers like `NewBaseCommand()` that provides common dependencies
- **Priority**: Medium - Improves code clarity and consistency

### Medium Priority

#### **Issue: Duplicated Config Path Resolution**
- **Affected Files**:
  - `clients/vscode/client.go:179-184`
  - `clients/cursor/client.go:165-170`
- **Problem**: Similar logic for resolving local config paths with custom path overrides
- **Impact**: Increases maintenance burden for path handling logic
- **Suggested Fix**: Move path resolution logic to `internal/client/base.go` as a shared helper function
- **Priority**: Medium - Moderate impact on maintainability

#### **Issue: Repeated Platform Support Checking**
- **Affected Files**:
  - `clients/vscode/client.go:55-58`
  - `clients/claude_code/client.go:85-88`
  - `clients/cursor/client.go:55-57`
- **Problem**: Same platform support validation pattern across all clients
- **Impact**: Inconsistent implementation details, potential for platform-specific bugs
- **Suggested Fix**: Already partially addressed via `internal/client/base.go:IsPlatformSupportedForPlatforms()` - ensure all clients use this consistently
- **Priority**: Medium - Security and compatibility concern

#### **Issue: Common File Operations Pattern**
- **Affected Files**:
  - `internal/cli/commands/secrets.go:262-297` (loadSecretsData)
  - `internal/project/manager.go` (various file operations)
  - Multiple client files (config file operations)
- **Problem**: Similar patterns for loading/saving YAML and JSON files with base64 encoding/decoding
- **Impact**: Inconsistent error handling and file operations across the codebase
- **Suggested Fix**: Create `internal/utils/fileops.go` with standardized file operation helpers for YAML, JSON, and base64 operations
- **Priority**: Medium - Improves error handling consistency

## Missing Test Coverage

### High Priority

#### **Issue: Missing Tests for Critical Commands**
- **Affected Files**: Commands without corresponding `*_test.go` files
  - `internal/cli/commands/secrets.go` - No test file exists
  - `internal/cli/commands/configure.go` - No test file exists
  - `internal/cli/commands/work.go` - No test file exists
  - `internal/cli/commands/validate.go` - No test file exists
- **Problem**: Core functionality lacks automated validation, increasing risk of regressions
- **Impact**: High risk of breaking changes going unnoticed; difficult to refactor with confidence
- **Suggested Fix**: Create comprehensive test suites for each command covering success paths, error conditions, and edge cases
- **Priority**: High - Critical for code quality and reliability

#### **Issue: Incomplete Client Test Coverage**
- **Affected Files**:
  - `clients/claude_code/client.go:232-273` (GenerateConfig method not tested)
  - `clients/vscode/client.go:269-316` (GenerateConfig method not tested)
  - `clients/cursor/client.go:223-271` (GenerateConfig method not tested)
- **Problem**: The most complex method in each client (config generation) lacks test coverage
- **Impact**: Configuration generation bugs could break MCP server setups for users
- **Suggested Fix**: Add comprehensive tests for `GenerateConfig` method including secret expansion, manifest parsing, and file generation
- **Priority**: High - Core functionality reliability

#### **Issue: Missing Edge Case Tests for Install Command**
- **Affected Files**: `internal/cli/commands/install_test.go` (limited scenarios)
- **Problem**: Install command tests only cover basic manifest storage; missing tests for:
  - Invalid source formats
  - Network failures during git operations
  - Permission errors during file operations
  - Malformed servo files
  - Session validation edge cases
- **Impact**: Install failures in production could leave projects in inconsistent states
- **Suggested Fix**: Expand test suite to cover error conditions, invalid inputs, and edge cases
- **Priority**: High - Install is a critical user-facing operation

### Medium Priority

#### **Issue: Missing Integration Tests for End-to-End Workflows**
- **Affected Files**: Limited end-to-end test coverage
- **Problem**: While unit tests exist, there are gaps in testing complete workflows like:
  - Init → Install → Configure → Work pipeline
  - Multi-session project management
  - Client configuration generation with real secrets
  - Docker compose generation with service dependencies
- **Impact**: Integration issues between components may not be caught until manual testing
- **Suggested Fix**: Expand integration test suite in `test/` directory to cover complete user workflows
- **Priority**: Medium - Important for user experience reliability

#### **Issue: Insufficient Error Scenario Testing**
- **Affected Files**: Multiple components lack error path testing
- **Problem**: Many functions return errors but lack tests that validate error conditions:
  - Network timeouts during git operations
  - File system permission errors
  - Invalid YAML/JSON parsing errors
  - Missing required secrets during config generation
- **Impact**: Poor error messages and handling in failure scenarios
- **Suggested Fix**: Add error scenario tests for each major operation that can fail
- **Priority**: Medium - Improves user experience and debugging

#### **Issue: Missing Session Manager Edge Case Tests**
- **Affected Files**: `internal/session/manager_test.go` (basic coverage only)
- **Problem**: Session manager tests don't cover:
  - Concurrent session operations
  - Invalid session names
  - Session directory corruption scenarios
  - Active session tracking edge cases
- **Impact**: Session management bugs could corrupt project state
- **Suggested Fix**: Expand session manager test suite to cover edge cases and error conditions
- **Priority**: Medium - Session integrity is important for project consistency

### Low Priority

#### **Issue: Missing Performance Tests**
- **Affected Files**: No performance or benchmark tests exist
- **Problem**: No automated validation of performance characteristics for operations like:
  - Large manifest parsing
  - Bulk secret operations
  - Configuration generation with many servers
- **Impact**: Performance regressions may go unnoticed
- **Suggested Fix**: Add benchmark tests for performance-critical operations
- **Priority**: Low - Performance is adequate for current use cases

## Other Maintainability Concerns

### Medium Priority

#### **Issue: Inconsistent Error Handling Patterns**
- **Affected Files**: Multiple files across the codebase
- **Problem**: Mix of error handling approaches:
  - Some functions return detailed wrapped errors (`fmt.Errorf("failed to X: %w", err)`)
  - Others return simple error messages
  - Inconsistent use of error context and user-friendly messages
- **Impact**: Debugging difficulty and inconsistent user experience
- **Suggested Fix**: Establish consistent error handling patterns and create error utilities in `internal/utils/errors.go`
- **Priority**: Medium - Affects debugging and user experience

#### **Issue: Magic String Usage**
- **Affected Files**: 
  - Multiple files use hardcoded strings like "local", "default", ".servo"
  - Client names as string literals across different files
- **Problem**: No centralized constants for frequently used strings
- **Impact**: Typos can cause subtle bugs; difficult to change standard values
- **Suggested Fix**: Create `pkg/constants.go` for commonly used strings and paths
- **Priority**: Medium - Reduces bug potential and improves maintainability

#### **Issue: Limited Input Validation**
- **Affected Files**: Command implementations and public API functions
- **Problem**: Limited validation of user inputs:
  - Session names (special characters, length limits)
  - Secret keys and values (format validation)
  - Client names (case sensitivity issues)
  - File paths (injection prevention)
- **Impact**: Potential for user confusion and edge case bugs
- **Suggested Fix**: Create validation utilities and apply consistently across commands
- **Priority**: Medium - Improves robustness and security

### Low Priority

#### **Issue: Documentation Inconsistencies**
- **Affected Files**: Various files with incomplete or inconsistent documentation
- **Problem**: Some interfaces and complex functions lack comprehensive documentation
- **Impact**: Reduces code comprehension for new contributors
- **Suggested Fix**: Audit and standardize Go documentation across all public interfaces
- **Priority**: Low - Code is generally well-structured and readable

#### **Issue: Configuration File Format Versioning**
- **Affected Files**: Configuration generation and parsing logic
- **Problem**: Limited versioning strategy for configuration file formats
- **Impact**: Future format changes may break backward compatibility
- **Suggested Fix**: Implement configuration schema versioning and migration utilities
- **Priority**: Low - Current format is stable and sufficient

## Summary

**High Priority Issues:** 7 issues requiring immediate attention
- 4 code deduplication opportunities
- 3 critical test coverage gaps

**Medium Priority Issues:** 10 issues for medium-term improvement
- 3 code quality improvements  
- 4 test coverage enhancements
- 3 maintainability concerns

**Low Priority Issues:** 3 issues for future consideration

**Estimated Impact:**
- Addressing high-priority deduplication could reduce maintenance effort by ~25%
- Adding missing test coverage would significantly improve confidence in refactoring
- The current codebase shows good architecture; these improvements would enhance its long-term maintainability without requiring major structural changes

**Recommended Approach:**
1. Start with secret expansion logic deduplication (highest impact, lowest risk)
2. Add test coverage for critical commands before making structural changes
3. Address client registration and command initialization patterns
4. Tackle remaining medium-priority issues incrementally