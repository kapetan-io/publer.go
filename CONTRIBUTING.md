# Contributing to publer.go

Thank you for your interest in contributing to publer.go! This document provides guidelines and information for contributors.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Project Structure](#project-structure)

## Code of Conduct

We are committed to providing a welcoming and inclusive environment for all contributors. Please be respectful and constructive in all interactions.

## Getting Started

### Prerequisites

- Go 1.24.4 or later
- Git
- Make (for running build tasks)

### Development Setup

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/publer.go.git
   cd publer.go
   ```

3. Add the upstream repository as a remote:
   ```bash
   git remote add upstream https://github.com/thrawn/publer.go.git
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

5. Run the test suite to ensure everything works:
   ```bash
   make test
   ```

## Making Changes

### Branch Strategy

- Create a new branch for each feature or bugfix
- Use descriptive branch names (e.g., `feature/add-media-upload`, `fix/rate-limit-handling`)
- Keep branches focused on a single change

```bash
git checkout -b feature/your-feature-name
```

### Development Process

1. **Understand the codebase**: Review the implementation plans in the [`plans/`](./plans/) directory
2. **Write tests first**: Follow TDD principles when possible
3. **Implement your changes**: Keep changes focused and well-documented
4. **Test thoroughly**: Ensure all tests pass and add new tests for your changes
5. **Update documentation**: Update README or code comments if needed

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make cover

# Run linting
make lint

# Run all CI checks
make ci
```

### Test Guidelines

Follow the project's testing patterns as defined in the codebase:

- Tests MUST be in the `package XXX_test` format (not `package XXX`)
- Test names should be camelCase starting with a capital letter
- Use table-driven tests for parameter validation
- Use `github.com/stretchr/testify/require` and `github.com/stretchr/testify/assert`
- Use `require` for critical assertions, `assert` for non-critical ones
- Avoid explanatory messages in assertions
- Be explicit rather than DRY in tests


### Mock Server

The project includes a mock server for testing. When adding new API endpoints:

1. Add mock responses to `v1/mock_server.go`
2. Ensure realistic response structures
3. Include error scenarios (rate limits, validation errors, etc.)
4. Test both success and failure cases

## Submitting Changes

### Pull Request Process

1. **Sync with upstream**:
   ```bash
   git fetch upstream
   git checkout main
   git merge upstream/main
   ```

2. **Rebase your branch**:
   ```bash
   git checkout your-feature-branch
   git rebase main
   ```

3. **Run tests**: Ensure all tests pass
   ```bash
   make ci
   ```

4. **Commit your changes**: Use clear, descriptive commit messages
   ```bash
   git add .
   git commit -m "Add support for media uploads in posts"
   ```

5. **Push to your fork**:
   ```bash
   git push origin your-feature-branch
   ```

6. **Create a Pull Request** on GitHub

### Pull Request Guidelines

- **Title**: Clear and descriptive summary of changes
- **Description**: Follow this format:
  ```markdown
  ### Purpose
  A paragraph describing what this change intends to achieve

  ### Implementation
  - An explanation of each major code change made
  ```
- **Testing**: Describe how you tested your changes
- **Breaking Changes**: Clearly document any breaking changes
- **Documentation**: Update relevant documentation

## Code Style

### Go Code Guidelines

Follow the project's established patterns:

- Use `const` for variables that don't change and are used more than once
- Prefer short, clear variable names (1-2 words)
- Avoid local variables used only once - inline values directly
- Use full words instead of abbreviations
- Use `const` for constants: `const numRetries = 3`
- Use `lo.ToPtr()` for creating pointers to local variables

## Questions?

If you have questions about contributing:

1. Check the [implementation plans](./plans/) for context
2. Look at existing code for patterns
3. Open an issue for discussion before making large changes
4. Ask questions in pull request comments

Thank you for contributing to publer.go! 🎉