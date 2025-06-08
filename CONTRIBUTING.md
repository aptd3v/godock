# Contributing to godock

Thank you for your interest in contributing to godock! We aim to make Docker container management in Go more intuitive and type-safe. This document provides guidelines and instructions for contributing to the project.

## Table of Contents
- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Pull Request Process](#pull-request-process)
- [Coding Guidelines](#coding-guidelines)
- [Documentation](#documentation)
- [Community](#community)

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct (coming soon). We expect all contributors to be respectful, inclusive, and professional in all interactions.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/godock.git
   cd godock
   ```
3. Add the upstream repository:
   ```bash
   git remote add upstream https://github.com/aptd3v/godock.git
   ```
4. Create a branch for your work:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

1. Ensure you have Go 1.21 or later installed
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Install Docker (required for integration tests)
4. Run tests to verify your setup:
   ```bash
   go test ./...
   ```

## Making Changes

1. Make your changes in your feature branch
2. Keep your changes focused and atomic
3. Follow the existing code style and patterns
4. Add or update tests as needed
5. Update documentation for any changed functionality
6. Commit your changes with clear, descriptive commit messages

### Commit Message Guidelines

Format:
```
<type>: <subject>

[optional body]
[optional footer]
```

Types:
- feat: New feature
- fix: Bug fix
- docs: Documentation changes
- style: Code style changes (formatting, etc)
- refactor: Code refactoring
- test: Adding or updating tests
- chore: Maintenance tasks

Example:
```
feat: add container resource monitoring support

Add methods to monitor container CPU and memory usage.
Includes real-time stats collection and aggregation.

Closes #123
```

## Testing

We maintain a comprehensive test suite including both unit and integration tests.

1. Run unit tests:
   ```bash
   go test ./...
   ```

2. Run integration tests (requires Docker):
   ```bash
   go test ./... -tags=integration
   ```

3. Run tests with coverage:
   ```bash
   go test -coverprofile=coverage.out ./...
   go tool cover -html=coverage.out
   ```

### Writing Tests

- Write both unit and integration tests for new functionality
- Follow table-driven test patterns when appropriate
- Use meaningful test names and descriptions
- Include edge cases and error conditions
- Keep tests focused and maintainable

## Pull Request Process

1. Update your fork with the latest upstream changes:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. Ensure your PR:
   - Has a clear, descriptive title
   - Includes relevant tests
   - Updates documentation as needed
   - Follows coding guidelines
   - Has a clean commit history

3. Submit your PR and:
   - Link any related issues
   - Provide a clear description of changes
   - Highlight any breaking changes
   - Request review from maintainers

## Coding Guidelines

1. Code Style
   - Follow standard Go formatting (use `gofmt`)
   - Use meaningful variable and function names
   - Keep functions focused and reasonable in length
   - Add comments for exported functions and types

2. Design Principles
   - Maintain type safety
   - Keep the API intuitive and user-friendly
   - Follow Go idioms and best practices
   - Consider backward compatibility

3. Error Handling
   - Return meaningful errors
   - Use error wrapping appropriately
   - Include context in error messages
   - Handle all error cases

## Documentation

- Update README.md for significant changes
- Document all exported functions and types
- Include examples for new features
- Keep examples in the examples/ directory up to date
- Use godoc formatting for code documentation

## Community

- Open issues for bugs or feature requests
- Ask questions and help others
- Be respectful and constructive
- Follow project conventions and guidelines

## Getting Help

If you need help or have questions:
1. Check existing issues and documentation
2. Open a new issue for questions
3. Tag maintainers for urgent matters

Thank you for contributing to godock! Your efforts help make the project better for everyone. 