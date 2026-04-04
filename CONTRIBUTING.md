# Contributing Guidelines

## Branching Strategy

- `main` is protected. Direct commits are forbidden.
- All changes must be submitted via Pull Request (PR) targeting the `main` branch.
- Use GitFlow branch names (for a concise explanation, read the [Atlassian Gitflow Workflow](https://www.atlassian.com/git/tutorials/comparing-workflows/gitflow-workflow) guide):
    - `feature/<description>` for new features
    - `bugfix/<description>` for bug fixes
    - `docs/<description>` for documentation updates
    - `visualizer/<description>` for visualizer-related changes

## Workflow

1. Update local `main`:

    ```bash
    git checkout main
    git pull origin main
    ```

2. Create working branch:

    ```bash
    git checkout -b <type>/<description>
    ```

3. Commit logical changes:
   Use [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) format. Keep commits focused and atomic.

    ```bash
    git add <files>
    git commit -m "<type>: <description>"
    ```

    _Valid types: feat, fix, docs, refactor, test, visualizer_

4. Push working branch:
   If pushing this branch for the first time, set the upstream:

    ```bash
    git push -u origin <type>/<description>
    ```

    For subsequent pushes on the same branch:

    ```bash
    git push
    ```

5. Open Pull Request to `main`.
6. Await Code Review and approval before merging.

## Development Setup

### Prerequisites

- **Go 1.21+** for backend services
- **Node.js 16+** for visualizer frontend
- **Docker & Docker Compose** for infrastructure

### Environment Setup

```bash
# Clone the repository
git clone <repository-url>
cd dominium

# Install Go dependencies
go mod download

# Install frontend dependencies (for visualizer)
cd web
npm install
cd ..
```

### Running Tests

```bash
# Run all Go tests
go test ./...

# Run frontend tests
cd web && npm test
```

## Code Standards

### Go Code

- Follow standard Go formatting (`go fmt`)
- Use `gofmt -s` for additional simplifications
- Run `go vet` and `golint` before committing
- Use meaningful variable names and add comments for complex logic

### React/JavaScript Code

- Use ESLint configuration provided
- Follow React best practices
- Use TypeScript for type safety (when applicable)
- Keep components small and focused

### Commit Messages

Follow conventional commit format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Examples:

- `feat(visualizer): add real-time blockchain canvas`
- `fix(api): handle race condition in transaction validation`
- `docs(readme): update installation instructions`

## Component-Specific Guidelines

### Backend (Go)

- Place business logic in `internal/` packages
- Use `pkg/` for reusable libraries
- Keep `cmd/` packages minimal (only main functions)
- Add unit tests for all exported functions

### Frontend (React)

- Use functional components with hooks
- Implement proper error boundaries
- Add loading states for async operations
- Use Ant Design components consistently

### Visualizer Features

- Test WebSocket connections thoroughly
- Handle network disconnections gracefully
- Ensure real-time updates don't cause performance issues
- Add proper TypeScript types for API responses

## Testing

### Unit Tests

- Write tests for all new functions
- Aim for 80%+ code coverage
- Use table-driven tests for multiple scenarios
- Mock external dependencies (Kafka, WebSocket)

### Integration Tests

- Test API endpoints with real HTTP requests
- Test WebSocket communication
- Verify blockchain consensus under various conditions

### Manual Testing

- Test visualizer with multiple nodes running
- Simulate network partitions and reconnections
- Test attack simulations in controlled environment

## Documentation

### Code Documentation

- Add GoDoc comments for all exported functions/types
- Document complex algorithms and edge cases
- Update README files when adding new features

### API Documentation

- Keep `API_EXAMPLES.md` updated with new endpoints
- Document request/response formats clearly
- Include error codes and their meanings

## Security Considerations

- Never log private keys or sensitive data
- Validate all user inputs thoroughly
- Use secure random generation for cryptographic operations
- Follow OWASP guidelines for web components

## Performance Guidelines

- Profile code before optimizing
- Use efficient data structures
- Minimize allocations in hot paths
- Consider concurrency patterns for I/O operations

## Getting Help

- Check existing issues and documentation first
- Use descriptive issue titles
- Provide minimal reproduction cases
- Include relevant logs and system information
