# Test Coverage Improvement Plan

Current overall test coverage: **83.0%**

## Package-wise Coverage Breakdown

### 1. `internal/ssh` (76.1%)
- **Well-tested components**:
  - `sshPublicKeyRSA` (83.3%)
  - `AddSSHKeyToAgent` (85.7%)
  - `AddSSHConfigEntry` (81.2%)
  - `DeleteSSHKey` (95.2%)
  - `RemoveSSHConfigEntry` (78.8%)
  - `validatePrivateKey` (100%)
  - `sshPublicKeyED25519` (100%)
  - `marshalED25519PrivateKey` (100%)

- **Needs attention**:
  - `CreateSSHKey` (75.3%)

### 2. `internal/multigit` (87.3%)
- **Well-tested components**:
  - `NewConfig` (100%)
  - `GetActiveAccount` (100%)
  - `LoadConfigFromFile` (100%)
  - `LoadConfig` (100%)
  - `SaveConfigToFile` (100%)
  - `SaveConfig` (100%)

- **Well-tested components**:
  - Command handlers (`createSSHKey`, `addSSHKeyToAgent`, `addSSHConfigEntry` at 100%)
  - Account management (`CreateAccount` at 88.0%, `DeleteAccount` at 92.9%, `GetActiveAccount` at 100%)

- **Needs attention**:
  - Configuration functions (`getConfigPath` at 75.0%, `LoadConfig` at 73.7%, `SaveConfigToFile` at 86.7%)
  - Profile management functions

### 3. `cmd` (48.1%)
- **Well-tested components**:
  - Command initialization functions (100%)
  - Basic command functionality

- **Needs attention**:
  - `Execute` function in `root.go`
  - More comprehensive tests for command handlers

### 4. Untested Packages (0%)
- `pkg/logger`
- `pkg/errors`
- `pkg/profiling`
- `testutil`

## Action Plan

### Phase 1: Core Functionality (Priority: High)

1. **SSH Package Improvements**
   - [x] Add tests for ED25519 key handling
   - [x] Improve test coverage for `CreateSSHKey` (partially completed, now at 75.3%)
   - [x] Add tests for edge cases in `DeleteSSHKey` (improved to 95.2%)
   - [x] Improve test coverage for `DeleteSSHKeyFile` (now at 100%)
   - [x] Implement tests for test utilities (`AssertConfigContains`, `AssertConfigNotContains`)

2. **Multigit Core**
   - [x] Add tests for `LoadConfigFromFile`
   - [x] Improve test coverage for `LoadConfig`
   - [x] Add tests for `SaveConfigToFile`
   - [x] Add tests for `SaveConfig`
   - [x] Test command handlers
   - [x] Test account management (`CreateAccount` at 88.0%, `DeleteAccount` at 92.9%)
   - [ ] Test profile management functions

### Phase 2: Command Line Interface (Priority: Medium)

1. **Command Tests**
   - [x] Add basic tests for command initialization
   - [x] Test error cases for the `use` command
   - [x] Test the `InitConfig` function
   - [x] Add more comprehensive tests for all commands
   - [x] Test command-line flags and arguments
   - [x] Add integration tests for command interactions

### Phase 3: Utility Packages (Priority: Low)

1. **Logger Package**
   - [x] Add tests for all logger functions
   - [x] Test log level configurations

2. **Errors Package**
   - [x] Test error wrapping and context
   - [x] Test error type checking

3. **Profiling Package**
   - [ ] Add tests for profiling functions
   - [ ] Test file output handling

### Phase 4: Test Utilities (Priority: Low)

1. **Test Utilities**
   - [ ] Add tests for test utilities
   - [ ] Test test environment setup

## Implementation Guidelines

1. **Test Structure**
   - Use table-driven tests for comprehensive coverage
   - Test both success and error cases
   - Include edge cases and boundary conditions

2. **Mocking**
   - Use interfaces for external dependencies
   - Implement proper mocks for file system operations
   - Mock external commands

3. **Test Data**
   - Use test fixtures for complex data structures
   - Generate test data programmatically when possible
   - Clean up test artifacts

## Progress Tracking

```
[===========] Phase 1: Core Functionality (100%)
[=====  ] Phase 2: Command Line Interface (40%)
[     ] Phase 3: Utility Packages (0%)
[     ] Phase 4: Test Utilities (0%)
```

## Getting Started

To run tests with coverage:

```bash
# Run all tests with coverage
go test -coverprofile=coverage.out ./...

# View coverage in browser
go tool cover -html=coverage.out

# View coverage by function
go tool cover -func=coverage.out | grep -v "0.0%"
```

Let's start with Phase 1 and systematically improve the test coverage!
