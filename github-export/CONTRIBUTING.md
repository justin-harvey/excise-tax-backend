# Contributing to Excise Tax Portal

Thank you for your interest in contributing to the Excise Tax Portal! This document provides guidelines and instructions for contributing.

## 🤝 How to Contribute

### Reporting Bugs

If you find a bug, please create an issue with:
- Clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Environment details (OS, Go version, Docker version)
- Screenshots if applicable

### Suggesting Features

Feature suggestions are welcome! Please:
- Check existing issues first
- Describe the feature and its benefits
- Explain use cases
- Consider backward compatibility

### Pull Requests

1. **Fork the repository**
2. **Create a feature branch** from `develop`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
3. **Make your changes** following our coding standards
4. **Add tests** for new functionality
5. **Update documentation** if needed
6. **Commit with clear messages**:
   ```bash
   git commit -m "feat: add amazing new feature"
   ```
7. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```
8. **Create a Pull Request** to `develop` branch

## 📝 Coding Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting
- Run `golangci-lint` before committing
- Maintain test coverage above 80%
- Write clear, descriptive variable names
- Add comments for exported functions and types

### Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Build process or auxiliary tool changes

**Examples:**
```
feat(payment): add XRPL payment monitoring
fix(auth): resolve JWT token expiration issue
docs(api): update API endpoint documentation
```

### Code Review Process

1. All PRs require at least one approval
2. All tests must pass
3. Code coverage should not decrease
4. Linter must pass
5. Documentation must be updated

## 🧪 Testing

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test ./internal/payment/...

# Integration tests
make test-integration
```

### Writing Tests

- Write unit tests for all new functions
- Use table-driven tests where appropriate
- Mock external dependencies
- Test both success and failure cases
- Test edge cases and boundary conditions

**Example:**
```go
func TestPaymentService_CreatePayment(t *testing.T) {
    tests := []struct {
        name    string
        input   CreatePaymentRequest
        want    *Payment
        wantErr bool
    }{
        {
            name: "valid payment",
            input: CreatePaymentRequest{
                ManufacturerID: 1,
                AmountUSD:      100.00,
            },
            want: &Payment{ID: 1, Amount: 100.00},
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## 📚 Documentation

- Update README.md for user-facing changes
- Add godoc comments for all exported items
- Update API documentation for endpoint changes
- Include examples where helpful
- Keep documentation in sync with code

## 🔧 Development Setup

1. **Install Prerequisites**:
   - Go 1.21+
   - Docker Desktop 24.0+
   - golang-migrate
   - golangci-lint

2. **Clone and Setup**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/excise-tax-portal.git
   cd excise-tax-portal
   make setup
   ```

3. **Start Infrastructure**:
   ```bash
   cd infrastructure/docker
   docker compose up -d
   ```

4. **Run Migrations**:
   ```bash
   cd backend
   make migrate-up
   ```

5. **Start Development**:
   ```bash
   make run-dev
   ```

## 🏗️ Project Structure

```
excise-tax-portal/
├── backend/
│   ├── cmd/           # Service entry points
│   ├── internal/      # Private code
│   ├── pkg/           # Public libraries
│   ├── migrations/    # Database migrations
│   └── tests/         # Tests
├── infrastructure/    # Docker, K8s, Terraform
└── docs/             # Documentation
```

## 🚀 Release Process

1. All changes go through `develop` branch
2. Create release branch: `release/vX.Y.Z`
3. Update CHANGELOG.md
4. Update version numbers
5. Merge to `main` via PR
6. Tag release: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
7. GitHub Actions handles deployment

## 📋 Checklist

Before submitting a PR, ensure:

- [ ] Code follows project style guidelines
- [ ] All tests pass locally
- [ ] New tests added for new functionality
- [ ] Documentation updated
- [ ] Commit messages follow convention
- [ ] No merge conflicts with target branch
- [ ] PR description clearly explains changes
- [ ] Breaking changes are clearly marked

## ❓ Questions?

- Check [existing issues](https://github.com/YOUR_USERNAME/excise-tax-portal/issues)
- Start a [discussion](https://github.com/YOUR_USERNAME/excise-tax-portal/discussions)
- Contact maintainers

## 🙏 Thank You!

Your contributions make this project better. We appreciate your time and effort!

## Code of Conduct

Please note we have a code of conduct. Follow it in all your interactions with the project.

### Our Pledge

We pledge to make participation in our project a harassment-free experience for everyone.

### Our Standards

**Positive behavior:**
- Using welcoming and inclusive language
- Being respectful of differing viewpoints
- Gracefully accepting constructive criticism
- Focusing on what is best for the community

**Unacceptable behavior:**
- Trolling, insulting/derogatory comments, personal attacks
- Public or private harassment
- Publishing others' private information without permission
- Other conduct which could reasonably be considered inappropriate

### Enforcement

Violations may be reported to project maintainers. All complaints will be reviewed and investigated.
