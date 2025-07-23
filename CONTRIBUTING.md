# Contributing to Goatest

We love your input! We want to make contributing to Goatest as easy and transparent as possible, whether it's:

- Reporting a bug
- Discussing the current state of the code
- Submitting a fix
- Proposing new features
- Becoming a maintainer

## Development Process

We use GitHub to host code, to track issues and feature requests, as well as accept pull requests.

## Pull Requests

Pull requests are the best way to propose changes to the codebase. We actively welcome your pull requests:

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

## Commit Messages

Please write clear and descriptive commit messages that explain what changes were made and why. Good commit messages help maintain a clear project history.

## Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code
- Follow the linting rules defined in `.golangci.yml`
- Write clear, self-documenting code with appropriate comments
- Keep functions focused and testable

## Testing

- Write tests for all new functionality
- Ensure existing tests continue to pass
- Use table-driven tests where appropriate
- Test both success and error cases
- Integration tests should use the `TestMain` pattern

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## References

This document was adapted from the open-source contribution guidelines for [Facebook's Draft](https://github.com/facebook/draft-js/blob/a9316a723f9e918afde44dea68b5f9f39b7d9b00/CONTRIBUTING.md)