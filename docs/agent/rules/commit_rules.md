# Commit Rules

To maintain a clean and searchable project history, all commits must follow these rules.

## Message Format

We use the **Conventional Commits** specification. The message should be structured as follows:

```text
<type>(<scope>): <description>

[optional body]

[optional footer(s)]
```

### Types

- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `chore`: Changes to the build process or auxiliary tools and libraries such as documentation generation

### Scope (Optional)

A scope may be provided to a commit's type, to provide additional contextual information and is contained within parenthesis, e.g., `feat(parser): add ability to parse arrays`.

## Rules for Pushing

1.  **Atomicity**: Each commit should represent a single, logical change.
2.  **Descriptive Body**: If the change is complex, the body should explain the "why" and "how" of the implementation.
3.  **Verification**: Never push code that breaks the build or fails existing tests.
4.  **Documentation**: If a feature or logic changes, ensure the corresponding documentation in `docs/` is updated in the same or a subsequent commit.
