# Contributing to Jitter Distill

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing.

## Code of Conduct

Be respectful, inclusive, and constructive. We follow the [Contributor Covenant](https://www.contributor-covenant.org/).

## How to Contribute

### Reporting Bugs

- Search existing issues first
- Use the bug report template
- Include Go version, OS, and reproduction steps

### Suggesting Features

- Check if it has been requested
- Explain the use case and benefit
- Consider implementation complexity

### Pull Requests

1. Fork the repo
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Run tests: `make test`
5. Run linter: `make lint`
6. Commit with clear message
7. Push to your fork
8. Open a Pull Request

## Development Setup

```bash
# Clone the repository
git clone https://github.com/nimish-nirmal/jitter-distill.git
cd jitter-distill

# Install tools
make install-tools

# Run tests
make test

# Run all checks
make all
```

## Code Style

- Follow Go conventions (gofmt)
- Write tests for new features
- Update documentation
- Keep commits focused and atomic

## Review Process

- All PRs require review
- CI must pass
- Maintainers will review within 1 week
