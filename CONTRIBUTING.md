# Contributing to SearchJet OSS

Thanks for your interest in contributing!

## Development Setup

```bash
# Clone the repo
git clone https://github.com/maidulcu/SearchJet-OSS.git
cd SearchJet-OSS

# Run tests
make test

# Run locally with Docker
make dev
```

## Code Standards

- Use Go 1.25+
- Run `make lint` before committing
- Add tests for new functionality
- Use interfaces for swappable components
- Use `slog` for structured logging

## Pull Request Process

1. Fork and create a branch
2. Run tests and lint
3. Update docs if needed
4. Submit PR with description

## PDPL Compliance

All code that processes user data must:
- Log audit events via `slog`
- Respect consent headers
- Include request ID in logs