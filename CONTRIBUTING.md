# Contributing to OrbitDeploy

Thank you for your interest in contributing to OrbitDeploy! We welcome contributions from the community.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Coding Guidelines](#coding-guidelines)
- [Submitting Changes](#submitting-changes)
- [Reporting Bugs](#reporting-bugs)
- [Suggesting Enhancements](#suggesting-enhancements)

## Code of Conduct

This project and everyone participating in it is governed by our [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- Use a clear and descriptive title
- Describe the exact steps to reproduce the problem
- Provide specific examples to demonstrate the steps
- Describe the behavior you observed and what you expected to see
- Include screenshots if applicable
- Include your environment details (OS, Go version, Podman version, etc.)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, include:

- A clear and descriptive title
- A detailed description of the proposed enhancement
- Explain why this enhancement would be useful
- List any similar features in other projects if applicable

### Pull Requests

1. Fork the repository and create your branch from `main`
2. Make your changes following our coding guidelines
3. Add tests for your changes if applicable
4. Ensure the test suite passes
5. Update documentation as needed
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.24 or later
- Node.js 18+ and npm
- Podman
- Git

### Backend Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/OrbitDeploy.git
cd OrbitDeploy

# Install Go dependencies
go mod download

# Run tests
go test ./...
```

### Frontend Setup

```bash
cd frontend

# Install dependencies (using bun)
bun install

# Start development server
bun run dev

# Build for production
bun run build
```

### Running the Application

```bash
# From the root directory
go run main.go
```

The backend will start on port `:8285` by default.

## Coding Guidelines

### Go Code

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Write meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and concise
- Write unit tests for new functionality

### TypeScript/JavaScript Code

- Follow the existing code style
- Use TypeScript for type safety
- Use Prettier for code formatting
- Use ESLint for code linting
- Write meaningful variable and function names
- Add JSDoc comments for complex functions

### Project Structure

- Backend code follows a modular structure with clear separation between HTTP layer, business logic (services), and data models
- Database models and API models are separated
- Frontend uses the API hooks from `frontend/src/api/apiHooksW.ts` (`useApiQuery`, `useApiMutation`)
- API endpoints are imported from `frontend/src/api/endpoints`

### Commit Messages

- Use clear and meaningful commit messages
- Start with a verb in the present tense (e.g., "Add", "Fix", "Update")
- Keep the first line under 72 characters
- Add a detailed description if necessary

Example:
```
Add user authentication middleware

- Implement JWT token validation
- Add authentication required decorator
- Update documentation
```

## Submitting Changes

1. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make Your Changes**
   - Write clear, documented code
   - Add tests if applicable
   - Follow the coding guidelines

3. **Commit Your Changes**
   ```bash
   git add .
   git commit -m "Add your meaningful commit message"
   ```

4. **Push to Your Fork**
   ```bash
   git push origin feature/your-feature-name
   ```

5. **Create a Pull Request**
   - Go to the original repository on GitHub
   - Click "New Pull Request"
   - Select your feature branch
   - Fill in the PR template with details about your changes
   - Link any related issues

6. **Code Review**
   - Respond to feedback from maintainers
   - Make requested changes
   - Keep the discussion focused and professional

## Testing

Before submitting your PR, ensure:

- All existing tests pass: `go test ./...`
- New features include appropriate tests
- Code is properly formatted
- No linting errors

## Documentation

- Update relevant documentation for any changes
- Add docstrings/comments for new functions
- Update README.md if you change user-facing features
- Create or update docs in the `docs/` directory for significant features

## Questions?

Feel free to open an issue with the label "question" if you need help or clarification on anything.

## License

By contributing to OrbitDeploy, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to OrbitDeploy! 🚀
