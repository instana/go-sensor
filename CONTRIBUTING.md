# Contributing to IBM Instana Go Tracer

Thank you for your interest in contributing to the IBM Instana Go Tracer! This project is part of the [IBM Instana Observability](https://www.ibm.com/products/instana) tool set, providing tracing, metrics, logs, and profiling capabilities for Go applications.

## Code of Conduct

We are committed to providing a welcoming and inclusive environment for all contributors. By participating in this project, you agree to abide by our code of conduct:

- **Be respectful**: Treat everyone with respect and consideration
- **Be collaborative**: Work together constructively and be open to feedback
- **Be inclusive**: Welcome newcomers and help them get started
- **Be professional**: Keep discussions focused and productive
- **Be patient**: Remember that people have different skill levels and backgrounds

We do not tolerate harassment, discrimination, or any form of inappropriate behavior. If you experience or witness unacceptable behavior, please report it to the project maintainers.

## How to Contribute

We welcome contributions of all kinds: bug fixes, new features, documentation improvements, and more. The contribution process is designed to maintain code quality while being accessible to contributors.

### Development Environment

Contributors should have a recent version of Go installed, along with standard development tools like Git and Make. The project uses linting tools to maintain code quality, which can be installed locally for validation before submitting changes.

### Code Standards

All contributions must follow established Go coding conventions and project-specific requirements. This includes proper code formatting, consistent style, and required copyright headers in source files. The project provides automated tooling to help verify compliance with these standards before submission.

### Testing

The project maintains comprehensive test coverage to ensure reliability and prevent regressions. Contributors should validate their changes against the existing test suite. The codebase is organized into multiple modules with independent testing capabilities, particularly for instrumentation packages, allowing focused testing of specific components.

### Pull Request Process

Contributions are submitted through GitHub pull requests against the `main` branch. Each PR should have a clear description of the changes and their purpose, reference any related issues, and ensure all tests pass. We encourage contributors to follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) for commit messages to maintain a clear and consistent project history. The Instana Go Tracer team reviews all submissions and may request changes or improvements. Once approved, maintainers will merge the contribution.

## Project Structure

The repository contains the core tracer implementation along with multiple instrumentation packages for popular Go frameworks and libraries. Each instrumentation package is independently versioned and maintained under the `instrumentation/` directory. Examples demonstrating usage patterns are available in the `example/` directory, and comprehensive documentation can be found in the `docs/` directory.

## Getting Help

For questions about contributing, check the [README.md](README.md) and [docs/](docs/) directory for documentation. The [example/](example/) directory contains practical usage examples. For bugs or feature requests, open an issue on GitHub. The maintainer team is available to help guide contributions and answer questions.

## License

By contributing to this project, you agree that your contributions will be licensed under the [MIT License](LICENSE.md).

---

Thank you for contributing to IBM Instana Go Tracer! 🎉