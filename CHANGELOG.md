# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-01-01

### Added
- Initial release of Jitter Distill
- Production-grade entropy pool with concurrent harvester workers
- Online AR-p LMS adaptive filter for ML-based residual extraction
- HMAC-SHA256 cryptographic whitening (HKDF-Extract)
- 256-bit and 512-bit token generation
- Thread-safe API with `NewEntropyPool()` and `GenerateToken()`
- CLI tool with benchmark mode
- Comprehensive test suite with race detection
- Performance benchmarks
- gzip token compression support
- Auto-reseeding mechanism
- GitHub Actions CI/CD workflows
- Multi-platform release automation with Goreleaser
- Comprehensive documentation with Mermaid diagrams

### Security
- Zero heap allocations in hot path
- Numerical stability guards against NaN/Inf
- Weight clamping to prevent model divergence
- Thread-safe concurrent access
- Graceful shutdown with resource cleanup
