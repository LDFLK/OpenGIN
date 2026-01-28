# Contributing Guidelines

Thank you for your interest in contributing to this project! We welcome contributions from everyone. This document provides guidelines and best practices for contributing.

## Code of Conduct
Please read our [Code of Conduct](CODE_OF_CONDUCT.md) before contributing.

By participating in this project, you agree to maintain a respectful and inclusive environment for everyone.

## How to Contribute

There are many ways to contribute to this project:

- **Report Bugs**: Submit bug reports with detailed information
- **Suggest Features**: Propose new features or improvements
- **Improve Documentation**: Fix typos, clarify explanations, add examples
- **Submit Code**: Fix bugs or implement new features
- **Review Pull Requests**: Help review and test contributions from others

## Getting Started

Please refer to the [README](README.md#getting-started) section for detailed instructions on how to get started.

## Making Changes

### Branching Strategy

Create a topic branch from the main branch for your changes:

```bash
git checkout -b feature/your-feature-name
```

### Commit Messages

Write clear and meaningful commit messages. We recommend following this format:

```
[TYPE] Short description (max 50 chars)

Longer description if needed. Explain the "why" behind the change,
not just the "what". Reference any related issues.

Fixes #123
```

**Types**: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

### Coding Standards

- Add unit tests directly inside the `tests/` directory and organize test files in a clear, maintainable structure.
- Ensure all test cases are correct, validated, and cover the intended functionality thoroughly.
- Run all tests locally to confirm they pass before committing, and update or add tests as needed for any new changes.

### Testing

All changes should include appropriate tests. Run the test suite before submitting:

```bash
pytest
```

## Submitting Changes

### Pull Request Process

1. Ensure your code follows the project's coding standards
2. Update documentation if needed
3. Add or update tests as appropriate
4. Run the full test suite and ensure it passes
5. Push your branch and create a Pull Request

### Pull Request Guidelines

- Provide a clear title and description
- Reference any related issues (e.g., "Fixes #123")
- Keep changes focused and atomic
- Be responsive to feedback and review comments

<!-- For projects requiring CLA -->
<!-- ### Contributor License Agreement (CLA)
For significant contributions, you may need to sign a Contributor License Agreement.
-->

### Review Process

- PRs require at least one approval from a maintainer before merging.
- All CI checks (tests, linting, build) must pass successfully.
- Reviewers may request changes; address feedback before merging.
- Keep commits clean and descriptive to maintain a clear project history.
- Merge using the approved method (e.g., merge commit, or rebase) as per project conventions.

## Communication

- GitHub Issues: For bug reports and feature requests
- [contact@datafoundation.lk](mailto:contact@datafoundation.lk)

## Recognition

- Contributors are recognized in the `CONTRIBUTORS.md` file and acknowledged in release notes.
- Major contributions, bug fixes, and feature additions are highlighted to give proper credit.
- Community involvement, such as reviewing PRs or reporting issues, is also appreciated and noted.

We value all contributions and appreciate your effort to improve this project!

## Additional Resources

- [See Issues](https://github.com/LDFLK/GI-SERVICE/issues)
- [See API Documentation](README.md#api-endpoints)
- [Getting Started](README.md#getting-started)

---

*These guidelines are inspired by the [Apache Way](https://www.apache.org/theapacheway/) and [Open Source Guides](https://opensource.guide/).*
